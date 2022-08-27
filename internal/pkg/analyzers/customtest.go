package analyzers

import (
	workflowruntime "github.com/super-yaoj/yaoj-core/internal/pkg/worker/workflow"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
)

type Customtest struct {
}

func (r Customtest) Analyze(w *workflowruntime.RtWorkflow) workflow.Result {
	ndCompile := w.RtNodes["compile"]
	ndRun := w.RtNodes["run"]

	fStdin := show(ndRun.Input["stdin"], "stdin", 1000)
	fStdout := show(ndRun.Output["stdout"], "stdout", 1000)
	fStderr := show(ndRun.Output["stderr"], "stderr", 1000)

	if !ndCompile.Result.Ok() {
		return workflow.Result{
			ResultMeta: workflow.ResultMeta{
				Title:     "Compile Error",
				Score:     0,
				Fullscore: w.Fullscore,
			},
			File: []workflow.ResultFile{
				show(ndCompile.Output["log"], "compile log", 1000),
			},
		}
	} else if !ndRun.Result.Ok() {
		return workflow.Result{
			ResultMeta: workflow.ResultMeta{
				Title:     codeName[ndRun.Result.Code],
				Score:     0,
				Fullscore: w.Fullscore,
			},
			File: []workflow.ResultFile{
				fStdin,
				fStderr,
				fStdout,
			},
		}
	} else {
		return workflow.Result{
			ResultMeta: workflow.ResultMeta{
				Title:     codeName[ndRun.Result.Code],
				Score:     w.Fullscore,
				Fullscore: w.Fullscore,
			},
			File: []workflow.ResultFile{
				fStdin,
				fStderr,
				fStdout,
			},
		}
	}
}
