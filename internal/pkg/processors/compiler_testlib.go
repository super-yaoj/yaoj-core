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

func (r CompilerTestlib) Run(input []string, output []string) *Result {
	file, err := os.Create("testlib.h")
	if err != nil {
		return RtErrRes(err)
	}
	_, err = file.Write(testlib)
	if err != nil {
		return RtErrRes(err)
	}
	file.Close()

	src := utils.RandomString(10) + ".cpp"
	if _, err := utils.CopyFile(input[0], src); err != nil {
		return RtErrRes(err)
	}
	res, err := judger.Judge(
		judger.WithArgument("/dev/null", "/dev/null", output[1], "/usr/bin/g++", src, "-o", output[0], "-O2", "-Wall"),
		judger.WithJudger(judger.General),
		judger.WithPolicy("builtin:free"),
		judger.WithLog(output[2], 0, false),
		judger.WithRealTime(time.Minute),
		judger.WithOutput(10*judger.MB),
	)
	if err != nil {
		return SysErrRes(err)
	}
	return res.ProcResult()
}

var _ Processor = CompilerTestlib{}
