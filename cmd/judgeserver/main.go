package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/super-yaoj/yaoj-core/internal/app/judgeserver"
)

var address string

func main() {
	flag.Parse()

	server := judgeserver.New()
	dir := path.Join(os.TempDir(), "yaoj-judgeserver")
	err := judgeserver.Init(dir)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	err = server.Run(address) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}

func init() {
	flag.StringVar(&address, "listen", "localhost:3000", "listening address")
}
