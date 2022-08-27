package judgeserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/super-yaoj/yaoj-core/internal/pkg/worker"
	"github.com/super-yaoj/yaoj-core/pkg/log"
)

type Server struct {
	*gin.Engine
	lg *log.Entry
}

type Context struct {
	*gin.Context
	lg *log.Entry
}

func (r *Server) Handle(name string, method string, handler func(ctx *Context) error) {
	r.Engine.Handle(method, name, func(ctx *gin.Context) {
		err := handler(&Context{
			Context: ctx,
			lg:      r.lg.WithField("route", name),
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
	}

	server.Handle("/judge", "POST", Judge)
	server.Handle("/custom", "POST", CustomTest)
	server.Handle("/sync", "POST", Sync)

	// handle signal
	// sigs := make(chan os.Signal, 1)
	// signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// go func() {
	// 	sig := <-sigs
	// 	server.lg.Printf("\nhandle signal %q\n", sig)

	// 	server.store.Map.Range(func(key, value any) bool {
	// 		prob := value.(*problem.Data)
	// 		prob.Finalize()
	// 		return true
	// 	})

	// 	fmt.Printf("done.\n")
	// 	os.Exit(0)
	// }()

	return server
}

var workerService *worker.Service

func Init(dir string) error {
	service, err := worker.New(dir, log.NewTerminal())
	if err != nil {
		return err
	}
	workerService = service
	return nil
}
