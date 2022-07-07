package processors

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/super-yaoj/yaoj-core/pkg/private/judger"
	"github.com/super-yaoj/yaoj-core/pkg/processor"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
)

// Compile source file in all language.
// Time limitation: 1min.
type CompilerAuto struct {
	// input: source
	// output: result, log, judgerlog
}

func (r CompilerAuto) Label() (inputlabel []string, outputlabel []string) {
	return []string{"source"}, []string{"result", "log", "judgerlog"}
}

func (r CompilerAuto) Run(input []string, output []string) *Result {
	ext := path.Ext(input[0])
	sub_ext := path.Ext(input[0][:len(input[0])-len(ext)])
	var arg judger.OptionProvider

	switch ext {
	case ".c":
		arg = judger.WithArgument(
			"/dev/null", "/dev/null", output[1], "/usr/bin/gcc", input[0], "-o", output[0],
			"-O2", "-lm", "-DONLINE_JUDGE",
		)
	case ".cpp", ".cc":
		// detect c++ version
		verArg := "--std=c++2a"
		switch utils.SourceLang(sub_ext) {
		case utils.Lcpp11:
			verArg = "--std=c++11"
		case utils.Lcpp14:
			verArg = "--std=c++14"
		case utils.Lcpp17:
			verArg = "--std=c++17"
		case utils.Lcpp20:
			verArg = "--std=c++2a"
		}

		logger.Printf("auto compile source lang ver: %s", verArg)

		arg = judger.WithArgument(
			"/dev/null", "/dev/null", output[1], "/usr/bin/g++", input[0], "-o", output[0],
			"-O2", "-lm", "-DONLINE_JUDGE", verArg,
		)
	case ".py", ".python":
		logger.Printf("detect python source")
		err := compilePy(input[0], output[0], utils.SourceLang(sub_ext))
		if err != nil {
			logger.Printf("compile error: %s", err)
			return &Result{
				Code: processor.RuntimeError,
				Msg:  err.Error(),
			}
		}
		return &processor.Result{
			Code: processor.Ok,
			Msg:  "ok",
		}
	default:
		return &Result{
			Code: processor.SystemError,
			Msg:  fmt.Sprintf("unknown source suffix %s", ext),
		}
	}

	res, err := judger.Judge(
		arg,
		judger.WithJudger(judger.General),
		judger.WithPolicy("builtin:free"),
		judger.WithLog(output[2], 0, false),
		judger.WithRealTime(time.Minute),
		judger.WithOutput(10*judger.MB),
	)
	if err != nil {
		return &Result{
			Code: processor.SystemError,
			Msg:  err.Error(),
		}
	}
	return res.ProcResult()
}

func compilePy(src, dest string, lang utils.LangTag) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer file.Close()

	if lang == utils.Lpython2 {
		file.WriteString("#!/bin/env python2\n\n")
	} else {
		file.WriteString("#!/bin/env python3\n\n")
	}
	file.Write(data)
	if err := file.Chmod(0744); err != nil {
		return err
	}
	return nil
}

var _ Processor = CompilerAuto{}
