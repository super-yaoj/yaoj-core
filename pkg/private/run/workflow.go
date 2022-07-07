package run

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"github.com/k0kubun/pp/v3"
	"github.com/super-yaoj/yaoj-core/pkg/buflog"
	"github.com/super-yaoj/yaoj-core/pkg/private/processors"
	"github.com/super-yaoj/yaoj-core/pkg/processor"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
	wk "github.com/super-yaoj/yaoj-core/pkg/workflow"
)

// perform a workflow in a directory.
// inboundPath: map[datagroup_name]*map[field]filename
// do not remove cache when running workflow!
func RunWorkflow(w wk.Workflow, dir string, inboundPath map[wk.Groupname]*map[string]string,
	fullscore float64) (*wk.Result, error) {
	nodes := runtimeNodes(w.Node)

	for i, group := range w.Inbound {
		if group == nil {
			return nil, fmt.Errorf("w.Inbound[%s] == nil", i)
		}
		data := inboundPath[i]
		if data == nil {
			return nil, fmt.Errorf("inboundPath[%s] == nil", i)
		}
		for j, bounds := range *group {
			if _, ok := (*data)[j]; !ok {
				return nil, fmt.Errorf("invalid inboundPath: missing field %s %s", i, j)
			}
			for _, bound := range bounds {
				nodes[bound.Name].Input[bound.LabelIndex] = (*inboundPath[i])[j]
			}
		}
	}

	previousWd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	defer func() {
		// logger.Printf("Chdir back to %q", previousWd)
		os.Chdir(previousWd)
	}()

	err = os.Chdir(dir)
	if err != nil {
		return nil, err
	}
	// logger.Printf("Chdir %q", dir)

	err = topologicalEnum(w, func(id string) error {
		node := nodes[id]
		if !node.inputFullfilled() {
			return fmt.Errorf("input not fullfilled")
		}
		node.calcHash()
		if pOutputCache.Has(node.hash) { // cache level 1
			logger.Printf("Run node[%s] (cached lv 1)", id)
			cache_outputs := pOutputCache.Get(node.hash)[:]
			result := pResultCache.Get(node.hash)
			node.Output = cache_outputs
			node.Result = &result
		} else if gcache.Has(node.hash, "@result") { // cache level 2
			logger.Printf("Run node[%s] (cached lv 2)", id)
			result := processor.Result{}
			err := result.Unserialize(gcache.Get(node.hash, "@result"))
			if err != nil {
				return err
			}
			node.Output = make([]string, 0)
			for _, label := range processor.OutputLabel(node.ProcName) {
				filename := gcache.GetSource(node.hash, label)
				node.Output = append(node.Output, filename)
			}
		} else { // no cache
			logger.Printf("Run node[%s] no cache", id)
			for i := 0; i < len(node.Output); i++ {
				node.Output[i] = utils.RandomString(10)
			}
			result := node.Processor().Run(node.Input, node.Output)

			node.Result = result
			pOutputCache.Set(node.hash, node.Output)
			pResultCache.Set(node.hash, *result)

			if node.Cache {
				gcache.Set(node.hash, "@result", result.Serialize())
				for i, label := range processor.OutputLabel(node.ProcName) {
					gcache.SetSource(node.hash, label, node.Output[i])
				}
			}
		}
		nodes[id] = node
		for _, edge := range w.EdgeFrom(id) {
			nodes[edge.To.Name].Input[edge.To.LabelIndex] = nodes[edge.From.Name].Output[edge.From.LabelIndex]
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	runtimeNodes := map[string]wk.RuntimeNode{}
	for name, node := range nodes {
		runtimeNodes[name] = node.RuntimeNode
	}
	res := w.Analyze(w, runtimeNodes, fullscore)

	return &res, nil
}

type sha [32]byte

func (r sha) String() string {
	s := ""
	for _, v := range r {
		s += fmt.Sprintf("%02x", v)
	}
	return s
}

// SHA256 hash for file content.
// for any error, return empty hash
func fileHash(name string) sha {
	hash := sha256.New()
	f, err := os.Open(name)
	if err != nil {
		return sha{}
	}
	defer f.Close()

	if _, err := io.Copy(hash, f); err != nil {
		return sha{}
	}
	var b = hash.Sum(nil)
	if len(b) != 32 {
		pp.Print(b)
		panic(b)
	}
	return *(*sha)(b)
}

type rtNode struct {
	wk.RuntimeNode
	hash sha
}

// Get the processor of the node.
func (r rtNode) Processor() processor.Processor {
	return processors.Get(r.ProcName)
}

// sum up hash of all input files
func (r *rtNode) calcHash() {
	hash := sha256.New()
	for _, path := range r.Input {
		hashval := fileHash(path)
		hash.Write(hashval[:])
	}
	hash.Write([]byte(r.ProcName))
	var b = hash.Sum(nil)
	// pp.Print(b)
	if len(b) != 32 {
		pp.Print(b)
		panic(b)
	}
	r.hash = *(*sha)(b)
}

func (r *rtNode) inputFullfilled() bool {
	for _, path := range r.Input {
		if path == "" {
			return false
		}
	}
	return true
}

func runtimeNodes(node map[string]wk.Node) (res map[string]rtNode) {
	res = map[string]rtNode{}
	for k, v := range node {
		res[k] = rtNode{
			RuntimeNode: wk.RuntimeNode{
				Node:   v,
				Input:  make([]string, len(processor.InputLabel(v.ProcName))),
				Output: make([]string, len(processor.OutputLabel(v.ProcName))),
				Attr:   map[string]string{},
			},
		}
	}
	return
}

func topologicalEnum(w wk.Workflow, handler func(id string) error) error {
	indegree := map[string]int{}
	for _, edge := range w.Edge {
		indegree[edge.To.Name]++
	}
	for {
		flag := false
		p := ""
		for id := range w.Node {
			if indegree[id] == 0 {
				flag = true
				p = id
				break
			}
		}
		if !flag {
			break
		}
		// logger.Printf("topo current id=%s", p)
		indegree[p] = -1
		for _, edge := range w.EdgeFrom(p) {
			indegree[edge.To.Name]--
		}
		err := handler(p)
		if err != nil {
			return err
		}
	}
	for id := range w.Node {
		if indegree[id] != -1 {
			return fmt.Errorf("invalid DAG! id=%s", id)
		}
	}
	return nil
}

var logger = buflog.New("[run] ")
