package workflow

import (
	"github.com/super-yaoj/yaoj-core/pkg/yerrors"
)

var (
	ErrInvalidGroupname    = yerrors.New("invalid inbound groupname")
	ErrInvalidEdge         = yerrors.New("invalid edge starting node or ending node")
	ErrInvalidInputLabel   = yerrors.New("invalid processor input label")
	ErrInvalidOutputLabel  = yerrors.New("invalid processor output label")
	ErrDuplicateDest       = yerrors.New("two edges have the same destination")
	ErrIncompleteNodeInput = yerrors.New("incomplete node input")
)
