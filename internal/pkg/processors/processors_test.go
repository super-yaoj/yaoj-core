package processors_test

import (
	"os"
	"testing"

	"github.com/k0kubun/pp/v3"
	"github.com/super-yaoj/yaoj-core/internal/pkg/processors"
	"github.com/super-yaoj/yaoj-core/pkg/data"
	"github.com/super-yaoj/yaoj-core/pkg/processor"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
	yutils "github.com/super-yaoj/yaoj-utils"
)

// go:generate go build -buildmode=plugin -o ./testdata/diff-go ./testdata/diff-go/main.go
// func TestLoad(t *testing.T) {
// 	proc, err := processor.LoadPlugin("testdata/diff-go/main.so")
// 	if err != nil {
// 		t.Error(err)
// 	}
//
// 	t.Log(proc.Label())
// }

var c_src = `
#include<stdio.h>
int main() {
	freopen("a.in", "r", stdin);
	freopen("a.out", "w", stdout);
	int a, b;
	scanf("%d%d", &a, &b);
	printf("%d\n", a + b);
	return 0;
}
`

var cpp_src = `
#include<iostream>
using namespace std;
int main() {
	int a, b;
	cin >> a >> b;
	cout << a + b << endl;
	return 0;
}
`

var py_src = `
a,b=map(int,input().split())
print(a+b)
`

var checker_yesno_src = `
#include "testlib.h"
#include <string>

using namespace std;

const string YES = "YES";
const string NO = "NO";

int main(int argc, char * argv[]) {
  setName("%s", (YES + " or " + NO + " (case insensetive)").c_str());
  registerTestlibCmd(argc, argv);
  std::string ja = upperCase(ans.readWord());
  std::string pa = upperCase(ouf.readWord());
  if (ja != YES && ja != NO)
      quitf(_fail, "%s or %s expected in answer, but %s found", YES.c_str(), NO.c_str(), compress(ja).c_str());
  if (pa != YES && pa != NO)
      quitf(_pe, "%s or %s expected, but %s found", YES.c_str(), NO.c_str(), compress(pa).c_str());
  if (ja != pa)
      quitf(_wa, "expected %s, found %s", compress(ja).c_str(), compress(pa).c_str());
  quitf(_ok, "answer is %s", ja.c_str());
}
`

func TestProcessors(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// change working dir
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	// change back after testing
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("PythonFlags", func(t *testing.T) {
		cflags, ldflags, err := processors.PythonFlags()
		if err != nil {
			t.Fatal(err)
		}
		t.Log("cflags: ", cflags)
		t.Log("ldflags: ", ldflags)
	})

	t.Run("CompilerAuto", func(t *testing.T) {
		var testcases = []struct {
			name string
			src  string
			exec string
			lang utils.LangTag
		}{
			{"c", c_src, "exec_c", yutils.Lc},
			{"cpp", cpp_src, "exec_cpp", yutils.Lcpp},
			{"python", py_src, "exec_py", yutils.Lpython},
		}

		for _, testcase := range testcases {
			t.Run(testcase.name, func(t *testing.T) {
				// source
				src := data.NewFile("main.txt", nil)
				src.Set([]byte(testcase.src))
				// option
				conf := processors.CompileConf{Lang: testcase.lang}

				inputs := processor.Inbounds{
					"source": src,
					"option": data.NewFile("tmp", conf.Serialize()),
				}
				outputs := processor.Outbounds{
					"result":    data.NewFile(testcase.exec, nil),
					"log":       data.NewFile("main.log", nil),
					"judgerlog": data.NewFile("runtime.log", nil),
				}
				res := processors.CompilerAuto{}.Process(inputs, outputs)
				if res.Code != processor.Ok {
					data_log, _ := outputs["log"].Get()
					t.Logf("log: %s", string(data_log))
					data_runtime, _ := outputs["judgerlog"].Get()
					t.Logf("runtime.log: %s", string(data_runtime))
					t.Fatal("invalid result", pp.Sprint(res))
					return
				}
			})
		}

	})

	t.Run("RunnerAuto", func(t *testing.T) {
		err := os.WriteFile("exec.in", []byte("1 2"), 0644)
		if err != nil {
			t.Fatal(err)
		}

		var testcases = []struct {
			name  string
			exec  string
			input string
			conf  processors.RunConf
		}{
			{"fileio", "exec_c", "exec.in", processors.RunConf{
				RealTime: 5 * 1000,
				CpuTime:  5 * 1000,
				VirMem:   512 * 1000 * 1000,
				RealMem:  512 * 1000 * 1000,
				StkMem:   512 * 1000 * 1000,
				Output:   64 * 1000 * 1000,
				Fileno:   10,
				Inf:      "a.in",
				Ouf:      "a.out",
			}},
			{"stdio", "exec_cpp", "exec.in", processors.RunConf{
				RealTime: 5 * 1000,
			}},
		}

		for _, testcase := range testcases {
			t.Run(testcase.name, func(t *testing.T) {
				inputs := processor.Inbounds{
					"executable": data.NewFileFile(testcase.exec),
					"stdin":      data.NewFileFile(testcase.input),
					"conf":       data.NewFile("tmp", testcase.conf.Serialize()),
				}
				outputs := processor.Outbounds{
					"stdout":    data.NewFile("exec.out", nil),
					"stderr":    data.NewFile("exec.err", nil),
					"judgerlog": data.NewFile("runtime.log", nil),
				}
				res := processors.RunnerAuto{}.Process(inputs, outputs)
				if res.Code != processor.Ok {
					data_runtime, _ := outputs["judgerlog"].Get()
					t.Logf("runtime.log: %s", string(data_runtime))
					t.Fatal("invalid result", pp.Sprint(res))
					return
				}
				data_stdout, _ := outputs["stdout"].Get()
				t.Log("stdout:", string(data_stdout))
			})
		}
	})
	t.Run("CompilerTestlib", func(t *testing.T) {
		inputs := processor.Inbounds{
			"source": data.NewFile("tmp", []byte(checker_yesno_src)),
		}
		outputs := processor.Outbounds{
			"result":    data.NewFile("exec_checker", nil),
			"log":       data.NewFile("checker.log", nil),
			"judgerlog": data.NewFile("runtime.log", nil),
		}

		res := processors.CompilerTestlib{}.Process(inputs, outputs)
		if res.Code != processor.Ok {
			t.Fatalf("expect %v, found %v Msg=%s", processor.Ok, res.Code, res.Msg)
		}
		t.Log(res)
	})
	t.Run("CheckerTestlib", func(t *testing.T) {
		inputs := processor.Inbounds{
			"checker": data.NewFileFile("exec_checker"),
			"input":   data.NewFile("input", []byte("yes")),
			"output":  data.NewFile("output", []byte("yes")),
			"answer":  data.NewFile("answer", []byte("yes")),
		}
		outputs := processor.Outbounds{
			"xmlreport": data.NewFile("report.xml", nil),
			"stderr":    data.NewFile("checker.err", nil),
			"judgerlog": data.NewFile("runtime.log", nil),
		}
		res := processors.CheckerTestlib{}.Process(inputs, outputs)
		if res.Code != processor.Ok {
			t.Fatalf("expect %v, found %v Msg=%s", processor.Ok, res.Code, res.Msg)
		}
		t.Log(res)
	})
}

func TestManager(t *testing.T) {
	mp := processors.GetAll()
	for k := range mp {
		input, output := processors.Get(k).Label()
		t.Log(k, input, output)
	}
}
