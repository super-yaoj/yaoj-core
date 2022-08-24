package preset

import "github.com/super-yaoj/yaoj-core/pkg/workflow"

// 传统题的 workflow
//
//	Gstatic:
//	  checker       校验器源码（testlib）
//	  runner_config 时空限制，文件 IO 等设置
//	Gsubm:
//	  option 源代码的语言等属性（用于哈希）
//	  source 源代码
//	Gtests:
//	  input  读入文件
//	  output 输出文件
var Traditional workflow.Workflow

func init() {
	var builder workflow.Builder
	builder.SetNode("compile", "compiler:auto", false, true)
	builder.SetNode("run", "runner:auto", true, false)
	builder.SetNode("check", "checker:testlib", false, false)
	builder.SetNode("checker_compile", "compiler:testlib", false, true)

	builder.AddEdge("checker_compile", "result", "check", "checker")
	builder.AddEdge("run", "stdout", "check", "output")
	builder.AddInbound(workflow.Gtests, "input", "check", "input")
	builder.AddInbound(workflow.Gtests, "output", "check", "answer")

	builder.AddInbound(workflow.Gsubm, "source", "compile", "source")
	builder.AddInbound(workflow.Gsubm, "option", "compile", "option")

	builder.AddInbound(workflow.Gstatic, "checker", "checker_compile", "source")

	builder.AddEdge("compile", "result", "run", "executable")
	builder.AddInbound(workflow.Gtests, "input", "run", "stdin")
	builder.AddInbound(workflow.Gstatic, "runner_config", "run", "conf")

	res, err := builder.Workflow()
	if err != nil {
		panic(err)
	}
	Traditional = *res
}
