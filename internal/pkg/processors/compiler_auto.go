package processors

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/bitfield/script"
	"github.com/k0kubun/pp/v3"
	"github.com/super-yaoj/yaoj-core/internal/pkg/judger"
	"github.com/super-yaoj/yaoj-core/pkg/processor"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
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

	switch utils.SourceLang(ext) {
	case utils.Lc:
		arg = judger.WithArgument(
			"/dev/null", "/dev/null", output[1], "/usr/bin/gcc", input[0], "-o", output[0],
			"-O2", "-lm", "-DONLINE_JUDGE",
		)
	case utils.Lcpp:
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
	case utils.Lpython: // 目前只编译 python3
		logger.Printf("detect python source")
		c_src := utils.RandomString(10) + ".c"
		py_src := utils.RandomString(10) + ".py"
		utils.CopyFile(input[0], py_src)
		res, err := judger.Judge(
			judger.WithPolicy("builtin:free"),
			judger.WithLog(output[2], 0, false),
			judger.WithRealTime(time.Minute),
			judger.WithOutput(10*judger.MB),
			// 名字里含有 '-' cython 会报错
			judger.WithArgument("/dev/null", "/dev/null", output[1], "/usr/bin/cython", py_src, "--embed", "-3", "-o", c_src),
		)
		if err != nil {
			return SysErrRes(err)
		}
		if res.Code != judger.Ok { // cython 编译出错
			logger.Printf("cython compile error!")
			pp.Print(res)
			pp.Print(workflow.FileDisplay(output[1], "compile log", 1000))
			return res.ProcResult()
		}

		CFLAGS, _ := script.Exec("python3-config --includes").String()

		LDFLAGS, _ := script.Exec("python3-config --ldflags").String()
		PY_VER, _ := script.Exec(`python3 -c 'import sys; print(".".join(map(str, sys.version_info[:2])))'`).String()
		/*arg = judger.WithArgument(
			"/dev/null", "/dev/null", output[1],
			"/usr/bin/gcc", strings.TrimSpace(CFLAGS),
			//"-Os",
			strings.TrimSpace(LDFLAGS), strings.TrimSpace("-lpython"+PY_VER),
			c_src, "-o", output[0],
			//"-DONLINE_JUDGE",
		)*/
		out, err := script.Exec(strings.Join([]string{
			"/usr/bin/gcc",
			strings.TrimSpace(CFLAGS),
			strings.TrimSpace(LDFLAGS),
			strings.TrimSpace("-lpython" + PY_VER),
			c_src,
			"-o",
			output[0],
			"-Os",
		}, " ")).String()
		if err != nil {
			return RtErrRes(err)
		}
		pp.Print(out)
		return &processor.Result{
			Code: processor.Ok,
			Msg:  "",
		}
	default:
		return SysErrRes(fmt.Errorf("unknown source suffix %s", ext))
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
		return SysErrRes(err)
	}
	pp.Print(res)
	pp.Print(workflow.FileDisplay(output[1], "compile log", 1000))
	log.Print(os.Environ())
	return res.ProcResult()
}

// for invalid lang tag, python3 is used
/*func compilePy(src, dest string, lang utils.LangTag) error {
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
}*/

var _ Processor = CompilerAuto{}
