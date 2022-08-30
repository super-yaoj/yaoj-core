package problemruntime

import "github.com/super-yaoj/yaoj-core/pkg/yerrors"

var (
	ErrInvalidSet      = yerrors.New("invalid test set")
	ErrUnknownAnalyzer = yerrors.New("unknown analyzer")
)
