package buflog

import (
	"bytes"
	"log"
	"os"
)

var Writer = os.Stderr

// tailing logs
func Tail() []string {
	return tails[:]
}

var TailThreshold = 300
var tails []string

type bufLogger struct {
	buf    *bytes.Buffer
	logger *log.Logger
}

func (r *bufLogger) Printf(fmt string, v ...any) {
	r.logger.Printf(fmt, v...)

	tails = append(tails, r.buf.String())
	if len(tails) > TailThreshold {
		tails = tails[1:]
	}

	r.buf.WriteTo(Writer)
}
func (r *bufLogger) Print(v ...any) {
	r.logger.Print(v...)

	tails = append(tails, r.buf.String())
	if len(tails) > TailThreshold {
		tails = tails[1:]
	}

	r.buf.WriteTo(Writer)
}
func (r *bufLogger) Fatal(v ...any) {
	r.Print(v...)
	os.Exit(1)
}

func New(prefix string) *bufLogger {
	logger := &bufLogger{}
	logger.buf = &bytes.Buffer{}
	logger.logger = log.New(logger.buf, prefix, log.LstdFlags|log.Lshortfile|log.Lmsgprefix)
	return logger
}
