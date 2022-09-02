package tests

import (
	"fmt"

	"github.com/super-yaoj/yaoj-core/pkg/data"
	"github.com/super-yaoj/yaoj-core/pkg/log"
	"github.com/super-yaoj/yaoj-core/pkg/problem"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
	"github.com/super-yaoj/yaoj-core/pkg/workflow/preset"
	utils "github.com/super-yaoj/yaoj-utils"
)

// testlib ncmp 源码
var NcmpSource = `
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

// a+b 问题满分源码c++
var APlusBSourceCpp = `
#include<bits/stdc++.h>
using namespace std;
int main() {
	int a, b;
	cin >> a >> b;
	cout << a + b << endl;
	return 0;
}
`

// 创建一个 a+b problem
//
// create dir if necessary
func CreateProblem(dir string, lg *log.Entry) (*problem.Data, error) {
	prob, err := problem.New(dir, lg)
	if err != nil {
		return nil, err
	}
	prob.Workflow = &preset.Traditional
	prob.Fullscore = 100
	prob.AnalyzerName = "traditional"

	// setup pretest
	prob.Pretest.InitTestcases()
	prob.Pretest.Fullscore = 100
	prob.Pretest.Method = problem.Msum
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

	// setup Data
	prob.Data.InitSubtasks()
	prob.Data.Fullscore = 100
	prob.Data.Method = problem.Msum
	subtaskdata := [][][2]int{{
		{1, 2},
		{3, 4},
		{5, 6},
	}, {
		{10, 2},
		{30, 4},
		{50, 6},
	}}
	for _, v := range subtaskdata {
		sub := prob.Data.NewSubtask(prob.Fullscore/float64(len(subtaskdata)), problem.Mmin)
		for _, v2 := range v {
			test := sub.NewTestcase()
			test.SetData("input", []byte(fmt.Sprint(v2[0], " ", v2[1])))
			test.SetData("output", []byte(fmt.Sprint(v2[0]+v2[1])))
		}
	}

	// setup static
	prob.Static.SetData("checker", []byte(NcmpSource))
	prob.Static.SetData("runner_config", (&data.RunConf{
		RealTime: 60 * 1000,
		CpuTime:  1000,
		RealMem:  512 * 1024 * 1024,
		StkMem:   512 * 1024 * 1024,
		Output:   64 * 1024 * 1024,
		Fileno:   5,
	}).Serialize())

	return prob, nil
}

// 创建 a+b problem 的满分提交
func CreateSubmission() problem.Submission {
	// setup submission
	submission := problem.Submission{}
	submission.SetData(workflow.Gsubm, "source", []byte(APlusBSourceCpp))
	submission.SetData(workflow.Gsubm, "option", (&data.CompileConf{
		Lang: utils.Lcpp11,
	}).Serialize())
	return submission
}
