package buflog

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
)

var Writer = os.Stderr

// tailing logs
func Tail() []string {
	return tails[:]
}

var TailThreshold = 300
var tails []string
var tailmu sync.Mutex

type bufLogger struct {
	logger *log.Logger
}

func (r *bufLogger) appendTail(s string) {
	tailmu.Lock()
	defer tailmu.Unlock()

	tails = append(tails, s)
	if len(tails) > TailThreshold {
		tails = tails[1:]
	}
}

func (r *bufLogger) Errorf(format string, v ...any) error {
	r.logger.Printf("ERROR: "+format, v...)
	errstr := fmt.Sprintf(format, v...)
	r.appendTail("ERROR: " + errstr)
	return errors.New(errstr)
}

func (r *bufLogger) Printf(format string, v ...any) {
	r.logger.Printf(format, v...)
	r.appendTail(fmt.Sprintf(format, v...))
}
func (r *bufLogger) Print(v ...any) {
	r.logger.Print(v...)
	r.appendTail(fmt.Sprint(v...))
}
func (r *bufLogger) Fatal(v ...any) {
	r.Print(v...)
	os.Exit(1)
}

func New(prefix string) *bufLogger {
	logger := &bufLogger{}
	logger.logger = log.New(Writer, prefix, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lmsgprefix)
	return logger
}
