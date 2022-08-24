package workflowruntime_test

import (
	"path"
	"testing"

	"github.com/super-yaoj/yaoj-core/internal/pkg/processors"
	workflowruntime "github.com/super-yaoj/yaoj-core/internal/pkg/worker/workflow"
	"github.com/super-yaoj/yaoj-core/pkg/data"
	"github.com/super-yaoj/yaoj-core/pkg/log"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
	"github.com/super-yaoj/yaoj-core/pkg/workflow/preset"
	utils "github.com/super-yaoj/yaoj-utils"
)

var ncmp = `
#include "testlib.h"
#include <sstream>
using namespace std;
int main(int argc, char * argv[]) {
  setName("compare ordered sequences of signed int%d numbers", 8 * int(sizeof(long long)));
  registerTestlibCmd(argc, argv);
  int n = 0;
  string firstElems;
  while (!ans.seekEof() && !ouf.seekEof()) {
    n++;
    long long j = ans.readLong();
    long long p = ouf.readLong();
    if (j != p)
      quitf(_wa, "%d%s numbers differ - expected: '%s', found: '%s'", n, englishEnding(n).c_str(), vtos(j).c_str(), vtos(p).c_str());
    else
      if (n <= 5) {
        if (firstElems.length() > 0)
          firstElems += " ";
        firstElems += vtos(j);
      }
  }
  int extraInAnsCount = 0;
  while (!ans.seekEof()) {
    ans.readLong();
    extraInAnsCount++;
  }
  int extraInOufCount = 0;
  while (!ouf.seekEof()) {
    ouf.readLong();
    extraInOufCount++;
  }
  if (extraInAnsCount > 0)
    quitf(_wa, "Answer contains longer sequence [length = %d], but output contains %d elements", n + extraInAnsCount, n);
  if (extraInOufCount > 0)
    quitf(_wa, "Output contains longer sequence [length = %d], but answer contains %d elements", n + extraInOufCount, n);
  if (n <= 5)
    quitf(_ok, "%d number(s): \"%s\"", n, compress(firstElems).c_str());
  else
    quitf(_ok, "%d numbers", n);
}
`

var main = `
#include<bits/stdc++.h>
using namespace std;
int main() {
	int a, b;
	cin >> a >> b;
	cout << a + b << endl;
	return 0;
}
`

var input = `114 514`
var output = `628`

func TestRtWorkflow(t *testing.T) {
	dir := t.TempDir()

	inbounds := workflow.InboundGroups{
		workflow.Gstatic: make(map[string]data.FileStore),
		workflow.Gtests:  make(map[string]data.FileStore),
		workflow.Gsubm:   make(map[string]data.FileStore),
	}
	inbounds[workflow.Gstatic]["checker"] = data.NewFlex(path.Join(dir, "_chk.cpp"), []byte(ncmp))
	inbounds[workflow.Gstatic]["runner_config"] = data.NewFlex(path.Join(dir, "_runconf"), (&processors.RunConf{
		RealTime: 60 * 1000,
		CpuTime:  1000,
		RealMem:  512 * 1024 * 1024,
		StkMem:   512 * 1024 * 1024,
		Output:   64 * 1024 * 1024,
		Fileno:   5,
	}).Serialize())
	inbounds[workflow.Gsubm]["source"] = data.NewFlex(path.Join(dir, "_main.cpp"), []byte(main))
	inbounds[workflow.Gsubm]["option"] = data.NewFlex(path.Join(dir, "_cpl"), (&processors.CompileConf{
		Lang: utils.Lcpp11,
	}).Serialize())
	inbounds[workflow.Gtests]["input"] = data.NewFlex(path.Join(dir, "_input"), []byte(input))
	inbounds[workflow.Gtests]["output"] = data.NewFlex(path.Join(dir, "_output"), []byte(output))

	cache, err := workflowruntime.NewCache(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	wk, err := workflowruntime.New(&preset.Traditional, dir, 100, log.NewTerminal())
	if err != nil {
		t.Fatal(err)
	}
	wk.UseCache(cache)
	err = wk.Run(inbounds, false)
	if err != nil {
		t.Fatal(err)
	}

	wk2, err := workflowruntime.New(&preset.Traditional, dir, 100, log.NewTerminal())
	if err != nil {
		t.Fatal(err)
	}
	wk2.UseCache(cache)
	err = wk2.Run(inbounds, false)
	if err != nil {
		t.Fatal(err)
	}
}
