package workflowruntime

import "github.com/super-yaoj/yaoj-core/pkg/yerrors"

var (
	ErrNilInboundGroup = yerrors.New("invalid nil inboundgroup")
	ErrIncompleteInput = yerrors.New("incomplete node input")
)
