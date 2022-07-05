package run

import (
	"fmt"
	"path"
	"strconv"

	"github.com/super-yaoj/yaoj-core/pkg/problem"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
)

// change a record's relative path to real path
func toPathMap(r *problem.ProbData, rcd map[string]string) *map[string]string {
	res := map[string]string{}
	for k, v := range rcd {
		res[k] = path.Join(r.Dir(), v)
	}
	return &res
}
func testcaseOf(r *problem.ProbData, subtaskid string) []map[string]string {
	res := []map[string]string{}
	for _, test := range r.Tests.Record {
		if test["_subtaskid"] == subtaskid {
			res = append(res, test)
		}
	}
	return res
}

func run(r *problem.ProbData, dir string, inboundPath map[workflow.Groupname]*map[string]string) (*problem.Result, error) {
	logger.Printf("run dir=%s", dir)
	var result = problem.Result{
		IsSubtask: r.IsSubtask(),
		Subtask:   []problem.SubtResult{},
	}
	if r.IsSubtask() {
		for _, subtask := range r.Subtasks.Record {
			sub_res := problem.SubtResult{
				Subtaskid: subtask["_subtaskid"],
				Testcase:  []workflow.Result{},
			}
			inboundPath[workflow.Gsubt] = toPathMap(r, subtask)
			tests := testcaseOf(r, subtask["_subtaskid"])

			// subtask score
			score, err := strconv.ParseFloat(subtask["_score"], 64)
			if err != nil {
				return nil, err
			}
			sub_res.Fullscore = score

			for _, test := range tests {
				inboundPath[workflow.Gtests] = toPathMap(r, test)

				// test score
				var test_score = score // Mmin or Mmax
				if r.CalcMethod == problem.Msum {
					test_score = score / float64(len(tests))
				}

				res, err := RunWorkflow(r.Workflow(), dir, inboundPath, test_score)
				if err != nil {
					return nil, err
				}
				sub_res.Testcase = append(sub_res.Testcase, *res)
			}
			result.Subtask = append(result.Subtask, sub_res)
		}
	} else {
		sub_res := problem.SubtResult{
			Testcase: []workflow.Result{},
		}
		for _, test := range r.Tests.Record {
			inboundPath[workflow.Gtests] = toPathMap(r, test)

			score := r.Fullscore // Mmin or Mmax
			if r.CalcMethod == problem.Mmin {
				score = r.Fullscore / float64(len(r.Tests.Record))
			}
			if f, err := strconv.ParseFloat(test["_score"], 64); err == nil {
				score = f
			}

			res, err := RunWorkflow(r.Workflow(), dir, inboundPath, score)
			if err != nil {
				return nil, err
			}
			sub_res.Testcase = append(sub_res.Testcase, *res)
		}
		sub_res.Fullscore = r.Fullscore
		result.Subtask = append(result.Subtask, sub_res)
	}
	return &result, nil
}

// Run all testcase in the dir.
func RunProblem(r *problem.ProbData, dir string, submission map[string]string) (*problem.Result, error) {
	logger.Printf("run dir=%s", dir)
	// check submission
	for k := range r.Submission {
		if _, ok := submission[k]; !ok {
			return nil, fmt.Errorf("submission missing field %s", k)
		}
	}

	var inboundPath = map[workflow.Groupname]*map[string]string{
		workflow.Gsubm: (*map[string]string)(&submission),
	}
	inboundPath[workflow.Gstatic] = toPathMap(r, r.Static)
	return run(r, dir, inboundPath)
}

// Custom test 即提供测试数据和提交数据
// 大部分 custom test 不关注答案正确性，只关注是否 MLE 或者 TLE 之类。
// Subtask 数据取第一个subtask的数据
func RunCustom(r *problem.ProbData, dir string, submission map[string]string, test map[string]string) (*workflow.Result, error) {
	var inboundPath = map[workflow.Groupname]*map[string]string{
		workflow.Gsubm:   &submission,
		workflow.Gtests:  &test,
		workflow.Gstatic: toPathMap(r, r.Static),
	}
	if !r.Tests.Fields().Check(test) {
		return nil, fmt.Errorf("invalid test data")
	}
	if r.IsSubtask() {
		inboundPath[workflow.Gsubt] = toPathMap(r, r.Subtasks.Record[0])
	}

	return RunWorkflow(r.Workflow(), dir, inboundPath, r.Fullscore)
}
