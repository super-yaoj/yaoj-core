package processors

import (
	"os/exec"
	"strings"
	"time"

	"github.com/super-yaoj/yaoj-core/internal/pkg/judger"
	"github.com/super-yaoj/yaoj-core/pkg/data"
	"github.com/super-yaoj/yaoj-core/pkg/processor"
)

// `s` contains a series of number seperated by space, denoting
// real time (ms), cpu time (ms), virtual memory (byte), real memory (byte),
// stack memory (byte), output limit (byte), fileno limitation respectively.
func runLimOptions(s data.RunConf) []judger.OptionProvider {
	var rt, ct, vm, rm, sm, ol, fl uint = s.RealTime, s.CpuTime, s.VirMem, s.RealMem, s.StkMem, s.Output, s.Fileno
	options := []judger.OptionProvider{}
	if rt > 0 {
		options = append(options, judger.WithRealTime(time.Millisecond*time.Duration(rt)))
	}
	if ct > 0 {
		options = append(options, judger.WithCpuTime(time.Millisecond*time.Duration(ct)))
	}
	if vm > 0 {
		options = append(options, judger.WithVirMemory(judger.ByteValue(vm)))
	}
	if rm > 0 {
		options = append(options, judger.WithRealMemory(judger.ByteValue(rm)))
	}
	if sm > 0 {
		options = append(options, judger.WithStack(judger.ByteValue(sm)))
	}
	if ol > 0 {
		options = append(options, judger.WithOutput(judger.ByteValue(ol)))
	}
	if fl > 0 {
		options = append(options, judger.WithFileno(int(fl)))
	}
	return options
}

func RtErrRes(err error) *Result {
	return &Result{
		Code: processor.RuntimeError,
		Msg:  err.Error(),
	}
}

func SysErrRes(err error) *Result {
	return &Result{
		Code: processor.SystemError,
		Msg:  err.Error(),
	}
}

func PythonFlags() (cflags string, ldflags string, err error) {
	var cflagsbuf, ldflagsbuf strings.Builder
	// cflags
	cmdcflags := exec.Command("python3-config", "--cflags", "--embed")
	cmdcflags.Stdout = &cflagsbuf
	err = cmdcflags.Run()
	if err != nil {
		return "", "", err
	}
	cflags = cflagsbuf.String()
	// ldflags
	cmdldflags := exec.Command("python3-config", "--ldflags", "--embed")
	cmdldflags.Stdout = &ldflagsbuf
	err = cmdldflags.Run()
	if err != nil {
		return "", "", err
	}
	ldflags = ldflagsbuf.String()

	return
}
