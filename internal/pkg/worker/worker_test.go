package worker_test

import (
	"fmt"
	"path"
	"path/filepath"
	"testing"

	"github.com/k0kubun/pp/v3"
	"github.com/super-yaoj/yaoj-core/internal/pkg/processors"
	problemruntime "github.com/super-yaoj/yaoj-core/internal/pkg/worker/problem"
	workflowruntime "github.com/super-yaoj/yaoj-core/internal/pkg/worker/workflow"
	"github.com/super-yaoj/yaoj-core/pkg/data"
	"github.com/super-yaoj/yaoj-core/pkg/log"
	"github.com/super-yaoj/yaoj-core/pkg/problem"
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

func TestWorker(t *testing.T) {
	t.Run("RtWorkflow", func(t *testing.T) {
		dir := t.TempDir()
		inbounds := workflow.InboundGroups{
			workflow.Gstatic: make(map[string]data.FileStore),
			workflow.Gtests:  make(map[string]data.FileStore),
			workflow.Gsubm:   make(map[string]data.FileStore),
		}
		inbounds[workflow.Gsubm]["source"] = data.NewFile(path.Join(dir, "_main.cpp"), []byte(main))
		inbounds[workflow.Gsubm]["option"] = data.NewFile(path.Join(dir, "_cpl"), (&processors.CompileConf{
			Lang: utils.Lcpp11,
		}).Serialize())

		inbounds[workflow.Gstatic]["checker"] = data.NewFile(path.Join(dir, "_chk.cpp"), []byte(ncmp))
		inbounds[workflow.Gstatic]["runner_config"] = data.NewFile(path.Join(dir, "_runconf"), (&processors.RunConf{
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

		wk, err := workflowruntime.New(&preset.Traditional, dir, 100, workflowruntime.DefaultAnalyzer{}, log.NewTerminal())
		if err != nil {
			t.Fatal(err)
		}
		wk.UseCache(cache)
		_, err = wk.Run(inbounds, false)
		if err != nil {
			t.Fatal(err)
		}

		wk2, err := workflowruntime.New(&preset.Traditional, dir, 100, workflowruntime.DefaultAnalyzer{}, log.NewTerminal())
		if err != nil {
			t.Fatal(err)
		}
		wk2.UseCache(cache)
		_, err = wk2.Run(inbounds, false)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("RtProblem", func(t *testing.T) {
		prob, err := problem.New(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		prob.Workflow = &preset.Traditional
		prob.Fullscore = 100

		// setup pretest
		prob.Pretest.InitTestcases()
		testdata := [][2]int{
			{1, 2},
			{3, 4},
			{5, 6},
		}
		for _, v := range testdata {
			tc := prob.Pretest.NewTestcase()
			tc.SetData("input", []byte(fmt.Sprint(v[0], " ", v[1])))
			tc.SetData("output", []byte(fmt.Sprint(v[0]+v[1])))
		}

		// setup static
		prob.Static.SetData("checker", []byte(ncmp))
		prob.Static.SetData("runner_config", (&processors.RunConf{
			RealTime: 60 * 1000,
			CpuTime:  1000,
			RealMem:  512 * 1024 * 1024,
			StkMem:   512 * 1024 * 1024,
			Output:   64 * 1024 * 1024,
			Fileno:   5,
		}).Serialize())

		// setup submission
		submission := problem.Submission{}
		submission.SetData(workflow.Gsubm, "source", []byte(main))
		submission.SetData(workflow.Gsubm, "option", (&processors.CompileConf{
			Lang: utils.Lcpp11,
		}).Serialize())

		// init RtProblem
		rtdir, _ := filepath.Abs("local")
		rtprob, err := problemruntime.New(prob, rtdir, log.NewTerminal())
		if err != nil {
			t.Fatal(err)
		}
		res, err := rtprob.RunTestset(rtprob.Pretest, submission)
		if err != nil {
			t.Fatal(err)
		}
		pp.Println(res)
	})
}
