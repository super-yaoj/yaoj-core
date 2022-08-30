package worker

import "github.com/super-yaoj/yaoj-core/pkg/yerrors"

var (
	ErrInvalidChecksum = yerrors.New("invalid checksum synchornizing data")
	ErrNoSuchProblem   = yerrors.New("no such problem")
)
