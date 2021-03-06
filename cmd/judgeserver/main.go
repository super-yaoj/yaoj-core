package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/super-yaoj/yaoj-core/pkg/buflog"
	"github.com/super-yaoj/yaoj-core/pkg/private/run"
	"github.com/super-yaoj/yaoj-core/pkg/problem"
)

// concurrent-safe storage
type Storage struct {
	Map sync.Map
}

func (r *Storage) Has(checksum string) bool {
	logger.Printf("has %s", checksum)
	_, ok := r.Map.Load(checksum)
	return ok
}
func (r *Storage) Set(checksum string, prob problem.Problem) {
	logger.Printf("set %s", checksum)
	r.Map.Store(checksum, prob)
}
func (r *Storage) Get(checksum string) problem.Problem {
	val, _ := r.Map.Load(checksum)
	return val.(problem.Problem)
}

var storage = Storage{Map: sync.Map{}}

var address string

func main() {
	flag.Parse()

	var cachedir = path.Join(os.TempDir(), "yaoj-judger-server-cache")
	os.RemoveAll(cachedir)

	if err := os.MkdirAll(cachedir, os.ModePerm); err != nil {
		logger.Fatal(err)
	}

	if err := run.CacheInit(cachedir); err != nil {
		logger.Fatal(err)
	}

	// handle signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Printf("\nhandle signal %q\n", sig)

		storage.Map.Range(func(key, value any) bool {
			prob := value.(problem.Problem)
			prob.Data().Finalize()
			return true
		})

		os.RemoveAll(cachedir)

		fmt.Printf("done.\n")
		os.Exit(0)
	}()

	r := gin.Default()
	r.POST("/judge", Judge)
	r.POST("/custom", CustomTest)
	r.POST("/sync", Sync)
	r.GET("/log", Log)

	err := r.Run(address) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	if err != nil {
		logger.Fatal(err)
	}
}

func init() {
	flag.StringVar(&address, "listen", "localhost:3000", "listening address")
}

var logger = buflog.New("[judgeserver] ")
