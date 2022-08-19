package processors

import (
	"time"

	"github.com/super-yaoj/yaoj-core/internal/pkg/judger"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
)

// Execute testlib checker
type CheckerTestlib struct {
	// input: checker input output answer
	// output: xmlreport stderr judgerlog
}

func (r CheckerTestlib) Label() (inputlabel []string, outputlabel []string) {
	return []string{"checker", "input", "output", "answer"},
		[]string{"xmlreport", "stderr", "judgerlog"}
}
func (r CheckerTestlib) Process(inputs Inbounds, outputs Outbounds) (result *Result) {
	inputs["checker"].SetMode(0744)

	chk := utils.RandomString(10)
	inf := utils.RandomString(10)
	ouf := utils.RandomString(10)
	asf := utils.RandomString(10)

	inputs["checker"].DupFile(chk, 0755)
	inputs["input"].DupFile(inf, 0644)
	inputs["output"].DupFile(ouf, 0644)
	inputs["answer"].DupFile(asf, 0644)

	res, err := judger.Judge(
		judger.WithArgument(
			"/dev/null", "/dev/null", outputs["stderr"].Path(),
			chk, inf, ouf, asf,
			outputs["xmlreport"].Path(), "-appes",
		),
		judger.WithJudger(judger.General),
		judger.WithPolicy("builtin:free"),
		judger.WithLog(outputs["judgerlog"].Path(), 0, false),
		judger.WithRealTime(time.Minute),
		judger.WithOutput(10*judger.MB),
	)
	if err != nil {
		return SysErrRes(err)
	}
	return res.ProcResult()
}

var _ Processor = CheckerTestlib{}
