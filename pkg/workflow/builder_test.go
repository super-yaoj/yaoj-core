package workflow_test

import (
	"errors"
	"testing"

	"github.com/super-yaoj/yaoj-core/pkg/workflow"
)

func TestBuilder(t *testing.T) {
	t.Run("Common", func(t *testing.T) {
		var builder workflow.Builder
		builder.SetNode("compile", "compiler", false, true)
		builder.SetNode("run", "runner:stdio", true, false)
		builder.SetNode("check", "checker:hcmp", false, false)
		builder.AddInbound(workflow.Gsubm, "source", "compile", "source")
		builder.AddInbound(workflow.Gstatic, "compilescript", "compile", "script")
		builder.AddInbound(workflow.Gstatic, "limitation", "run", "limit")
		builder.AddInbound(workflow.Gtests, "input", "run", "stdin")
		builder.AddInbound(workflow.Gtests, "answer", "check", "ans")
		builder.AddEdge("compile", "result", "run", "executable")
		builder.AddEdge("run", "stdout", "check", "out")
		graph, err := builder.WorkflowGraph()
		if err != nil {
			t.Fatal(err)
		}
		_ = graph.Serialize()
	})
	t.Run("InvalidGroupname", func(t *testing.T) {
		var builder workflow.Builder
		builder.AddInbound("badgroup", "", "", "")
		_, err := builder.WorkflowGraph()
		if !errors.Is(err, workflow.ErrInvalidGroupname) {
			t.Fatal(err)
		}
	})
	t.Run("InvalidEdge(From)", func(t *testing.T) {
		var builder workflow.Builder
		builder.AddEdge("badfrom", "", "runner", "")
		_, err := builder.WorkflowGraph()
		if !errors.Is(err, workflow.ErrInvalidEdge) {
			t.Fatal(err)
		}
	})
	t.Run("InvalidEdge(To)", func(t *testing.T) {
		var builder workflow.Builder
		builder.SetNode("runner", "runner:stdio", true, false)
		builder.AddEdge("runner", "", "badto", "")
		_, err := builder.WorkflowGraph()
		if !errors.Is(err, workflow.ErrInvalidEdge) {
			t.Fatal(err)
		}
	})
	t.Run("InvalidInboundEdge", func(t *testing.T) {
		var builder workflow.Builder
		builder.AddInbound(workflow.Gstatic, "", "", "")
		_, err := builder.WorkflowGraph()
		if !errors.Is(err, workflow.ErrInvalidEdge) {
			t.Fatal(err)
		}
	})
	t.Run("InvalidInboundInput", func(t *testing.T) {
		var builder workflow.Builder
		builder.SetNode("runner", "runner:stdio", true, false)
		builder.AddInbound(workflow.Gstatic, "", "runner", "")
		_, err := builder.WorkflowGraph()
		if !errors.Is(err, workflow.ErrInvalidInputLabel) {
			t.Fatal(err)
		}
	})
	t.Run("InvalidOutputLabel", func(t *testing.T) {
		var builder workflow.Builder
		builder.SetNode("runner", "runner:stdio", true, false)
		builder.AddEdge("runner", "", "runner", "stdin")
		_, err := builder.WorkflowGraph()
		if !errors.Is(err, workflow.ErrInvalidOutputLabel) {
			t.Fatal(err)
		}
	})
	t.Run("InvalidInputLabel", func(t *testing.T) {
		var builder workflow.Builder
		builder.SetNode("runner", "runner:stdio", true, false)
		builder.AddEdge("runner", "stdout", "runner", "")
		_, err := builder.WorkflowGraph()
		if !errors.Is(err, workflow.ErrInvalidInputLabel) {
			t.Fatal(err)
		}
	})
	t.Run("DuplicateDest", func(t *testing.T) {
		var builder workflow.Builder
		builder.SetNode("runner", "runner:stdio", true, false)
		builder.AddEdge("runner", "stdout", "runner", "stdin")
		builder.AddEdge("runner", "stdout", "runner", "stdin")
		_, err := builder.WorkflowGraph()
		if !errors.Is(err, workflow.ErrDuplicateDest) {
			t.Fatal(err)
		}
	})
	t.Run("DuplicateDest(Inbound)", func(t *testing.T) {
		var builder workflow.Builder
		builder.SetNode("runner", "runner:stdio", true, false)
		builder.AddEdge("runner", "stdout", "runner", "stdin")
		builder.AddInbound(workflow.Gstatic, "stdout", "runner", "stdin")
		_, err := builder.WorkflowGraph()
		if !errors.Is(err, workflow.ErrDuplicateDest) {
			t.Fatal(err)
		}
	})
	t.Run("DuplicateDest(Inbound)", func(t *testing.T) {
		var builder workflow.Builder
		builder.SetNode("runner", "runner:stdio", true, false)
		_, err := builder.WorkflowGraph()
		if !errors.Is(err, workflow.ErrIncompleteNodeInput) {
			t.Fatal(err)
		}
	})
}
