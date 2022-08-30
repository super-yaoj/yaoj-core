package judger

import (
	"github.com/super-yaoj/yaoj-core/pkg/yerrors"
)

var (
	ErrLogSet        = yerrors.New("log_set return non zero")
	ErrSetPolicy     = yerrors.New("set policy error")
	ErrSetRunner     = yerrors.New("set runner error")
	ErrUnknownRunner = yerrors.New("unknown runner")
	ErrRun           = yerrors.New("runner runtime error")
)
