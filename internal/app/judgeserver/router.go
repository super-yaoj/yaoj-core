package judgeserver

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Judge(ctx *Context) error {
	type Judge struct {
		Callback string `form:"cb" binding:"required"`
		Checksum string `form:"sum" binding:"required"`
		// "pretest" "extra"
		Mode string `form:"mode"`
	}
	var qry Judge
	if err := ctx.BindQuery(&qry); err != nil {
		return &HttpError{http.StatusBadRequest, &Error{"bind query", err}}
	}

	submdata, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return err
	}
	// ready to judge
	ctx.lg.Debug("ready to judge")
	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})

	go func() {
		result, err := workerService.RunProblem(qry.Checksum, submdata, qry.Mode)

		if err != nil {
			ctx.lg.Errorf("run problem: %v", err)
			return
		}

		_, err = http.Post(qry.Callback, "text/json; charset=utf-8", bytes.NewReader(result.Byte()))
		if err != nil {
			ctx.lg.Errorf("callback request: %v", err)
		}
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

	// ready to judge
	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})

	go func() {
		result, err := workerService.CustomTest(dat)
		if err != nil {
			ctx.lg.Error(err)
			return
		}
		_, err = http.Post(qry.Callback, "text/json; charset=utf-8", bytes.NewReader(result.Byte()))
		if err != nil {
			ctx.lg.Errorf("callback request: %v", err)
		}
		ctx.lg.Infof("custom test: %f/%f", result.Score, result.Fullscore)
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

	err := workerService.SetProblem(qry.Checksum, ctx.Request.Body)
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
	return nil
}
