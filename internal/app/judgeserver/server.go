package judgeserver

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/super-yaoj/yaoj-core/pkg/log"
	"github.com/super-yaoj/yaoj-core/pkg/problem"
)

type Server struct {
	*gin.Engine
	lg    *log.Entry
	store *Storage
}

type Context struct {
	*gin.Context
	lg    *log.Entry
	store *Storage
}

func (r *Context) RespondError(err error) {
	r.Context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func (r *Server) Handle(name string, method string, handler func(ctx *Context) error) {
	r.Engine.Handle(method, name, func(ctx *gin.Context) {
		err := handler(&Context{
			Context: ctx,
			lg:      r.lg.WithField("route", name),
			store:   r.store,
		})
		if err != nil {
			r.lg.Error(err)
			if httperr, ok := err.(*HttpError); ok {
				ctx.JSON(httperr.status, gin.H{"error": httperr.Error()})
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
		}
	})
}

func New() *Server {
	server := &Server{
		Engine: gin.Default(),
		lg:     log.NewTerminal(),
		store:  &Storage{Map: sync.Map{}},
	}

	server.Handle("/judge", "POST", Judge)
	server.Handle("/custom", "POST", CustomTest)
	server.Handle("/sync", "POST", Sync)

	// handle signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		server.lg.Printf("\nhandle signal %q\n", sig)

		server.store.Map.Range(func(key, value any) bool {
			prob := value.(*problem.Data)
			prob.Finalize()
			return true
		})

		fmt.Printf("done.\n")
		os.Exit(0)
	}()

	return server
}
