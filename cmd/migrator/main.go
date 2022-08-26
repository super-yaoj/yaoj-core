package main

import (
	"flag"
	"os"

	"github.com/super-yaoj/yaoj-core/internal/app/migrator"
	"github.com/super-yaoj/yaoj-core/pkg/log"
	"github.com/super-yaoj/yaoj-core/pkg/utils"
)

var isUoj bool
var srcDir string
var destFile string
var lg = log.NewTerminal()

func Main() error {
	var mig migrator.Migrator
	if isUoj {
		mig = migrator.NewUojTraditional(srcDir, lg)
	} else {
		return ErrUnknownType
	}

	dir, err := os.MkdirTemp(os.TempDir(), "yaoj-migrator-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	err = mig.Migrate(destFile)
	if err != nil {
		return err
	}
	chk := utils.FileChecksum(destFile)
	lg.Infof("checksum: %s\n", chk.String())

	lg.Infof("done.")
	return nil
}

func main() {
	flag.Parse()

	err := Main()
	if err != nil {
		lg.WithError(err).Error("some things was wrong")
		flag.Usage()
		return
	}
}

func init() {
	flag.StringVar(&srcDir, "src", "", "source directory")
	flag.StringVar(&destFile, "dump", "", "output a zip archive with given name")
	flag.BoolVar(&isUoj, "uoj", false, "migrate from uoj problem data")
}
