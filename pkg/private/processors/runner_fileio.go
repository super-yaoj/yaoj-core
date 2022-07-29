package processors

import (
	"os"

	"github.com/super-yaoj/yaoj-core/pkg/private/judger"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
)

// Run a program reading from file and print to file and stderr.
// File "config" contains two lines, the first of which acts the same as
// "limit" of RunnerStdio while the second contains two strings denoting input
// file and output file.
type RunnerFileio struct {
	// input: executable, fin, config
	// output: fout, stderr, judgerlog
}

func (r RunnerFileio) Label() (inputlabel []string, outputlabel []string) {
	return []string{"executable", "fin", "config"}, []string{"fout", "stderr", "judgerlog"}
}

func (r RunnerFileio) Run(input []string, output []string) *Result {
	// make it executable
	os.Chmod(input[0], 0744)

	data, err := os.ReadFile(input[2])
	if err != nil {
		return RtErrRes(err)
	}
	var lim RunConf
	if err := lim.Deserialize(data); err != nil {
		return RtErrRes(err)
	}
	var inf, ouf string = lim.Inf, lim.Ouf
	logger.Printf("inf=%q, out=%q", inf, ouf)
	if _, err := utils.CopyFile(input[1], inf); err != nil {
		return RtErrRes(err)
	}
	options := []judger.OptionProvider{
		judger.WithArgument("/dev/null", "/dev/null", output[1], input[0]),
		judger.WithJudger(judger.General),
		judger.WithPolicy("builtin:yaoj"),
		judger.WithLog(output[2], 0, false),
	}
	options = append(options, runLimOptions(lim)...)
	res, err := judger.Judge(options...)
	if err != nil {
		return SysErrRes(err)
	}
	utils.CopyFile(ouf, output[0])
	return res.ProcResult()
}

var _ Processor = RunnerFileio{}
