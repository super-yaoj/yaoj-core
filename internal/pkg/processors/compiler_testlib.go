package processors

import (
	"os"
	"time"

	_ "embed"

	"github.com/super-yaoj/yaoj-core/internal/pkg/judger"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
)

//go:embed testlib.h
var testlib []byte

// Compile codeforces testlib source file using g++.
// For input files, "source" represents source file.
type CompilerTestlib struct {
	// input: source
	// output: result, log, judgerlog
}

func (r CompilerTestlib) Label() (inputlabel []string, outputlabel []string) {
	return []string{"source"}, []string{"result", "log", "judgerlog"}
}

func (r CompilerTestlib) Process(inputs Inbounds, outputs Outbounds) (result *Result) {
	// create testlib.h
	err := os.WriteFile("testlib.h", testlib, os.ModePerm)
	if err != nil {
		return SysErrRes(err)
	}
	// create src (*.cpp)
	src := utils.RandomString(10) + ".cpp"
	data, err := inputs["source"].Get()
	if err != nil {
		return SysErrRes(err)
	}
	err = os.WriteFile(src, data, os.ModePerm)
	if err != nil {
		return SysErrRes(err)
	}
	// compile
	res, err := judger.Judge(
		judger.WithArgument("/dev/null", "/dev/null", outputs["log"].Path(),
			"/usr/bin/g++", src, "-o", outputs["result"].Path(), "-O2", "-Wall"),
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

var _ Processor = CompilerTestlib{}
