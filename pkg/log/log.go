// log utilties
package log

import (
	"context"

	"github.com/sirupsen/logrus"
)

// context key
type key int

// context log entry key
const Kentry key = 0

func MustCtxLogger(ctx context.Context) (logger *logrus.Entry) {
	logger = ctx.Value(Kentry).(*logrus.Entry)
	if logger == nil {
		panic("context without logger")
	}
	return
}

func CtxWithLog(ctx context.Context, logger *logrus.Entry) context.Context {
	return context.WithValue(ctx, Kentry, logger)
}
