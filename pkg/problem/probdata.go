package problem

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/super-yaoj/yaoj-core/pkg/utils"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
)

// 子任务中测试点得分的汇总方式
type CalcMethod int

const (
	// default
	Mmin CalcMethod = iota
	Mmax
	Msum
)

// Problem result
type Result struct {
	IsSubtask  bool
	CalcMethod CalcMethod
	Subtask    []SubtResult
}

func (r Result) Byte() []byte {
	data, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return data
}

var briefTpl = template.Must(template.New("brief").Parse(`
subtask: {{ .IsSubtask }}
{{if .IsSubtask}}{{range .Subtask}}{{ .Subtaskid }} ({{ .Fullscore }}pts)
{{range .Testcase}}{{ .Title }} {{ .Score }}pts {{ .Time }} {{ .Memory }}
{{end}}{{end}}
{{else}}{{range .Subtask}}{{range .Testcase}}{{ .Title }} {{ .Score }}pts {{ .Time }} {{ .Memory }}
{{end}}{{end}}
{{end}}
`))

func (r Result) Brief() string {
	var b bytes.Buffer
	if err := briefTpl.Execute(&b, r); err != nil {
		panic(err)
	}
	return b.String()
}

// Subtask result
type SubtResult struct {
	Subtaskid string
	Fullscore float64
	Score     float64
	Testcase  []workflow.Result
}

func (r SubtResult) IsFull() bool {
	return r.Fullscore-r.Score < 1e-5
}

// 题目评测时用到的数据
type ProbTestdata struct {
	// 子任务计分方式
	CalcMethod CalcMethod
	// "tests" _subtaskid, _score ("average", {number})
	Tests table
	// "subtask" _subtaskid, _score, _depend (separated by ",")
	Subtasks table
}

// Whether subtask is enabled.
func (r *ProbTestdata) IsSubtask() bool {
	return len(r.Subtasks.Field) > 0 && len(r.Subtasks.Record) > 0
}

func (r *ProbTestdata) Info() TestdataInfo {
	res := TestdataInfo{
		IsSubtask:  r.IsSubtask(),
		Subtasks:   make([]SubtaskInfo, 0),
		CalcMethod: r.CalcMethod,
	}
	if res.IsSubtask {
		for i, task := range r.Subtasks.Record {
			var tests = []TestInfo{}
			for j, test := range r.Tests.Record {
				if test["_subtaskid"] != task["_subtaskid"] {
					continue
				}
				tests = append(tests, TestInfo{
					Id:    j,
					Field: copyRecord(test),
				})
			}

			depend := []int{}
			if task["_depend"] != "" {
				deps := strings.Split(task["_depend"], ",")
				for _, dep := range deps {
					dep = strings.TrimSpace(dep)
					for id, subt := range r.Subtasks.Record {
						if subt["_subtaskid"] == dep {
							depend = append(depend, id)
						}
					}
				}
			}
			score, _ := strconv.ParseFloat(task["_score"], 64)
			res.Subtasks = append(res.Subtasks, SubtaskInfo{
				Id:        i,
				Fullscore: score,
				Field:     task,
				Tests:     tests,
				Depend:    depend,
			})
		}
	} else {
		var tests = []TestInfo{}
		for j, test := range r.Tests.Record {
			tests = append(tests, TestInfo{
				Id:    j,
				Field: copyRecord(test),
			})
		}

		res.Subtasks = append(res.Subtasks, SubtaskInfo{
			Fullscore: 0,
			Tests:     tests,
		})
	}
	return res
}

// Problem data module
type ProbData struct {
	// Usually 100.
	// Full score can be used to determine the point of testcase
	Fullscore float64
	workflow  workflow.Workflow
	// pretest 常用于样例评测
	Pretest ProbTestdata
	// 额外数据例如 hack 数据
	Extra ProbTestdata
	// 题目本身数据
	ProbTestdata
	// "submission" configuration
	Submission SubmConf
	// "static"
	Static record
	// "statement"
	// Statement has 1 record. "s.{lang}", "t.{lang}" represents statement and
	// tutorial respectively. "_tl" "_ml" "_ol" denotes cpu time limit (ms),
	// real memory limit (MB) and output limit (MB) respectively.
	// Others are just filename.
	Statement record
	// hack 时 tests 里需要提交的字段，以及其对应的限制
	// 为 nil 表示不支持 hack
	HackFields SubmConf
	// 由 tests 的字段映射到 workflow 的中间输出文件
	// 为 nil 表示不支持 hack
	HackIOMap map[string]workflow.Outbound
	// absolute dir
	dir string
}

// Add file to r.dir/patch and return relative path
func (r *ProbData) AddFile(name string, pathname string) (string, error) {
	name = path.Join("patch", name)
	logger.Printf("AddFile: %#v => %#v", pathname, name)
	if _, err := utils.CopyFile(pathname, path.Join(r.dir, name)); err != nil {
		return "", err
	}
	return name, nil
}

func (r *ProbData) AddFileReader(name string, file io.Reader) (string, error) {
	name = path.Join("patch", name)
	logger.Printf("AddFile: reader => %#v", name)
	destination, err := os.Create(path.Join(r.dir, name))
	if err != nil {
		return "", err
	}
	defer destination.Close()
	_, err = io.Copy(destination, file)
	if err != nil {
		return "", err
	}
	return name, nil
}

// export the problem's data to another empty dir and change itself to the new one
func (r *ProbData) Export(dir string) error {
	os.Mkdir(path.Join(dir, "workflow"), os.ModePerm)
	graph_json, err := json.Marshal(r.workflow.WorkflowGraph)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path.Join(dir, "workflow", "graph.json"), graph_json, 0644); err != nil {
		return err
	}
	os.Mkdir(path.Join(dir, "patch"), os.ModePerm)

	var statement, static record
	var testdata, pretest, extra ProbTestdata
	if testdata, err = r.exportTestdata(r.ProbTestdata, dir, path.Join("data")); err != nil {
		return err
	}
	if pretest, err = r.exportTestdata(r.Pretest, dir, path.Join("data", "pretest")); err != nil {
		return err
	}
	if extra, err = r.exportTestdata(r.Extra, dir, path.Join("data", "extra")); err != nil {
		return err
	}
	if static, err = r.exportRecord(0, r.Static, dir, path.Join("data", "static")); err != nil {
		return err
	}
	if statement, err = r.exportRecord(0, r.Statement, dir, path.Join("statement")); err != nil {
		return err
	}

	// modify r from now
	r.ProbTestdata = testdata
	r.Pretest = pretest
	r.Extra = extra
	r.Static = static
	r.Statement = statement

	prob_json, err := json.Marshal(*r)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(path.Join(dir, "problem.json"), prob_json, 0644); err != nil {
		panic(err)
	}
	r.dir = dir
	return nil
}

func copyTable(tb table) (res table) {
	if res_json, err := json.Marshal(tb); err != nil {
		panic(err)
	} else {
		if err := json.Unmarshal(res_json, &res); err != nil {
			panic(err)
		}
	}
	return
}

func copyRecord(rcd record) (res record) {
	res = record{}
	for field, val := range rcd {
		res[field] = val
	}
	return
}

func (r *ProbData) exportRecord(id int, rcd record, newroot, dircd string) (res record, err error) {
	logger.Printf("Export Record #%d %#v", id, dircd)
	os.MkdirAll(path.Join(newroot, dircd), os.ModePerm)
	res = make(record)
	for field, val := range rcd {
		if field[0] == '_' { // private field
			res[field] = rcd[field]
		} else {
			name := fmt.Sprintf("%s%d%s", field, id, path.Ext(val))
			if _, err := utils.CopyFile(path.Join(r.dir, val), path.Join(newroot, dircd, name)); err != nil {
				return res, err
			}
			res[field] = path.Join(dircd, name)
		}
	}
	return res, nil
}

func (r *ProbData) exportTable(tb table, newroot, dirtb string) (table, error) {
	logger.Printf("Export Table %#v", dirtb)
	os.MkdirAll(path.Join(newroot, dirtb), os.ModePerm)
	res := copyTable(tb)

	for i, record := range tb.Record {
		for field := range record {
			if _, ok := tb.Field[field]; !ok {
				return tb, fmt.Errorf("invalid field %s in record #%d", field, i)
			}
		}
		rcd, err := r.exportRecord(i, record, newroot, dirtb)
		if err != nil {
			return tb, err
		}
		res.Record[i] = rcd
	}
	return res, nil
}

func (r *ProbData) exportTestdata(data ProbTestdata, newroot, dir string) (ProbTestdata, error) {
	logger.Printf("Export Testdata %#v", dir)
	os.MkdirAll(path.Join(newroot, dir), os.ModePerm)
	res := newTestdata()
	res.CalcMethod = data.CalcMethod
	tests, err := r.exportTable(data.Tests, newroot, path.Join(dir, "tests"))
	if err != nil {
		return data, err
	}
	res.Tests = tests
	subtasks, err := r.exportTable(data.Subtasks, newroot, path.Join(dir, "subtasks"))
	if err != nil {
		return data, err
	}
	res.Subtasks = subtasks
	return res, nil
}

// Set workflow graph
func (r *ProbData) SetWkflGraph(serial []byte) error {
	graph, err := workflow.Load(serial)
	if err != nil {
		return err
	}
	r.workflow.WorkflowGraph = graph
	return nil
}

// load problem from a dir
func LoadProbData(dir string) (*ProbData, error) {
	serial, err := os.ReadFile(path.Join(dir, "problem.json"))
	if err != nil {
		return nil, err
	}
	var prob ProbData
	if err := json.Unmarshal(serial, &prob); err != nil {
		return nil, err
	}
	// initialize
	absdir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	prob.dir = absdir
	wkgh, err := workflow.LoadFile(path.Join(dir, "workflow", "graph.json"))
	if err != nil {
		return nil, err
	}
	prob.workflow = workflow.Workflow{
		WorkflowGraph: wkgh,
		Analyzer:      workflow.DefaultAnalyzer{},
	}
	return &prob, nil
}

// create a new problem in an empty dir
func NewProbData(dir string) (*ProbData, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	graph := workflow.NewGraph()
	var prob = ProbData{
		dir: dir,
		workflow: workflow.Workflow{
			WorkflowGraph: &graph,
			Analyzer:      workflow.DefaultAnalyzer{},
		},
		ProbTestdata: newTestdata(),
		Extra:        newTestdata(),
		Pretest:      newTestdata(),
		Static:       make(record),
		Submission:   map[string]SubmLimit{},
		Statement:    make(record),
	}
	if err := prob.Export(dir); err != nil {
		return nil, err
	}
	return &prob, nil
}

// get the workflow
func (r *ProbData) Workflow() workflow.Workflow {
	return r.workflow
}

// get problem dir
func (r *ProbData) Dir() string {
	return r.dir
}

// Set statement content to file in r.dir
func (r *ProbData) SetStmt(lang string, file string) {
	r.Statement["s."+GuessLang(lang)] = file
}

func (r *ProbData) SetValFile(rcd record, field string, filename string) error {
	pin, err := r.AddFile(utils.RandomString(5)+"_"+path.Base(filename), filename)
	if err != nil {
		return err
	}
	rcd[field] = pin
	return nil
}

// remove problem (dir)
func (r *ProbData) Finalize() error {
	logger.Printf("finalize %q", r.dir)
	return os.RemoveAll(r.dir)
}

func (r *ProbData) Hackable() bool {
	return r.HackFields != nil && r.HackIOMap != nil
}

func newTestdata() ProbTestdata {
	return ProbTestdata{
		Tests:    newTable(),
		Subtasks: newTable(),
	}
}
