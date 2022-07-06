package run

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/super-yaoj/yaoj-core/pkg/problem"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
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
func testcaseOf(r *problem.ProbTestdata, subtaskid string) []map[string]string {
	res := []map[string]string{}
	for _, test := range r.Tests.Record {
		if test["_subtaskid"] == subtaskid {
			res = append(res, test)
		}
	}
	return res
}

// Run all testcase in the dir.
func RunProblem(r *problem.ProbData, dir string, subm problem.Submission, mode ...string) (*problem.Result, error) {
	logger.Printf("run dir=%s", dir)

	// check submission
	for k := range r.Submission {
		if _, ok := (*subm[workflow.Gsubm])[k]; !ok {
			return nil, fmt.Errorf("submission missing field %s", k)
		}
	}

	// clear cache
	gOutputCache.Reset()
	gResultCache.Reset()

	// download submission
	inboundPath := subm.Download(dir)
	inboundPath[workflow.Gstatic] = toPathMap(r, r.Static)

	var result = problem.Result{
		IsSubtask: r.IsSubtask(),
		Subtask:   []problem.SubtResult{},
	}

	testdata := r.ProbTestdata
	if len(mode) > 0 {
		switch mode[0] {
		case "pretest":
			testdata = r.Pretest
		case "extra":
			testdata = r.Extra
		}
	}

	if testdata.IsSubtask() {
		records := testdata.Subtasks.Record
		dependon := func(i, j int) bool {
			deps := strings.Split(records[i]["_depend"], ",")
			for _, sid := range deps {
				if records[j]["_subtaskid"] == sid {
					logger.Printf("%s need %s", records[i]["_subtaskid"], records[j]["_subtaskid"])
					return true
				}
			}
			return false
		}
		order, err := utils.TopoSort(len(records), dependon)
		if err != nil {
			return nil, err
		}

		result.Subtask = make([]problem.SubtResult, len(records))
		for _, id := range order {
			subtask := records[id]
			sub_res := problem.SubtResult{
				Subtaskid: subtask["_subtaskid"],
				Score:     0,
				Testcase:  []workflow.Result{},
			}

			var skip bool
			if testdata.CalcMethod == problem.Mmin {
				for j := range records {
					if j != id && dependon(id, j) && !result.Subtask[j].IsFull() {
						skip = true
					}
				}
			}

			inboundPath[workflow.Gsubt] = toPathMap(r, subtask)
			tests := testcaseOf(&testdata, subtask["_subtaskid"])

			// subtask score
			score, err := strconv.ParseFloat(subtask["_score"], 64)
			if err != nil {
				return nil, err
			}
			sub_res.Fullscore = score

			for _, test := range tests {
				inboundPath[workflow.Gtests] = toPathMap(r, test)

				// calc test fullscore
				var test_score = score // Mmin or Mmax
				if testdata.CalcMethod == problem.Msum {
					test_score = score / float64(len(tests))
				}

				var data workflow.Result
				if skip {
					data.Title = "Skipped"
					data.Fullscore = test_score
					data.Score = 0
				} else {
					res, err := RunWorkflow(r.Workflow(), dir, inboundPath, test_score)
					if err != nil {
						return nil, err
					}
					data = *res
				}
				sub_res.Testcase = append(sub_res.Testcase, data)

				// accumulate subtask score
				switch testdata.CalcMethod {
				case problem.Mmax:
					if sub_res.Score < data.Score {
						sub_res.Score = data.Score
					}
				case problem.Mmin:
					if sub_res.Score > data.Score {
						sub_res.Score = data.Score
					}
					if sub_res.Score == 0 { // 已经是 0 分了
						skip = true // 后面的都没必要测了
					}
				default:
					sub_res.Score += data.Score
				}
			}
			result.Subtask[id] = sub_res
		}
	} else {
		sub_res := problem.SubtResult{
			Testcase: []workflow.Result{},
		}
		for _, test := range testdata.Tests.Record {
			inboundPath[workflow.Gtests] = toPathMap(r, test)

			score := r.Fullscore // Mmin or Mmax
			if testdata.CalcMethod == problem.Mmin {
				score = r.Fullscore / float64(len(testdata.Tests.Record))
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

// Custom test 即提供测试数据和提交数据
// 大部分 custom test 不关注答案正确性，只关注是否 MLE 或者 TLE 之类。
// Subtask 数据取第一个subtask的数据
func RunCustom(r *problem.ProbData, dir string, subm problem.Submission) (*workflow.Result, error) {
	logger.Printf("run custom dir=%s", dir)

	inboundPath := subm.Download(dir)
	inboundPath[workflow.Gstatic] = toPathMap(r, r.Static)
	if !r.Tests.Fields().Check(*inboundPath[workflow.Gtests]) {
		return nil, fmt.Errorf("invalid test data")
	}

	if r.IsSubtask() {
		inboundPath[workflow.Gsubt] = toPathMap(r, r.Subtasks.Record[0])
	}

	return RunWorkflow(r.Workflow(), dir, inboundPath, r.Fullscore)
}
