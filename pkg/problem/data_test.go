package problem_test

import (
	"path"
	"testing"

	"github.com/super-yaoj/yaoj-core/pkg/log"
	"github.com/super-yaoj/yaoj-core/pkg/problem"
)

func TestData(t *testing.T) {
	stmt_str := "读入两个整数 a, b, 请输出 a + b 的值。"
	pdir := t.TempDir()
	// new problem
	prob, err := problem.New(pdir, log.NewTest())
	if err != nil {
		t.Fatal(err)
	}

	// bad new
	_, err = problem.New("", nil)
	if err == nil {
		t.Fatal("invalid err")
	}
	t.Log(err)

	// set data
	prob.Statement.Range(func(field, name string) {
		t.Fatalf("unknown field: %s, %s", field, name)
	})
	err = prob.Statement.SetData("zh", []byte(stmt_str))
	if err != nil {
		t.Fatal(err)
	}
	prob.Pretest.InitTestcases()
	testcase_1 := prob.Pretest.NewTestcase()
	testcase_1.SetData("input", []byte("1 2"))
	testcase_1.SetData("output", []byte("3"))

	prob.Data.InitSubtasks()
	sbt_1 := prob.Data.NewSubtask(100, problem.Mmin)
	ststc_1 := sbt_1.NewTestcase()
	ststc_1.SetData("input", []byte("3 4"))
	ststc_1.SetData("output", []byte("7"))

	// dump file
	dst := path.Join(t.TempDir(), "lcoal.zip")
	err = prob.DumpFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	// bad dump
	err = prob.DumpFile("")
	if err == nil {
		t.Fatal("invalid err")
	}
	t.Log(err)

	// finalize
	defer prob.Finalize()

	// hackable
	t.Log(prob.Hackable())

	// load file
	prob2, err := problem.LoadFileTo(dst, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	stmt, err := prob2.Statement.GetData("zh")
	if err != nil {
		t.Fatal(err)
	}
	if stmt_str != string(stmt) {
		t.Fatalf("statement changed: %s", string(stmt))
	}
	prob2.Statement.Range(func(field, name string) {
		t.Logf("range: %s, %s", field, name)
	})
	prob2.Statement.Delete("zh")
	prob2.Statement.Range(func(field, name string) {
		t.Fatalf("unknown field: %s, %s", field, name)
	})
	output_1, err := prob2.Data.Subtasks[0].Testcases[0].GetData("output")
	if err != nil {
		t.Fatal(err)
	}
	if string(output_1) != "7" {
		t.Fatalf("output changed: %s", string(stmt))
	}
	// bad load
	_, err = problem.LoadFileTo("", t.TempDir())
	if err == nil {
		t.Fatal("invalid err", err)
	}
	t.Log(err)
	// pp.Println(prob2)
}
