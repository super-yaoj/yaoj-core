package workflow_test

import (
	"log"

	"github.com/super-yaoj/yaoj-core/pkg/workflow"
)

func ExampleBuilder() {
	var b workflow.Builder
	b.SetNode("compile", "compiler", false, true)
	b.SetNode("run", "runner:stdio", true, false)
	b.SetNode("check", "checker:hcmp", false, false)
	b.AddInbound(workflow.Gsubm, "source", "compile", "source")
	b.AddInbound(workflow.Gstatic, "compilescript", "compile", "script")
	b.AddInbound(workflow.Gstatic, "limitation", "run", "limit")
	b.AddInbound(workflow.Gtests, "input", "run", "stdin")
	b.AddInbound(workflow.Gtests, "answer", "check", "ans")
	b.AddEdge("compile", "result", "run", "executable")
	b.AddEdge("run", "stdout", "check", "out")
	_, err := b.Workflow()
	if err != nil {
		log.Print(err)
	}
}
