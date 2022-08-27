package preset

import "github.com/super-yaoj/yaoj-core/pkg/workflow"

var Customtest workflow.Workflow

func init() {
	var builder workflow.Builder
	builder.SetNode("compile", "compiler:auto", false, true)
	builder.SetNode("run", "runner:auto", true, false)

	builder.AddInbound(workflow.Gsubm, "source", "compile", "source")
	builder.AddInbound(workflow.Gsubm, "option", "compile", "option")

	builder.AddEdge("compile", "result", "run", "executable")
	builder.AddInbound(workflow.Gsubm, "input", "run", "stdin")
	builder.AddInbound(workflow.Gstatic, "runner_config", "run", "conf")

	res, err := builder.Workflow()
	if err != nil {
		panic(err)
	}
	Customtest = *res
}
