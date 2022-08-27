package problemruntime_test

import (
	"testing"

	problemruntime "github.com/super-yaoj/yaoj-core/internal/pkg/worker/problem"
	"github.com/super-yaoj/yaoj-core/internal/tests"
	"github.com/super-yaoj/yaoj-core/pkg/log"
)

func TestRtProblem(t *testing.T) {
	prob, err := tests.CreateProblem(t.TempDir(), log.NewTerminal())
	if err != nil {
		t.Fatal(err)
	}

	// setup submission
	submission := tests.CreateSubmission()

	// init RtProblem
	rtdir := t.TempDir()
	rtprob, err := problemruntime.New(prob, rtdir, log.NewTerminal())
	if err != nil {
		t.Fatal(err)
	}
	res, err := rtprob.RunTestset(rtprob.Pretest, submission)
	if err != nil {
		t.Fatal(err)
	}
	if res.Score != res.Fullscore {
		t.Fatal("invalid result", res)
	}
	res, err = rtprob.RunTestset(rtprob.Data.Data, submission)
	if err != nil {
		t.Fatal(err)
	}
	if res.Score != res.Fullscore {
		t.Fatal("invalid result", res)
	}
}
