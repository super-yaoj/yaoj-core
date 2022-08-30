package problem_test

import (
	"bytes"
	"testing"

	"github.com/super-yaoj/yaoj-core/internal/pkg/processors"
	"github.com/super-yaoj/yaoj-core/pkg/problem"
	"github.com/super-yaoj/yaoj-core/pkg/workflow"
	utils "github.com/super-yaoj/yaoj-utils"
)

func TestSubmission(t *testing.T) {
	// create a submission
	subm := problem.Submission{}

	subm.SetData(workflow.Gsubm, "source", []byte("your source code"))
	subm.SetData(workflow.Gsubm, "option", (&processors.CompileConf{
		Lang: utils.Lcpp,
	}).Serialize())

	var buf bytes.Buffer
	err := subm.DumpTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	// LoadSubmData
	_, err = problem.LoadSubmData(buf.Bytes()[:])
	if err != nil {
		t.Fatal(err)
	}

	err = subm.SetReader(workflow.Gstatic, "trash", &buf)
	if err != nil {
		t.Fatal(err)
	}
	subm.Download(t.TempDir())
}
