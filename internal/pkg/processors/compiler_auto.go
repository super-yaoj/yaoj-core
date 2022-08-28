package processors

import (
	"errors"
	"strings"
	"time"

	"github.com/bitfield/script"
	"github.com/super-yaoj/yaoj-core/internal/pkg/judger"
	"github.com/super-yaoj/yaoj-core/pkg/processor"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
)

var (
	ErrUnknownLang = errors.New("unknown language tag")
)

// Compile source file in all language.
//
// Time limitation: 1min.
//
// 各语言默认参数如下：
//
// Lc: gcc [source] -o [result]
//
// Lcpp: g++ [source] -o [result]
//
// Lcpp11: g++ [source] -o [result] --std=c++11
//
// Lcpp14: g++ [source] -o [result] --std=c++14
//
// Lcpp17: g++ [source] -o [result] --std=c++17
//
// Lcpp20: g++ [source] -o [result] --std=c++20
//
// Lpython, Lpython3: 用 cython 转化为 c 语言文件然后编译
type CompilerAuto struct {
	// input: source option
	// output: result, log, judgerlog
}

func (r CompilerAuto) Label() (inputlabel []string, outputlabel []string) {
	return []string{"source", "option"}, []string{"result", "log", "judgerlog"}
}

func (r CompilerAuto) Process(inputs Inbounds, outputs Outbounds) (result *Result) {
	var argv []string
	// parse compile option
	data, err := inputs["option"].Get()
	if err != nil {
		return SysErrRes(err)
	}
	var conf = &CompileConf{}
	err = conf.Deserialize(data)
	if err != nil {
		return SysErrRes(err)
	}

	basename := utils.RandomString(10)

	switch conf.Lang {
	case utils.Lc:
		inputs["source"].DupFile(basename+".c", 0644)
		argv = []string{
			"/dev/null", "/dev/null", outputs["log"].Path(),
			"/usr/bin/gcc", basename + ".c", "-o", outputs["result"].Path(),
		}
	case utils.Lcpp, utils.Lcpp11, utils.Lcpp14, utils.Lcpp17, utils.Lcpp20:
		inputs["source"].DupFile(basename+".cpp", 0644)
		// detect c++ version
		verArg := ""
		switch conf.Lang {
		case utils.Lcpp11:
			verArg = "--std=c++11"
		case utils.Lcpp14:
			verArg = "--std=c++14"
		case utils.Lcpp17:
			verArg = "--std=c++17"
		case utils.Lcpp20:
			verArg = "--std=c++2a"
		}

		// logger.Printf("auto compile source lang ver: %s", verArg)

		args := []string{
			"/dev/null", "/dev/null", outputs["log"].Path(),
			"/usr/bin/g++", basename + ".cpp", "-o", outputs["result"].Path(),
		}
		if verArg != "" {
			args = append(args, verArg)
		}
		argv = args
	case utils.Lpython, utils.Lpython3: // 目前只编译 python3
		// logger.Printf("detect python source")
		c_src := utils.RandomString(10) + ".c"
		py_src := utils.RandomString(10) + ".py"
		// compile source to c file
		utils.CopyFile(inputs["source"].Path(), py_src)
		res, err := judger.Judge(
			judger.WithPolicy("builtin:free"),
			judger.WithLog(outputs["judgerlog"].Path(), 0, false),
			judger.WithRealTime(time.Minute),
			judger.WithOutput(10*judger.MB),
			// 名字里含有 '-' cython 会报错
			judger.WithArgument("/dev/null", "/dev/null", outputs["log"].Path(),
				"/usr/bin/cython", py_src, "--embed", "-3", "-o", c_src),
		)
		if err != nil {
			return SysErrRes(err)
		}
		if res.Code != processor.Ok { // cython 编译出错
			// logger.Printf("cython compile error!")
			res := res.ProcResult()
			res.Msg = "cython compile: " + res.Msg
			return res
		}

		// compile c file using gcc
		CFLAGS, LDFLAGS, err := PythonFlags()
		if err != nil {
			return SysErrRes(err)
		}
		_, err = script.Exec(strings.Join([]string{
			"/usr/bin/gcc",
			"-Wall", "-Wextra", "-fpie",
			strings.TrimSpace(CFLAGS),
			"-o", outputs["result"].Path(), c_src,
			strings.TrimSpace(LDFLAGS),
		}, " ")).String()
		if err != nil {
			return RtErrRes(err)
		}
		return &processor.Result{
			Code: processor.Ok,
			Msg:  "",
		}
	default:
		return SysErrRes(ErrUnknownLang)
	}

	// compile other language
	argv = append(argv, conf.ExtraArgs...)
	res, err := judger.Judge(
		judger.WithArgument(argv...),
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

var _ Processor = CompilerAuto{}
