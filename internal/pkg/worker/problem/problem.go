package problemruntime

import (
	"fmt"
	"os"
	"path"

	"github.com/super-yaoj/yaoj-core/internal/pkg/analyzers"
	workflowruntime "github.com/super-yaoj/yaoj-core/internal/pkg/worker/workflow"
	"github.com/super-yaoj/yaoj-core/pkg/log"
	"github.com/super-yaoj/yaoj-core/pkg/problem"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
)

type RtProblem struct {
	*problem.Data

	// working dir
	dir string
	// logger
	lg *log.Entry
	// test
	tot_dir int
	// global cache
	cache workflowruntime.RtNodeCache
}

// 创建一个新的临时文件夹用于数据组的评测
//
// 同时会创建一个名为 subm 的子文件夹用于加载提交记录
//
// 同时会创建一个名为 work 的子文件夹用于测试点评测
func (r *RtProblem) TestsetDir() (string, error) {
	r.tot_dir++
	dir := path.Join(r.dir, "testset"+fmt.Sprintf("%03d", r.tot_dir))
	err := os.MkdirAll(dir, 0750)
	if err != nil {
		return "", err
	}
	err = os.MkdirAll(path.Join(dir, "subm"), 0750)
	if err != nil {
		return "", err
	}
	err = os.MkdirAll(path.Join(dir, "work"), 0750)
	if err != nil {
		return "", err
	}
	return dir, nil
}

func (r *RtProblem) RunTestset(set *problem.TestdataGroup, subm problem.Submission) (*problem.Result, error) {
	// check test set
	if set == nil {
		return nil, ErrInvalidSet
	}
	if r.Extra != set && r.Pretest != set && r.Data.Data != set {
		return nil, ErrInvalidSet
	}
	testdir, err := r.TestsetDir()
	workdir := path.Join(testdir, "work")
	if err != nil {
		return nil, err
	}
	inbounds := subm.Download(path.Join(testdir, "subm"))
	inbounds[workflow.Gstatic] = r.Static.InboundGroup()

	result := &problem.Result{}

	grader := NewGrader(set.Method, set.Fullscore, len(set.Testcases))
	if set.Testcases != nil {
		res, err := r.RunTestcases(set.Testcases, inbounds, workdir, grader)
		if err != nil {
			return nil, err
		}
		result.Testcases = res
	} else {
		for id, subtask := range set.Subtasks {
			sub_grader := NewGrader(subtask.Method, subtask.Fullscore, len(subtask.Testcases))
			sub_res, err := r.RunTestcases(subtask.Testcases, inbounds, workdir, sub_grader)
			if err != nil {
				return nil, err
			}
			result.Subtasks = append(result.Subtasks, problem.SubtResult{
				Subtaskid: fmt.Sprint(id),
				Fullscore: subtask.Fullscore,
				Score:     sub_grader.Sum(),
				Testcases: sub_res,
			})
			grader.Add(sub_grader.Sum())
		}
	}
	result.Fullscore = set.Fullscore
	result.Score = grader.Sum()
	return result, nil
}

func (r *RtProblem) RunTestcases(testcases []*problem.TestcaseData,
	inbounds workflow.InboundGroups, workdir string, grader *Grader) ([]workflow.Result, error) {
	// testcase fullscore
	results := make([]workflow.Result, 0)
	fullscore := grader.TaskFullscore()
	for _, testcase := range testcases {
		if grader.Skipable() {
			results = append(results, workflow.Result{
				ResultMeta: workflow.ResultMeta{
					Title:     "skipped",
					Score:     0,
					Fullscore: fullscore,
				},
			})
		} else {
			inbounds[workflow.Gtests] = testcase.InboundGroup()
			analyzer := analyzers.Get(r.AnalyzerName)
			if analyzer == nil {
				return nil, &DataError{r.AnalyzerName, ErrUnknownAnalyzer}
			}
			wk, err := workflowruntime.New(r.Workflow, workdir, fullscore, analyzer, r.lg)
			if err != nil {
				return nil, err
			}
			wk.UseCache(r.cache)
			test_res, err := wk.Run(inbounds, false)
			if err != nil {
				return nil, err
			}
			results = append(results, *test_res)
			grader.Add(test_res.Score)
		}
	}
	return results, nil
}

// 删除所有文件（销毁自身）
func (r *RtProblem) Finalize() error {
	err := os.RemoveAll(r.dir)
	if err != nil {
		r.lg.WithError(err).Warn("finalizing runtime problem")
	}
	return err
}

// create dir if necessary
func New(data *problem.Data, dir string, logger *log.Entry) (*RtProblem, error) {
	err := os.MkdirAll(dir, 0750)
	if err != nil {
		return nil, err
	}
	gcache, err := workflowruntime.NewCache(path.Join(dir, "cache"))
	if err != nil {
		return nil, err
	}
	return &RtProblem{
		Data:  data,
		dir:   dir,
		lg:    logger.WithField("problem", dir),
		cache: gcache,
	}, nil
}
