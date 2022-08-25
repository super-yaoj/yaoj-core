package analyzers

import (
	"encoding/xml"
	"fmt"
	"io"

	workflowruntime "github.com/super-yaoj/yaoj-core/internal/pkg/worker/workflow"
	"github.com/super-yaoj/yaoj-core/pkg/processor"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
	"golang.org/x/text/encoding/charmap"
)

// workflow/preset 传统题的分析器
type Traditional struct {
}

var codeName = map[processor.Code]string{
	processor.Ok:               "Accepted",
	processor.TimeExceed:       "Time Limit Exceed",
	processor.RuntimeError:     "Runtime Error",
	processor.MemoryExceed:     "Memory Limit Exceed",
	processor.SystemError:      "System Error",
	processor.DangerousSyscall: "Dangerous System Call",
	processor.OutputExceed:     "Output Limit Exceed",
	processor.ExitError:        "Exit Code Error",
}

func (r Traditional) Analyze(w *workflowruntime.RtWorkflow) workflow.Result {
	ndCompile := w.RtNodes["compile"]
	ndCheckCompile := w.RtNodes["checker_compile"]
	ndRun := w.RtNodes["run"]
	ndCheck := w.RtNodes["check"]

	fStdin := show(ndRun.Input["stdin"], "stdin", 1000)
	fStdout := show(ndRun.Output["stdout"], "stdout", 1000)
	fStderr := show(ndRun.Output["stderr"], "stderr", 1000)
	fAnswer := show(ndCheck.Input["answer"], "answer", 1000)

	if !ndCheckCompile.Result.Ok() {
		return workflow.Result{
			ResultMeta: workflow.ResultMeta{
				Title:     "Checker Compile Error",
				Score:     0,
				Fullscore: w.Fullscore,
			},
			File: []workflow.ResultFile{
				show(ndCheckCompile.Output["log"], "compile log", 1000),
			},
		}
	} else if !ndCompile.Result.Ok() {
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
	} else if !ndCheck.Result.Ok() && ndCheck.Result.Code != processor.ExitError {
		return workflow.Result{
			ResultMeta: workflow.ResultMeta{
				Title:     "Checker " + codeName[ndRun.Result.Code],
				Score:     0,
				Fullscore: w.Fullscore,
			},
		}
	} else { // Wrong Answer or Accepted
		// parse xml file
		type Result struct {
			XMLName xml.Name `xml:"result"`
			Msg     string   `xml:",chardata"`
			Outcome string   `xml:"outcome,attr"`
		}
		var result Result
		file, _ := ndCheck.Output["xmlreport"].File()
		defer file.Close()
		// parse xml encoded windows1251
		d := xml.NewDecoder(file)
		d.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
			switch charset {
			case "windows-1251":
				return charmap.Windows1251.NewDecoder().Reader(input), nil
			default:
				return nil, fmt.Errorf("unknown charset: %s", charset)
			}
		}
		d.Decode(&result)

		fMsg := workflow.ResultFile{
			Title:   "checker message",
			Content: result.Msg,
		}
		if ndCheck.Result.Ok() {
			return workflow.Result{
				ResultMeta: workflow.ResultMeta{
					Title:     "Accepted",
					Score:     w.Fullscore,
					Fullscore: w.Fullscore,
					Time:      *ndRun.Result.CpuTime,
					Memory:    *ndRun.Result.Memory,
				},
				File: []workflow.ResultFile{
					fStdin,
					fStderr,
					fStdout,
					fAnswer,
					fMsg,
				},
			}
		}
		return workflow.Result{
			ResultMeta: workflow.ResultMeta{
				Title:  "Wrong Answer",
				Score:  0,
				Time:   *ndRun.Result.CpuTime,
				Memory: *ndRun.Result.Memory,
			},
			File: []workflow.ResultFile{
				fStdin,
				fStderr,
				fStdout,
				fAnswer,
				fMsg,
			},
		}
	}
}
