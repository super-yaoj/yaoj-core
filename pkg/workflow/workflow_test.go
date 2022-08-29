package workflow_test

import (
	"encoding/json"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/super-yaoj/yaoj-core/pkg/workflow"
)

func TestWorkflow(t *testing.T) {
	var builder workflow.Builder
	builder.SetNode("compile", "compiler:auto", false, true)
	builder.SetNode("run", "runner:auto", true, false)
	builder.SetNode("check", "checker:testlib", false, false)
	builder.AddInbound(workflow.Gsubm, "source", "compile", "source")
	builder.AddInbound(workflow.Gsubm, "option", "compile", "option")
	builder.AddInbound(workflow.Gstatic, "runconf", "run", "conf")
	builder.AddInbound(workflow.Gstatic, "chk", "check", "checker")
	builder.AddInbound(workflow.Gtests, "input", "run", "stdin")
	builder.AddInbound(workflow.Gtests, "input", "check", "input")
	builder.AddInbound(workflow.Gtests, "answer", "check", "answer")
	builder.AddEdge("compile", "result", "run", "executable")
	builder.AddEdge("run", "stdout", "check", "output")
	work, err := builder.Workflow()
	if err != nil {
		t.Fatal(err)
	}

	var work2 *workflow.Workflow
	var workdata []byte

	t.Run("Load", func(t *testing.T) {
		workdata, err = json.Marshal(work)
		if err != nil {
			t.Fatal(err)
		}

		work2, err = workflow.Load(workdata)
		if err != nil {
			t.Fatal(err)
		}

		_, err = workflow.Load([]byte(""))
		if err == nil || !strings.Contains(err.Error(), "Load") {
			t.Fatal("invalid err:", err)
		}
	})

	t.Run("LoadFile", func(t *testing.T) {
		_, err = workflow.LoadFile("")
		if err == nil || !strings.Contains(err.Error(), "LoadFile") {
			t.Fatal("invalid err:", err)
		}

		filename := path.Join(t.TempDir(), "file")
		err = os.WriteFile(filename, workdata, 0777)
		if err != nil {
			t.Fatal(err)
		}
		_, err = workflow.LoadFile(filename)
		if err != nil {
			t.Fatal(err)
		}
	})

	// others
	t.Log(work2.EdgeFrom("run"), work2.EdgeTo("run"))
	t.Log(err, (&workflow.Result{}).Byte())

}
