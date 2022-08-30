package problem

import (
	"encoding/json"

	"github.com/super-yaoj/yaoj-core/pkg/workflow"
)

// 测试点得分的汇总方式
//
// 对于不开子任务的题目同样有效
//
// 通常的子任务模式：外 Msum 内 Mmin
//
// 传统模式：Msum 不开子任务
type CalcMethod int

const (
	// default
	Mmin CalcMethod = iota
	Mmax
	Msum
)

// Problem result
type Result struct {
	// 题目满分
	Fullscore float64 `json:"fullscore"`
	// 实际得分
	Score float64 `json:"score"`
	// Testcases 与 Subtasks 必有一个为 nil
	Testcases []workflow.Result `json:"testcases"`
	// Testcases 与 Subtasks 必有一个为 nil
	Subtasks []SubtResult `json:"subtasks"`
}

func (r Result) JSON() []byte {
	data, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return data
}

// Subtask result
type SubtResult struct {
	Subtaskid string            `json:"id"`
	Fullscore float64           `json:"fullscore"`
	Score     float64           `json:"score"`
	Testcases []workflow.Result `json:"testcases"`
}

func (r SubtResult) IsFull() bool {
	return r.Fullscore-r.Score < 1e-5
}
