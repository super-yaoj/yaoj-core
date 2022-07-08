package run

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/super-yaoj/yaoj-core/pkg/problem"
	"github.com/super-yaoj/yaoj-core/pkg/processor"
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

var pOutputCache = inMemoryCache[[]string]{
	data: map[sha][]string{},
}
var pResultCache = inMemoryCache[processor.Result]{
	data: map[sha]processor.Result{},
}

var CacheSize = 1000

var runMutex sync.Mutex

// Run all testcase in the dir. User option mode to choose from original tests,
// pretests and extra tests.
func RunProblem(r *problem.ProbData, dir string, subm problem.Submission, mode ...string) (*problem.Result, error) {
	runMutex.Lock()
	defer runMutex.Unlock()

	logger.Printf("run dir=%s", dir)

	if gcache == nil {
		return nil, fmt.Errorf("global cache not initialized")
	}

	gcache.Resize(CacheSize)

	// check submission
	for k := range r.Submission {
		if _, ok := (*subm[workflow.Gsubm])[k]; !ok {
			return nil, fmt.Errorf("submission missing field %s", k)
		}
	}

	// clear cache
	pOutputCache.Reset()
	pResultCache.Reset()

	// download submission
	inboundPath := subm.Download(dir)
	inboundPath[workflow.Gstatic] = toPathMap(r, r.Static)

	testdata := r.ProbTestdata
	if len(mode) > 0 {
		switch mode[0] {
		case "pretest":
			testdata = r.Pretest
		case "extra":
			testdata = r.Extra
		}
	}

	var result = problem.Result{
		IsSubtask:  r.IsSubtask(),
		CalcMethod: testdata.CalcMethod,
		Subtask:    []problem.SubtResult{},
	}

	// accumulate subtask score
	calcScore := func(sub_res *problem.SubtResult, score float64) bool {
		switch testdata.CalcMethod {
		case problem.Mmax:
			if sub_res.Score < score {
				sub_res.Score = score
			}
		case problem.Mmin:
			if sub_res.Score > score {
				sub_res.Score = score
			}
			if sub_res.Score == 0 { // 已经是 0 分了
				return true
				// skip = true // 后面的都没必要测了
			}
		default:
			sub_res.Score += score
		}
		return false
	}

	if testdata.IsSubtask() {
		records := testdata.Subtasks.Record
		dependon := func(i, j int) bool {
			deps := strings.Split(records[i]["_depend"], ",")
			for _, sid := range deps {
				if records[j]["_subtaskid"] == sid {
					// logger.Printf("%s need %s", records[i]["_subtaskid"], records[j]["_subtaskid"])
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

			if testdata.CalcMethod == problem.Mmin {
				sub_res.Score = sub_res.Fullscore
			}

			for tid, test := range tests {
				logger.Printf("test #%d", tid)
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
					res, err := runWorkflow(r.Workflow(), dir, inboundPath, test_score)
					if err != nil {
						return nil, err
					}
					data = *res
				}
				sub_res.Testcase = append(sub_res.Testcase, data)

				if calcScore(&sub_res, data.Score) {
					skip = true
				}
			}
			result.Subtask[id] = sub_res
		}
	} else {
		sub_res := problem.SubtResult{
			Testcase:  []workflow.Result{},
			Fullscore: r.Fullscore,
			Score:     0,
		}
		if testdata.CalcMethod == problem.Mmin {
			sub_res.Score = sub_res.Fullscore
		}
		var skip bool
		for _, test := range testdata.Tests.Record {
			inboundPath[workflow.Gtests] = toPathMap(r, test)

			score := r.Fullscore // Mmin or Mmax
			if testdata.CalcMethod == problem.Msum {
				score = r.Fullscore / float64(len(testdata.Tests.Record))
			}
			if f, err := strconv.ParseFloat(test["_score"], 64); err == nil {
				score = f
			}

			var data workflow.Result
			if skip {
				data.Title = "Skipped"
				data.Fullscore = score
				data.Score = 0
			} else {
				res, err := runWorkflow(r.Workflow(), dir, inboundPath, score)
				if err != nil {
					return nil, err
				}
				data = *res
			}
			sub_res.Testcase = append(sub_res.Testcase, data)

			if calcScore(&sub_res, data.Score) {
				skip = true
			}
		}
		sub_res.Fullscore = r.Fullscore
		result.Subtask = append(result.Subtask, sub_res)
	}
	return &result, nil
}

// mutex
func RunWorkflow(w workflow.Workflow, dir string, inboundPath map[workflow.Groupname]*map[string]string,
	fullscore float64) (*workflow.Result, error) {
	runMutex.Lock()
	defer runMutex.Unlock()

	logger.Printf("run workflow directly dir=%s", dir)

	if gcache == nil {
		return nil, fmt.Errorf("global cache not initialized")
	}

	gcache.Resize(CacheSize)

	// clear cache
	pOutputCache.Reset()
	pResultCache.Reset()

	return runWorkflow(w, dir, inboundPath, fullscore)
}

type hackAnalyzer struct {
	capture map[string]workflow.Outbound
	data    map[string][]byte
}

func (r *hackAnalyzer) Analyze(w workflow.Workflow, nodes map[string]workflow.RuntimeNode,
	fullscore float64) workflow.Result {

	r.data = map[string][]byte{}
	for field, bound := range r.capture {
		data, _ := os.ReadFile(nodes[bound.Name].Output[bound.LabelIndex])
		r.data[field] = data
	}

	return workflow.Result{}
}

// hackSubm 包含被 hack 的提交以及 hackinput
func RunHack(r *problem.ProbData, dir string, hackSubm, std problem.Submission) (*workflow.Result, error) {
	runMutex.Lock()
	defer runMutex.Unlock()

	logger.Printf("run hack dir=%s", dir)

	if gcache == nil {
		return nil, fmt.Errorf("global cache not initialized")
	}

	gcache.Resize(CacheSize)

	// clear cache
	pOutputCache.Reset()
	pResultCache.Reset()

	hackin := hackSubm.Download(dir)
	stdin := std.Download(dir)
	stdin[workflow.Gtests] = hackin[workflow.Gtests]
	stdin[workflow.Gstatic] = toPathMap(r, r.Static)

	halyz := hackAnalyzer{capture: r.HackIOMap}
	wk := workflow.Workflow{
		WorkflowGraph: r.Workflow().WorkflowGraph,
		Analyzer:      &halyz,
	}

	_, err := runWorkflow(wk, dir, stdin, r.Fullscore)
	if err != nil {
		return nil, err
	}

	for field, data := range halyz.data {
		file, err := os.CreateTemp(dir, "hackinput-*")
		if err != nil {
			return nil, err
		}
		file.Write(data)
		file.Close()

		if hackin[workflow.Gtests] == nil {
			hackin[workflow.Gtests] = &map[string]string{}
		}
		(*hackin[workflow.Gtests])[field] = file.Name()
		hackin[workflow.Gstatic] = toPathMap(r, r.Static)

		logger.Printf("tests add %q: %q", field, (*hackin[workflow.Gtests])[field])
	}

	return runWorkflow(r.Workflow(), dir, hackin, r.Fullscore)
}
