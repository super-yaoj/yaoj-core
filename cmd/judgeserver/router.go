package main

import (
	"bytes"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sshwy/yaoj-core/pkg/buflog"
	"github.com/sshwy/yaoj-core/pkg/private/run"
	"github.com/sshwy/yaoj-core/pkg/problem"
	"github.com/sshwy/yaoj-core/pkg/utils"
	"github.com/sshwy/yaoj-core/pkg/workflow"
)

func Judge(ctx *gin.Context) {
	type Judge struct {
		Callback string `form:"cb" binding:"required"`
		Checksum string `form:"sum" binding:"required"`
		// default: judge. options: "custom" "hack"
		Type string `form:"type"`
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

	// remove after judging
	tmpdir, _ := os.MkdirTemp("", "yaoj-runtime-*")

	// load submission
	file, _ := os.CreateTemp(os.TempDir(), "judge-*")
	_, err = io.Copy(file, ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	file.Close()
	defer os.Remove(file.Name())
	submission, err := problem.LoadSubm(file.Name(), tmpdir)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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
		if qry.Type == "custom" {
			result, err := run.RunCustom(prob.Data(), tmpdir, *submission[workflow.Gsubm], *submission[workflow.Gtests])
			if err != nil {
				logger.Printf("run problem: %v", err)
				return
			}

			_, err = http.Post(qry.Callback, "text/json; charset=utf-8", bytes.NewReader(result.Byte()))
			if err != nil {
				logger.Printf("callback request error: %v", err)
			}
		} else if qry.Type == "hack" {
			http.Post(qry.Callback, "text/plain; charset=utf-8", bytes.NewReader([]byte("not implemented")))
		} else {
			result, err := run.RunProblem(prob.Data(), tmpdir, *submission[workflow.Gsubm])
			if err != nil {
				logger.Printf("run problem: %v", err)
				return
			}
			logger.Print(result.Brief())

			_, err = http.Post(qry.Callback, "text/json; charset=utf-8", bytes.NewReader(result.Byte()))
			if err != nil {
				logger.Printf("callback request error: %v", err)
			}
		}
		os.RemoveAll(tmpdir)
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
