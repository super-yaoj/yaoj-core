package judger_test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/super-yaoj/yaoj-core/internal/pkg/judger"
	"github.com/super-yaoj/yaoj-core/pkg/processor"
)

func TestJudge(t *testing.T) {
	dir := t.TempDir()
	res, err := judger.Judge(
		judger.WithArgument("/dev/null", path.Join(dir, "output"), "/dev/null", "/usr/bin/ls", "."),
		judger.WithJudger(judger.General),
		judger.WithPolicy("builtin:free"),
		judger.WithLog(path.Join(dir, "runtime.log"), 0),
		judger.WithRealMemory(300*judger.MB),
		judger.WithStack(300*judger.MB),
		judger.WithVirMemory(300*judger.MB),
		judger.WithRealTime(time.Millisecond*1000),
		judger.WithCpuTime(time.Millisecond*1000),
		judger.WithOutput(30*judger.MB),
		judger.WithFileno(10),
		judger.WithEnviron(os.Environ()...),
	)
	if err != nil {
		t.Error(err)
		return
	}
	if res.Code != processor.Ok {
		t.Fatal("invalid result", res)
	}
	t.Log(*res, res.ProcResult())
}
