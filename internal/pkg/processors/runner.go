package processors

import (
	"os"

	"github.com/super-yaoj/yaoj-core/internal/pkg/judger"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
)

// Run a program automatically.
type Runner struct {
	// input: executable, stdin, conf
	// output: stdout, stderr, judgerlog
}

func (r Runner) Label() (inputlabel []string, outputlabel []string) {
	return []string{"executable", "stdin", "conf"}, []string{"stdout", "stderr", "judgerlog"}
}
func (r Runner) Run(input []string, output []string) *Result {
	// make it executable
	os.Chmod(input[0], 0744)

	// parse config
	data, err := os.ReadFile(input[2])
	if err != nil {
		return RtErrRes(err)
	}
	var conf RunConf
	if err := conf.Deserialize(data); err != nil {
		return RtErrRes(err)
	}

	options := []judger.OptionProvider{
		judger.WithJudger(judger.General),
		judger.WithPolicy("builtin:yaoj"),
		judger.WithLog(output[2], 0, false),
	}

	if conf.IsFileIO() {
		if _, err := utils.CopyFile(input[1], conf.Inf); err != nil {
			return RtErrRes(err)
		}
		options = append(options, judger.WithArgument("/dev/null", "/dev/null", output[1], input[0]))
	} else { // stdio
		options = append(options, judger.WithArgument(input[1], output[0], output[1], input[0]))
	}

	options = append(options, runLimOptions(conf)...)

	res, err := judger.Judge(options...)
	if err != nil {
		return SysErrRes(err)
	}

	if conf.IsFileIO() {
		utils.CopyFile(conf.Ouf, output[0])
	}
	return res.ProcResult()
}

var _ Processor = Runner{}
