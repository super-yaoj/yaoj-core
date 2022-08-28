package main

import (
	"flag"
	"os"
	"path"

	"github.com/super-yaoj/yaoj-core/internal/app/judgeserver"
	"github.com/super-yaoj/yaoj-core/pkg/log"
)

var address string

func main() {
	flag.Parse()

	lg := log.NewTerminal()
	server := judgeserver.New(lg)
	dir := path.Join(os.TempDir(), "yaoj-judgeserver")
	err := judgeserver.Init(dir, lg)
	if err != nil {
		lg.Fatal(err)
	}

	err = server.Run(address) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	if err != nil {
		lg.Fatal(err)
	}
}

func init() {
	flag.StringVar(&address, "listen", "localhost:3000", "listening address")
}
