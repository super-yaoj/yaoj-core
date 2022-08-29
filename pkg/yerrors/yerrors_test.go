package yerrors_test

import (
	"testing"

	"github.com/super-yaoj/yaoj-core/pkg/yerrors"
)

func TestAll(t *testing.T) {
	shiterr := yerrors.New("shit")
	err := yerrors.Annotated("testname", "TestAll", shiterr)
	err = yerrors.Situated("yerrors.New", err)
	err = yerrors.Annotated("alice", "bob", err)
	if !yerrors.Is(err, shiterr) {
		t.Fatal("invalid err", err)
	}
	t.Log(err)
}
