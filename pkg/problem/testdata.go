package problem

import (
	"fmt"
	"os"
	"path"

	"github.com/super-yaoj/yaoj-core/pkg/data"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
)

// 依赖文件夹的以文件相对路径的形式存储字段值
//
// field 字段对应的文件为 probDir/Dir/field
type DirRecord struct {
	prob *Data
	// 存放数据的关于 ProbDir 的相对路径形式文件夹
	Dir string `json:"dir"`
}

func (r *DirRecord) makeDir() error {
	err := os.MkdirAll(path.Join(r.prob.dir, r.Dir), 0750)
	if err != nil {
		return err
	}
	return nil
}
func (r *DirRecord) Delete(field string) error {
	return os.RemoveAll(path.Join(r.prob.dir, r.Dir, field))
}
func (r *DirRecord) SetData(field string, data []byte) error {
	if err := r.makeDir(); err != nil {
		return err
	}
	return os.WriteFile(path.Join(r.prob.dir, r.Dir, field), data, 0644)
}
func (r *DirRecord) GetData(field string) ([]byte, error) {
	data, err := os.ReadFile(path.Join(r.prob.dir, r.Dir, field))
	return data, err
}
func (r *DirRecord) SetSource(field string, source string) error {
	if err := r.makeDir(); err != nil {
		return err
	}
	_, err := utils.CopyFile(source, path.Join(r.prob.dir, r.Dir, field))
	return err
}

// 遍历文件夹中的文件（name 是完整的文件名）
func (r *DirRecord) Range(visitor func(field string, name string)) {
	entries, err := os.ReadDir(path.Join(r.prob.dir, r.Dir))
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		visitor(name, path.Join(r.prob.dir, r.Dir, name))
	}
}

// 转化为读入数据
func (r *DirRecord) InboundGroup() workflow.InboundGroup {
	res := workflow.InboundGroup{}
	r.Range(func(field, name string) {
		res[field] = data.NewFlexFile(name)
	})
	return res
}

// create dir if necessary
//
// dir: 关于题目路径的相对文件夹
func (r *Data) newDirRecord(dir string) *DirRecord {
	return &DirRecord{
		prob: r,
		Dir:  dir,
	}
}

// 单个测试点数据
type TestcaseData = DirRecord

// 子任务数据
type SubtaskData struct {
	prob *Data
	// 存放数据的文件夹
	Dir string `json:"dir"`
	// 测试点得分汇总方式
	Method CalcMethod `json:"method"`
	// 该子任务的分数
	Fullscore float64 `json:"fullscore"`
	// 测试点数据
	Testcases []*TestcaseData `json:"data"`
}

func (r *SubtaskData) NewTestcase() *TestcaseData {
	tdir := path.Join(r.Dir, fmt.Sprint(len(r.Testcases)))
	res := (*TestcaseData)(r.prob.newDirRecord(tdir))
	r.Testcases = append(r.Testcases, res)
	return res
}

// 一系列的评测数据，例如样例数据组、终测数据组、Hack 数据组
//
// 添加数据前必须显式地进行初始化
type TestdataGroup struct {
	prob *Data
	// 存放数据的文件夹
	Dir string `json:"dir"`
	// 该数据组的分数，一般等同于题目满分
	Fullscore float64 `json:"fullscore"`
	// 子测试数据的计分方式
	Method CalcMethod `json:"method"`
	// "tests" _subtaskid, _score ("average", {number})
	// Testcases 和 Subtasks 必有一方为 nil
	Testcases []*TestcaseData `json:"testcases"`
	// Testcases 和 Subtasks 必有一方为 nil
	Subtasks []*SubtaskData `json:"subtasks"`
}

// Whether subtask is enabled.
// 如果没有数据那么默认没有子任务
func (r *TestdataGroup) IsSubtask() bool {
	return r.Subtasks != nil
}

func (r *TestdataGroup) initProb(prob *Data) {
	r.prob = prob
	for _, td := range r.Testcases {
		td.prob = prob
	}
	for _, sd := range r.Subtasks {
		sd.prob = prob
		for _, td := range sd.Testcases {
			td.prob = prob
		}
	}
}

func (r *TestdataGroup) InitTestcases() {
	r.Subtasks = nil
	r.Testcases = []*TestcaseData{}
}
func (r *TestdataGroup) InitSubtasks() {
	r.Subtasks = []*SubtaskData{}
	r.Testcases = nil
}
func (r *TestdataGroup) NewTestcase() *TestcaseData {
	if r.Testcases == nil {
		panic("testcase not init")
	}
	tdir := path.Join(r.Dir, fmt.Sprint(len(r.Testcases)))
	res := (*TestcaseData)(r.prob.newDirRecord(tdir))
	r.Testcases = append(r.Testcases, res)
	return res
}
func (r *TestdataGroup) NewSubtask() *SubtaskData {
	if r.Subtasks == nil {
		panic("subtasks not init")
	}
	tdir := path.Join(r.Dir, fmt.Sprint(len(r.Subtasks)))
	res := &SubtaskData{
		prob: r.prob,
		Dir:  tdir,
	}
	r.Subtasks = append(r.Subtasks, res)
	return res
}

// dir: 相对于题目根目录的路径
func (r *Data) newTestdataGroup(dir string) *TestdataGroup {
	return &TestdataGroup{
		prob:      r,
		Dir:       dir,
		Fullscore: r.Fullscore,
	}
}

/*
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
*/
