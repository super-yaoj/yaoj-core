package processors

import (
	"os"

	"github.com/super-yaoj/yaoj-core/internal/pkg/judger"
)

// Run a program reading from stdin and print to stdout and stderr.
// For "limit", it contains a series of number seperated by space, denoting
// real time (ms), cpu time (ms), virtual memory (byte), real memory (byte),
// stack memory (byte), output limit (byte), fileno limitation respectively.
type RunnerStdio struct {
	// input: executable, stdin, limit
	// output: stdout, stderr, judgerlog
}

func (r RunnerStdio) Label() (inputlabel []string, outputlabel []string) {
	return []string{"executable", "stdin", "limit"}, []string{"stdout", "stderr", "judgerlog"}
}
func (r RunnerStdio) Run(input []string, output []string) *Result {
	// make it executable
	os.Chmod(input[0], 0744)

	data, err := os.ReadFile(input[2])
	if err != nil {
		return RtErrRes(err)
	}
	options := []judger.OptionProvider{
		judger.WithArgument(input[1], output[0], output[1], input[0]),
		judger.WithJudger(judger.General),
		judger.WithPolicy("builtin:yaoj"),
		judger.WithLog(output[2], 0, false),
	}
	var lim RunConf
	if err := lim.Deserialize(data); err != nil {
		return RtErrRes(err)
	}

	options = append(options, runLimOptions(lim)...)
	res, err := judger.Judge(options...)
	if err != nil {
		return SysErrRes(err)
	}
	return res.ProcResult()
}

var _ Processor = RunnerStdio{}
