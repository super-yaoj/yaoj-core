package problem

import (
	"fmt"
	"path"

	"github.com/super-yaoj/yaoj-core/pkg/utils"
)

// create dir if necessary
//
// dir: 关于题目路径的相对文件夹
func (r *Data) newDirRecord(dir string) *DirRecord {
	return &DirRecord{
		prob: r,
		Dir:  dir,
		Lang: map[string]utils.LangTag{},
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
func (r *TestdataGroup) NewSubtask(fullscore float64, method CalcMethod) *SubtaskData {
	if r.Subtasks == nil {
		panic("subtasks not init")
	}
	tdir := path.Join(r.Dir, fmt.Sprint(len(r.Subtasks)))
	res := &SubtaskData{
		prob:      r.prob,
		Dir:       tdir,
		Fullscore: fullscore,
		Method:    method,
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
