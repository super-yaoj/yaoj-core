package data_test

import (
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/super-yaoj/yaoj-core/pkg/data"
)

func TestFile(t *testing.T) {
	f := data.NewFile(path.Join(t.TempDir(), "File"), nil)
	f.Set([]byte("hello"))
	if ctnt, err := f.Get(); err != nil || string(ctnt) != "hello" {
		t.Fatal(ctnt, err)
	}
	if err := f.Set([]byte("#!/bin/bash\nls .")); err != nil {
		t.Fatal(err)
	}
	if err := f.ChangePath(path.Join(t.TempDir(), "File")); err != nil {
		t.Fatal(err)
	}
	f.SetMode(0744)
	cmd := exec.Command("bash", f.Path())
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	f.ChangePath(path.Join(t.TempDir(), "File_changed"))
	if !strings.Contains(f.Path(), "File_changed") {
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
	// test dupfile
	filename := path.Join(t.TempDir(), "file")
	if err := f.DupFile(filename, 0777); err != nil {
		t.Fatal(err)
	}
	f3 := data.NewFileFile(filename)
	f3data, _ := f3.Get()
	fdata, _ := f.Get()
	if string(f3data) != string(fdata) {
		t.Fatal("dup file error")
	}
}
