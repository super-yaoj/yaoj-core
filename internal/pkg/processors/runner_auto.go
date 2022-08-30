package processors

import (
	"github.com/super-yaoj/yaoj-core/internal/pkg/judger"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
)

// Run a program automatically.
type RunnerAuto struct {
	// input: executable, stdin, conf
	// output: stdout, stderr, judgerlog
}

func (r RunnerAuto) Label() (inputlabel []string, outputlabel []string) {
	return []string{"executable", "stdin", "conf"}, []string{"stdout", "stderr", "judgerlog"}
}

func (r RunnerAuto) Process(inputs Inbounds, outputs Outbounds) *Result {
	// make it executable
	inputs["executable"].SetMode(0744)
	// to file
	_, err := outputs["stdout"].File()
	if err != nil {
		return RtErrRes(err)
	}
	_, err = outputs["stderr"].File()
	if err != nil {
		return RtErrRes(err)
	}
	_, err = outputs["judgerlog"].File()
	if err != nil {
		return RtErrRes(err)
	}

	// parse config
	data, err := inputs["conf"].Get()
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
		judger.WithLog(outputs["judgerlog"].Path(), 0),
	}

	if conf.IsFileIO() {
		if _, err := utils.CopyFile(inputs["stdin"].Path(), conf.Inf); err != nil {
			return RtErrRes(err)
		}
		options = append(options, judger.WithArgument("/dev/null", "/dev/null",
			outputs["stderr"].Path(), inputs["executable"].Path()))
	} else { // stdio
		options = append(options, judger.WithArgument(
			inputs["stdin"].Path(),
			outputs["stdout"].Path(),
			outputs["stderr"].Path(),
			inputs["executable"].Path(),
		))
	}

	options = append(options, runLimOptions(conf)...)

	res, err := judger.Judge(options...)
	if err != nil {
		return SysErrRes(err)
	}

	if conf.IsFileIO() {
		utils.CopyFile(conf.Ouf, outputs["stdout"].Path())
	}
	return res.ProcResult()
}

var _ Processor = RunnerAuto{}
