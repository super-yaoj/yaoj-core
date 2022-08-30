package worker

import (
	"io"
	"os"
	"path"
	"sync"
	"time"

	"github.com/super-yaoj/yaoj-core/internal/pkg/analyzers"
	"github.com/super-yaoj/yaoj-core/internal/pkg/processors"
	problemruntime "github.com/super-yaoj/yaoj-core/internal/pkg/worker/problem"
	workflowruntime "github.com/super-yaoj/yaoj-core/internal/pkg/worker/workflow"
	"github.com/super-yaoj/yaoj-core/pkg/data"
	"github.com/super-yaoj/yaoj-core/pkg/log"
	"github.com/super-yaoj/yaoj-core/pkg/problem"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
	"github.com/super-yaoj/yaoj-core/pkg/workflow/preset"
	"github.com/super-yaoj/yaoj-core/pkg/yerrors"
)

// 提供题目评测的服务
//
// 原则上全局只有一个 Service 实例
type Service struct {
	*sync.Mutex

	// 总的工作目录
	dir string
	// 题目数据的存放目录
	data_dir string

	// 题目的 checksum 与其对应的文件路径
	//
	// 目前的 store 过于简陋，没有考虑到题目长时间不评测的空间回收问题
	store sync.Map
	// 评测的目录
	work_dir string

	lg *log.Entry
}

// 存入题目的数据
func (r *Service) SetProblem(checksum string, reader io.Reader) error {
	r.Lock()
	defer r.Unlock()

	file, err := os.CreateTemp(r.work_dir, "p-*.zip")
	if err != nil {
		return yerrors.Situated("create temp", err)
	}
	defer func() {
		file.Close()
		os.Remove(file.Name())
	}()

	_, err = io.Copy(file, reader)
	if err != nil {
		return yerrors.Situated("copy", err)
	}

	chk := utils.FileChecksum(file.Name()).String()
	if checksum != chk {
		return yerrors.Annotated("chk", chk, ErrInvalidChecksum)
	}

	prob_dir, err := os.MkdirTemp(r.data_dir, "p-")
	if err != nil {
		return yerrors.Situated("mkdir temp", err)
	}

	prob, err := problem.LoadFileTo(file.Name(), prob_dir)
	if err != nil {
		return yerrors.Situated("load problem file", err)
	}

	r.store.Store(checksum, prob)
	r.lg.Infof("SetProblem checksum=%s prob=%s", checksum, prob_dir)
	return nil
}

// checksum 为题目数据的校验值
//
// submission_data 为提交的数据
//
// mode: 目前可选 "pretest", "extra"
func (r *Service) RunProblem(checksum string, submission_data []byte, mode string) (*problem.Result, error) {
	r.Lock()
	defer r.Unlock()

	val, ok := r.store.Load(checksum)
	if !ok {
		return nil, yerrors.Annotated("checksum", checksum, ErrNoSuchProblem)
	}
	prob := val.(*problem.Data)

	rtprob, err := problemruntime.New(prob, path.Join(r.work_dir, utils.RandomString(8)), r.lg)
	if err != nil {
		return nil, yerrors.Situated("create RtProblem", err)
	}
	defer rtprob.Finalize()
	// determine testset
	testset := prob.Data
	if mode == "pretest" {
		testset = prob.Pretest
	}
	if mode == "extra" {
		testset = prob.Extra
	}

	// load submission
	submission, err := problem.LoadSubmData(submission_data)
	if err != nil {
		return nil, yerrors.Situated("load submission", err)
	}

	start_time := time.Now()
	defer r.lg.Infof("total judging time: %v", time.Since(start_time))

	result, err := rtprob.RunTestset(testset, submission)
	if err != nil {
		return nil, yerrors.Situated("run testset", err)
	}
	return result, nil
}

func (r *Service) CustomTest(submission_data []byte) (*workflow.Result, error) {
	r.Lock()
	defer r.Unlock()

	submission, err := problem.LoadSubmData(submission_data)
	if err != nil {
		return nil, yerrors.Situated("load submission", err)
	}
	dir := path.Join(r.work_dir, utils.RandomString(8))
	rtwork, err := workflowruntime.New(&preset.Customtest, dir, 100, analyzers.Customtest{}, r.lg)
	if err != nil {
		return nil, yerrors.Situated("create RtWorkflow", err)
	}
	defer rtwork.Finalize()

	inbounds := submission.Download(dir)
	inbounds[workflow.Gstatic] = map[string]data.FileStore{
		"runner_config": data.NewFile(path.Join(dir, "_limit"), (&processors.RunConf{
			RealTime: 10000,
			CpuTime:  10000,
			VirMem:   504857600,
			RealMem:  504857600,
			StkMem:   504857600,
			Output:   50485760,
			Fileno:   10,
		}).Serialize()),
	}
	cache, err := workflowruntime.NewCache(path.Join(dir, "cache"))
	if err != nil {
		return nil, err
	}
	rtwork.UseCache(cache)
	result, err := rtwork.Run(inbounds, false)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// create a new worker in a dir
//
// create the dir if necessary
func New(dir string, logger *log.Entry) (*Service, error) {
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return nil, err
	}

	data_dir := path.Join(dir, "data")
	err = os.MkdirAll(data_dir, 0777)
	if err != nil {
		return nil, err
	}

	work_dir := path.Join(dir, "work")
	err = os.MkdirAll(work_dir, 0777)
	if err != nil {
		return nil, err
	}

	return &Service{
		Mutex:    &sync.Mutex{},
		dir:      dir,
		data_dir: data_dir,
		work_dir: work_dir,
		store:    sync.Map{},
		lg:       logger.WithField("worker", dir),
	}, nil
}
