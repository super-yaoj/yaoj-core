package workflowruntime

import (
	"errors"
	"os"

	"github.com/super-yaoj/yaoj-core/internal/pkg/processors"
	"github.com/super-yaoj/yaoj-core/pkg/data"
	"github.com/super-yaoj/yaoj-core/pkg/log"
	"github.com/super-yaoj/yaoj-core/pkg/processor"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
)

type (
	Workflow   = workflow.Workflow
	Result     = workflow.Result
	ResultFile = workflow.ResultFile
	ResultMeta = workflow.ResultMeta
)

// runtime node use hash to identify cache
type RtNode struct {
	workflow.Node
	// paths of input files
	Input processor.Inbounds
	// paths of output files
	Output processor.Outbounds
	// whether its output is determined by problem-wide things only
	Attr map[string]string
	// result of processor
	Result *processor.Result

	// hash is calculated during workflow testing
	hash *SHA

	// logger
	lg *log.Entry
}

// sum up hash of all input files and the node its self
//
// should be invoked after all inputs getting ready
//
// 算哈希的时候不考虑 nil input 的情况
func (r *RtNode) Hash() SHA {
	if r.hash == nil {
		hash := newShaHash()
		for _, name := range processor.InputLabel(r.ProcName) {
			store := r.Input[name]
			if store == nil {
				r.lg.WithField("input", name).Warn("nil input")
			} else {
				data, err := store.Get()
				if err == nil {
					hash.Write(data)
				} else {
					r.lg.WithError(err).Warn("error getting store")
				}
			}
		}
		hash.WriteString(r.ProcName)
		value := hash.SHA()
		r.hash = &value
	}
	r.lg.Debugf("hash: %s", r.hash.String())
	return *r.hash
}

func (r *RtNode) run(name string, cachers []RtNodeCache) error {
	cached := false
	for _, cacher := range cachers {
		if cacher.Exist(r) {
			if err := cacher.Assign(r); err != nil {
				return &Error{"cacher.Assign", err}
			}
			cached = true
			break
		}
	}
	if !cached {
		// check input complete
		for _, label := range processor.InputLabel(r.ProcName) {
			if r.Input[label] == nil {
				return &DataError{label, ErrIncompleteInput}
			}
		}
		r.lg.Info("run node without cache")
		// init output stores
		for _, label := range processor.OutputLabel(r.ProcName) {
			r.Output[label] = data.NewFile(utils.RandomString(10), nil)
		}
		r.Result = processors.Get(r.ProcName).Process(r.Input, r.Output)
	}
	if !cached {
		for _, cacher := range cachers {
			if err := cacher.Add(r); err != nil {
				return &Error{"cacher.Add", err}
			}
		}
	}
	return nil
}

type RtWorkflow struct {
	*workflow.Workflow
	RtNodes   map[string]*RtNode
	Fullscore float64
	// runtime working dir
	dir string
	// 外部提供的缓存，下标越小优先级越高
	caches []RtNodeCache
	// node names sorted topologically
	sortedNames []string

	analyzer Analyzer

	// logger
	lg *log.Entry
}

// create a new runtime workflow working in dir
//
// create dir if necessary
func New(wk *Workflow, dir string, fullscore float64, analyzer Analyzer, logger *log.Entry) (*RtWorkflow, error) {
	err := os.MkdirAll(dir, 0750)
	if err != nil {
		return nil, err
	}
	logger = logger.WithField("workflow", dir)

	res := &RtWorkflow{
		Workflow:  wk,
		RtNodes:   map[string]*RtNode{},
		Fullscore: fullscore,
		dir:       dir,
		lg:        logger,
		analyzer:  analyzer,
	}
	for name, node := range wk.Node {
		res.sortedNames = append(res.sortedNames, name)
		res.RtNodes[name] = &RtNode{
			Node:   node,
			Input:  make(processor.Inbounds),
			Output: make(processor.Outbounds),
			Attr:   map[string]string{},
			lg:     logger.WithField("node", name),
		}
	}
	sorted, err := utils.TopSort(res.sortedNames, func(u, v string) bool {
		for _, e := range wk.EdgeFrom(u) {
			if e.To.Name == v {
				return true
			}
		}
		return false
	})
	if err != nil {
		return nil, &Error{"topsort", err}
	}
	res.sortedNames = sorted
	return res, nil
}

// append cachers
func (r *RtWorkflow) UseCache(cachers ...RtNodeCache) {
	r.caches = append(r.caches, cachers...)
}

// make sure to pass logger by context
//
// dismiss_incomplete: 如果是在 hack 评测时跑 std，那么我们允许不完整的读入
// 在此模式下如果一个 processor 的读入不完整，那么它就不会被执行（即 result 是 nil）
func (r *RtWorkflow) Run(inbounds workflow.InboundGroups, dismiss_incomplete bool) (*workflow.Result, error) {
	// bind inbound to workflow
	for gname, group := range r.Inbound {
		if group == nil {
			panic("invalid workflow inbound")
		}
		if data := inbounds[gname]; data != nil {
			for j, bounds := range group {
				if store, ok := data[j]; ok {
					for _, bound := range bounds {
						r.RtNodes[bound.Name].Input[bound.Label] = store
					}
				} else {
					r.lg.WithFields(log.Fields{
						"group": gname,
						"field": j,
					}).Warn("inboundPath missing field")
				}
			}
		} else {
			r.lg.WithField("group", gname).Warn("inboundPath nil group")
		}
	}

	previousWd, err := os.Getwd()
	if err != nil {
		return nil, &Error{"getwd", err}
	}
	// go back after testing
	defer os.Chdir(previousWd)

	// change working dir
	err = os.Chdir(r.dir)
	if err != nil {
		return nil, &Error{"chdir", err}
	}
	r.lg.Debug("change working dir")

	for _, name := range r.sortedNames {
		err := r.RtNodes[name].run(name, r.caches)
		if errors.Is(err, ErrIncompleteInput) && dismiss_incomplete {
			r.lg.WithField("node", name).Debug("dismiss incomplete input")
			continue
		} else if err != nil {
			return nil, &DataError{"node: " + name, err}
		}
		for _, edge := range r.EdgeFrom(name) {
			r.RtNodes[edge.To.Name].Input[edge.To.Label] = r.RtNodes[edge.From.Name].Output[edge.From.Label]
		}
	}

	res := r.analyzer.Analyze(r)
	return &res, nil
}

// 删除所有文件（销毁自身）
func (r *RtWorkflow) Finalize() error {
	err := os.RemoveAll(r.dir)
	if err != nil {
		r.lg.WithError(err).Warn("finalizing runtime workflow")
	}
	return err
}

// 总结信息统计结果
// func (r *RtWorkflow) Sum(alyz Analyzer) Result {
// 	res := alyz.Analyze(r)
// 	sort.Slice(res.File, func(i, j int) bool {
// 		return res.File[i].Title < res.File[j].Title
// 	})
// 	return res
// }
