package workflow_test

import (
	"testing"

	"github.com/super-yaoj/yaoj-core/pkg/workflow"
	"github.com/super-yaoj/yaoj-core/pkg/yerrors"
)

func TestBuilder(t *testing.T) {
	t.Run("Common", func(t *testing.T) {
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
		_, err := builder.Workflow()
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("InvalidGroupname", func(t *testing.T) {
		var builder workflow.Builder
		builder.AddInbound("badgroup", "", "", "")
		_, err := builder.Workflow()
		if !yerrors.Is(err, workflow.ErrInvalidGroupname) {
			t.Fatal(err)
		}
	})
	t.Run("InvalidEdge(From)", func(t *testing.T) {
		var builder workflow.Builder
		builder.AddEdge("badfrom", "", "runner", "")
		_, err := builder.Workflow()
		if !yerrors.Is(err, workflow.ErrInvalidEdge) {
			t.Fatal(err)
		}
	})
	t.Run("InvalidEdge(To)", func(t *testing.T) {
		var builder workflow.Builder
		builder.SetNode("runner", "runner:stdio", true, false)
		builder.AddEdge("runner", "", "badto", "")
		_, err := builder.Workflow()
		if !yerrors.Is(err, workflow.ErrInvalidEdge) {
			t.Fatal(err)
		}
	})
	t.Run("InvalidInboundEdge", func(t *testing.T) {
		var builder workflow.Builder
		builder.AddInbound(workflow.Gstatic, "", "", "")
		_, err := builder.Workflow()
		if !yerrors.Is(err, workflow.ErrInvalidEdge) {
			t.Fatal(err)
		}
	})
	t.Run("InvalidInboundInput", func(t *testing.T) {
		var builder workflow.Builder
		builder.SetNode("runner", "runner:stdio", true, false)
		builder.AddInbound(workflow.Gstatic, "", "runner", "")
		_, err := builder.Workflow()
		if !yerrors.Is(err, workflow.ErrInvalidInputLabel) {
			t.Fatal(err)
		}
	})
	t.Run("InvalidOutputLabel", func(t *testing.T) {
		var builder workflow.Builder
		builder.SetNode("runner", "runner:stdio", true, false)
		builder.AddEdge("runner", "", "runner", "stdin")
		_, err := builder.Workflow()
		if !yerrors.Is(err, workflow.ErrInvalidOutputLabel) {
			t.Fatal(err)
		}
	})
	t.Run("InvalidInputLabel", func(t *testing.T) {
		var builder workflow.Builder
		builder.SetNode("runner", "runner:auto", true, false)
		builder.AddEdge("runner", "stdout", "runner", "")
		_, err := builder.Workflow()
		if !yerrors.Is(err, workflow.ErrInvalidInputLabel) {
			t.Fatal(err)
		}
	})
	t.Run("DuplicateDest", func(t *testing.T) {
		var builder workflow.Builder
		builder.SetNode("runner", "runner:auto", true, false)
		builder.AddEdge("runner", "stdout", "runner", "stdin")
		builder.AddEdge("runner", "stdout", "runner", "stdin")
		_, err := builder.Workflow()
		if !yerrors.Is(err, workflow.ErrDuplicateDest) {
			t.Fatal(err)
		}
	})
	t.Run("DuplicateDest(Inbound)", func(t *testing.T) {
		var builder workflow.Builder
		builder.SetNode("runner", "runner:auto", true, false)
		builder.AddEdge("runner", "stdout", "runner", "stdin")
		builder.AddInbound(workflow.Gstatic, "stdout", "runner", "stdin")
		_, err := builder.Workflow()
		if !yerrors.Is(err, workflow.ErrDuplicateDest) {
			t.Fatal(err)
		}
	})
	t.Run("IncompleteNodeInput", func(t *testing.T) {
		var builder workflow.Builder
		builder.SetNode("runner", "runner:auto", true, false)
		_, err := builder.Workflow()
		if !yerrors.Is(err, workflow.ErrIncompleteNodeInput) {
			t.Fatal(err)
		}
	})
}
