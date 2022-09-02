package workflowruntime_test

import (
	"path"
	"testing"

	"github.com/super-yaoj/yaoj-core/internal/pkg/analyzers"
	workflowruntime "github.com/super-yaoj/yaoj-core/internal/pkg/worker/workflow"
	"github.com/super-yaoj/yaoj-core/internal/tests"
	"github.com/super-yaoj/yaoj-core/pkg/data"
	"github.com/super-yaoj/yaoj-core/pkg/log"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
	"github.com/super-yaoj/yaoj-core/pkg/workflow/preset"
	utils "github.com/super-yaoj/yaoj-utils"
)

var input = `114 514`
var output = `628`

func TestRtWorkflow(t *testing.T) {
	lg := log.NewTest()
	dir := t.TempDir()
	inbounds := workflow.InboundGroups{
		workflow.Gstatic: make(map[string]data.FileStore),
		workflow.Gtests:  make(map[string]data.FileStore),
		workflow.Gsubm:   make(map[string]data.FileStore),
	}
	inbounds[workflow.Gsubm]["source"] = data.NewFile(path.Join(dir, "_main.cpp"), []byte(tests.APlusBSourceCpp))
	inbounds[workflow.Gsubm]["option"] = data.NewFile(path.Join(dir, "_cpl"), (&data.CompileConf{
		Lang: utils.Lcpp11,
	}).Serialize())

	inbounds[workflow.Gstatic]["checker"] = data.NewFile(path.Join(dir, "_chk.cpp"), []byte(tests.NcmpSource))
	inbounds[workflow.Gstatic]["runner_config"] = data.NewFile(path.Join(dir, "_runconf"), (&data.RunConf{
		RealTime: 60 * 1000,
		CpuTime:  1000,
		RealMem:  512 * 1024 * 1024,
		StkMem:   512 * 1024 * 1024,
		Output:   64 * 1024 * 1024,
		Fileno:   5,
	}).Serialize())
	inbounds[workflow.Gtests]["input"] = data.NewFile(path.Join(dir, "_input"), []byte(input))
	inbounds[workflow.Gtests]["output"] = data.NewFile(path.Join(dir, "_output"), []byte(output))

	cache, err := workflowruntime.NewCache(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	wk, err := workflowruntime.New(&preset.Traditional, t.TempDir(), 100, analyzers.Traditional{}, lg)
	if err != nil {
		t.Fatal(err)
	}
	wk.UseCache(cache)
	res, err := wk.Run(inbounds, false)
	if err != nil {
		t.Fatal(err)
	}
	if res.Title != "Accepted" {
		t.Fatal("invalid result", res)
	}
	wk.Finalize()

	wk2, err := workflowruntime.New(&preset.Traditional, t.TempDir(), 100, analyzers.Traditional{}, lg)
	if err != nil {
		t.Fatal(err)
	}
	wk2.UseCache(cache)
	_, err = wk2.Run(inbounds, false)
	if err != nil {
		t.Fatal(err)
	}
}
