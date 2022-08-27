package judgeserver

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/super-yaoj/yaoj-core/internal/pkg/analyzers"
	"github.com/super-yaoj/yaoj-core/internal/pkg/processors"
	problemruntime "github.com/super-yaoj/yaoj-core/internal/pkg/worker/problem"
	workflowruntime "github.com/super-yaoj/yaoj-core/internal/pkg/worker/workflow"
	"github.com/super-yaoj/yaoj-core/pkg/data"
	"github.com/super-yaoj/yaoj-core/pkg/problem"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
	"github.com/super-yaoj/yaoj-core/pkg/workflow/preset"
)

func Judge(ctx *Context) error {
	type Judge struct {
		Callback string `form:"cb" binding:"required"`
		Checksum string `form:"sum" binding:"required"`
		// options: "pretest" "extra" （这三个互斥）
		// options: "nocache"
		// "hack": 返回 workflow.Result
		// 多个 mode 可重复指定: &mode=hack&mode=nocache
		Modes []string `form:"mode"`
	}
	var qry Judge
	if err := ctx.BindQuery(&qry); err != nil {
		return &HttpError{http.StatusBadRequest, &Error{"bind query", err}}
	}

	if !ctx.store.Has(qry.Checksum) {
		return &HttpError{http.StatusBadRequest, ErrInvalidChecksum}
	}

	// load submission
	data, _ := io.ReadAll(ctx.Request.Body)
	submission, err := problem.LoadSubmData(data)
	if err != nil {
		return err
	}

	// load problem
	prob := ctx.store.Get(qry.Checksum)
	if err != nil {
		return err
	}

	// ready to judge
	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})

	go func() {
		start_time := time.Now()

		tmpdir, _ := os.MkdirTemp("", "yaoj-runtime-*")
		defer os.RemoveAll(tmpdir)

		rtprob, err := problemruntime.New(prob, tmpdir, ctx.lg)
		if err != nil {
			ctx.lg.Errorf("run problem: %v", err)
			return
		}
		// determine testset
		testset := prob.Data
		if utils.FindIndex(qry.Modes, "pretest") != -1 {
			testset = prob.Pretest
		}
		if utils.FindIndex(qry.Modes, "extra") != -1 {
			testset = prob.Extra
		}

		result, err := rtprob.RunTestset(testset, submission)
		if err != nil {
			ctx.lg.Errorf("run problem: %v", err)
			return
		}
		_, err = http.Post(qry.Callback, "text/json; charset=utf-8", bytes.NewReader(result.Byte()))
		if err != nil {
			ctx.lg.Errorf("callback request: %v", err)
		}

		ctx.lg.Infof("Total judging time: %v", time.Since(start_time))
	}()

	return nil
}

func CustomTest(ctx *Context) error {
	type CustomTest struct {
		Callback string `form:"cb" binding:"required"`
	}
	var qry CustomTest
	if err := ctx.BindQuery(&qry); err != nil {
		return err
	}

	// load submission
	dat, _ := io.ReadAll(ctx.Request.Body)
	submission, err := problem.LoadSubmData(dat)
	if err != nil {
		return err
	}

	// ready to judge
	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})

	go func() {
		tmpdir, _ := os.MkdirTemp(os.TempDir(), "custom-*")
		defer os.RemoveAll(tmpdir)

		inbounds := submission.Download(tmpdir)
		inbounds[workflow.Gstatic] = map[string]data.FileStore{
			"runner_config": data.NewFile(path.Join(tmpdir, "_limit"), (&processors.RunConf{
				RealTime: 10000,
				CpuTime:  10000,
				VirMem:   504857600,
				RealMem:  504857600,
				StkMem:   504857600,
				Output:   50485760,
				Fileno:   10,
			}).Serialize()),
		}
		rtwork, err := workflowruntime.New(&preset.Customtest, tmpdir, 100, analyzers.Customtest{}, ctx.lg)
		if err != nil {
			ctx.lg.Error(err)
			return
		}
		cache, err := workflowruntime.NewCache(path.Join(tmpdir, "cache"))
		if err != nil {
			ctx.lg.Error(err)
			return
		}
		rtwork.UseCache(cache)
		result, err := rtwork.Run(inbounds, false)
		if err != nil {
			ctx.lg.Error(err)
			return
		}
		_, err = http.Post(qry.Callback, "text/json; charset=utf-8", bytes.NewReader(result.Byte()))
		if err != nil {
			ctx.lg.Errorf("callback request: %v", err)
		}
		ctx.lg.Printf("custom test: %f/%f", result.Score, result.Fullscore)
		// pp.Print(result)
	}()
	return nil
}

func Sync(ctx *Context) error {
	type Sync struct {
		Checksum string `form:"sum" binding:"required"`
	}
	var qry Sync
	if err := ctx.BindQuery(&qry); err != nil {
		return err
	}
	// store problem
	file, err := os.CreateTemp(os.TempDir(), "prob-*.zip")
	if err != nil {
		return err
	}
	_, err = io.Copy(file, ctx.Request.Body)
	if err != nil {
		return err
	}
	file.Close()
	defer os.Remove(file.Name())

	probdir, _ := os.MkdirTemp(os.TempDir(), "prob-*")
	prob, err := problem.LoadFileTo(file.Name(), probdir)
	if err != nil {
		return err
	}

	chk := utils.FileChecksum(file.Name()).String()
	if qry.Checksum != chk {
		return &HttpError{http.StatusBadRequest, &DataError{chk, ErrInvalidChecksum}}
	}
	ctx.store.Set(qry.Checksum, prob)
	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
	return nil
}
