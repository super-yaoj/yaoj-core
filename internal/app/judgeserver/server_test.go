package judgeserver_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"testing"
	"time"

	"github.com/super-yaoj/yaoj-core/internal/app/judgeserver"
	"github.com/super-yaoj/yaoj-core/internal/tests"
	"github.com/super-yaoj/yaoj-core/pkg/log"
	"github.com/super-yaoj/yaoj-core/pkg/problem"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
)

func TestServer(t *testing.T) {
	lg := log.NewTest()
	// create server
	server := judgeserver.New(lg)
	err := judgeserver.Init(t.TempDir(), lg)
	if err != nil {
		t.Fatal(err)
	}

	// a+b problem 的校验值，在测试 Sync 后初始化
	var Checksum string

	t.Run("Sync", func(t *testing.T) {
		tmpdir := t.TempDir()
		filename := path.Join(tmpdir, "prob.zip")
		prob, err := tests.CreateProblem(path.Join(tmpdir, "prob"), lg)
		if err != nil {
			t.Fatal(err)
		}
		err = prob.DumpFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		file, err := os.Open(filename)
		checksum := utils.FileChecksum(filename)
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest("POST", "/sync?sum="+checksum.String(), file)
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			data, _ := io.ReadAll(rec.Result().Body)
			t.Fatal(string(data))
		}
		Checksum = checksum.String()
	})

	cbserver := http.NewServeMux()
	cbaddr := "localhost:3926"
	// start server asyncly
	go func() {
		lg.Info("start callback listening server")
		err := http.ListenAndServe(cbaddr, cbserver)
		if err != nil {
			lg.Fatal(err)
		}
	}()

	t.Run("Judge", func(t *testing.T) {
		finish := make(chan int)
		// add handler
		cbserver.HandleFunc("/judge_cb", func(w http.ResponseWriter, r *http.Request) {
			t.Logf("some thing")
			resdata, _ := io.ReadAll(r.Body)
			lg.Debug(string(resdata))
			result := problem.Result{}
			err := json.Unmarshal(resdata, &result)
			if err != nil {
				lg.Error(err)
				finish <- 1
				return
			}
			if result.Score != result.Fullscore {
				lg.Error(err)
				finish <- 1
				return
			}
			finish <- 0
		})
		// wait for callback server
		time.Sleep(time.Millisecond * 100)

		submission := tests.CreateSubmission()
		var buf bytes.Buffer
		submission.DumpTo(&buf)

		req, err := http.NewRequest("POST", "/judge?sum="+Checksum+"&cb="+url.QueryEscape("http://"+cbaddr+"/judge_cb"), &buf)
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			data, _ := io.ReadAll(rec.Result().Body)
			lg.Error(string(data))
			t.Fatal(string(data))
		}
		// wait judgement finish
		rescode := <-finish
		if rescode != 0 {
			t.Fatal("res code not zero")
		}
	})

	t.Run("Judge(BadRequest)", func(t *testing.T) {
		// bind error
		req, err := http.NewRequest("POST", "/judge?sum="+Checksum, nil)
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			data, _ := io.ReadAll(rec.Result().Body)
			lg.Error(string(data))
			t.Fatal(string(data))
		}
	})

	t.Run("CustomTest", func(t *testing.T) {
		finish := make(chan int)
		// add handler
		cbserver.HandleFunc("/custom_cb", func(w http.ResponseWriter, r *http.Request) {
			t.Logf("some thing")
			resdata, _ := io.ReadAll(r.Body)
			lg.Debug(string(resdata))
			result := workflow.Result{}
			err := json.Unmarshal(resdata, &result)
			if err != nil {
				lg.Error(err)
				finish <- 1
				return
			}
			if result.Score != result.Fullscore {
				lg.Error(err)
				finish <- 1
				return
			}
			finish <- 0
		})
		submission := tests.CreateSubmission()
		submission.SetData(workflow.Gsubm, "input", []byte("114 514"))
		var buf bytes.Buffer
		submission.DumpTo(&buf)

		req, err := http.NewRequest("POST", "/custom?cb="+url.QueryEscape("http://"+cbaddr+"/custom_cb"), &buf)
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			data, _ := io.ReadAll(rec.Result().Body)
			lg.Error(string(data))
			t.Fatal(string(data))
		}
		// wait judgement finish
		rescode := <-finish
		if rescode != 0 {
			t.Fatal("res code not zero")
		}
	})
}
