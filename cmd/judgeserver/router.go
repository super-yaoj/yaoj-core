package main

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/super-yaoj/yaoj-core/pkg/buflog"
	"github.com/super-yaoj/yaoj-core/pkg/private/run"
	"github.com/super-yaoj/yaoj-core/pkg/problem"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
)

func Judge(ctx *gin.Context) {
	type Judge struct {
		Callback string `form:"cb" binding:"required"`
		Checksum string `form:"sum" binding:"required"`
		// default: options: "pretest" "extra" "hack"
		// "hack": 返回 workflow.Result
		Mode string `form:"mode"`
	}
	var qry Judge
	err := ctx.BindQuery(&qry)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !storage.Has(qry.Checksum) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":      "checksum not found",
			"error_code": 1,
		})
		return
	}

	// load submission
	data, _ := io.ReadAll(ctx.Request.Body)
	submission, err := problem.LoadSubmData(data)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var std, hackee problem.Submission
	if qry.Mode == "hack" {
		if submission[workflow.Gsubm] == nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid submission"})
			return
		}
		// 外层的submission就 workflow.Gsubm 里 "std" "hackee" 两个字段
		std, err = problem.LoadSubmData((*submission[workflow.Gsubm])["std"].Ctnt)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		hackee, err = problem.LoadSubmData((*submission[workflow.Gsubm])["hackee"].Ctnt)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// load problem
	prob := storage.Get(qry.Checksum)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ready to judge
	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})

	go func() {
		start_time := time.Now()

		tmpdir, _ := os.MkdirTemp("", "yaoj-runtime-*")
		defer os.RemoveAll(tmpdir)

		if qry.Mode == "hack" {
			result, err := run.RunHack(prob.Data(), tmpdir, hackee, std)
			if err != nil {
				logger.Printf("run hack error: %v", err)
				return
			}
			logger.Printf("%+v", result)

			_, err = http.Post(qry.Callback, "text/json; charset=utf-8", bytes.NewReader(result.Byte()))
			if err != nil {
				logger.Printf("callback request error: %v", err)
			}
		} else {
			result, err := run.RunProblem(prob.Data(), tmpdir, submission, qry.Mode)
			if err != nil {
				logger.Printf("run problem error: %v", err)
				return
			}
			logger.Print(result.Brief())

			_, err = http.Post(qry.Callback, "text/json; charset=utf-8", bytes.NewReader(result.Byte()))
			if err != nil {
				logger.Printf("callback request error: %v", err)
			}
		}

		logger.Printf("Total judging time: %v", time.Since(start_time))
	}()
}

func CustomTest(ctx *gin.Context) {
	type CustomTest struct {
		Callback string `form:"cb" binding:"required"`
	}
	var qry CustomTest
	err := ctx.BindQuery(&qry)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// load submission
	data, _ := io.ReadAll(ctx.Request.Body)
	submission, err := problem.LoadSubmData(data)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ready to judge
	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})

	go func() {
		tmpdir, _ := os.MkdirTemp(os.TempDir(), "custom-*")
		defer os.RemoveAll(tmpdir)

		os.WriteFile(path.Join(tmpdir, "_limit"),
			[]byte("10000 10000 504857600 504857600 504857600 54857600 10"), os.ModePerm)

		pathmap := submission.Download(tmpdir)
		pathmap[workflow.Gstatic] = &map[string]string{
			"limit": path.Join(tmpdir, "_limit"),
		}

		result, err := run.RunWorkflow(*customTestWkfl, tmpdir, pathmap, 100)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		_, err = http.Post(qry.Callback, "text/json; charset=utf-8", bytes.NewReader(result.Byte()))
		if err != nil {
			logger.Printf("callback request error: %v", err)
		}
		logger.Printf("custom test: %f/%f", result.Score, result.Fullscore)
		// pp.Print(result)
	}()
}

func Sync(ctx *gin.Context) {
	type Sync struct {
		Checksum string `form:"sum" binding:"required"`
	}
	var qry Sync
	err := ctx.BindQuery(&qry)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// store problem
	file, _ := os.CreateTemp(os.TempDir(), "prob-*.zip")
	io.Copy(file, ctx.Request.Body)
	file.Close()
	defer os.Remove(file.Name())

	probdir, _ := os.MkdirTemp(os.TempDir(), "prob-*")
	prob, err := problem.LoadDump(file.Name(), probdir)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	chk := utils.FileChecksum(file.Name()).String()
	if qry.Checksum != chk {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":    "invalid checksum",
			"checksum": chk,
		})
		return
	}
	storage.Set(qry.Checksum, prob)
	ctx.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

func Log(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"logs":    buflog.Tail(),
	})
}

var customTestWkfl *workflow.Workflow

func init() {
	// custom test workflow
	var customTestWk workflow.Builder
	customTestWk.SetNode("compile", "compiler:auto", false, false)
	customTestWk.SetNode("run", "runner:stdio", true, false)
	customTestWk.AddInbound(workflow.Gsubm, "source", "compile", "source")
	customTestWk.AddInbound(workflow.Gsubm, "input", "run", "stdin")
	customTestWk.AddInbound(workflow.Gstatic, "limit", "run", "limit")
	customTestWk.AddEdge("compile", "result", "run", "executable")
	graph, err := customTestWk.WorkflowGraph()
	if err != nil {
		panic(err)
	}
	customTestWkfl = &workflow.Workflow{
		WorkflowGraph: graph,
		Analyzer:      workflow.DefaultAnalyzer{},
	}
}
