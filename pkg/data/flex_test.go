package data_test

import (
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/super-yaoj/yaoj-core/pkg/data"
)

func TestFlex(t *testing.T) {
	f := data.FlexWithPath("")
	// content mode
	f.Set([]byte("hello"))
	if ctnt, err := f.Get(); err != nil || string(ctnt) != "hello" {
		t.Fatal(ctnt, err)
	}
	if err := f.Set([]byte("#!/bin/bash\nls .")); err != nil {
		t.Fatal(err)
	}
	if err := f.ChangePath(path.Join(t.TempDir(), "flex")); err != nil {
		t.Fatal(err)
	}
	// file mode
	f.SetMode(0744)
	cmd := exec.Command("bash", f.Path())
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	f.ChangePath(path.Join(t.TempDir(), "flex_changed"))
	if !strings.Contains(f.Path(), "flex_changed") {
		t.Fatal("change path failed", f.Path())
	}
	if err := f.Set([]byte("#!/bin/bash\necho hello")); err != nil {
		t.Fatal(err)
	}
	cmd = exec.Command(f.Path())
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Get(); err != nil {
		t.Fatal(err)
	}
}
