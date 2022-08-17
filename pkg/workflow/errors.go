package workflow

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidGroupname    = errors.New("invalid inbound groupname")
	ErrInvalidEdge         = errors.New("invalid edge starting node or ending node")
	ErrInvalidInputLabel   = errors.New("invalid processor input label")
	ErrInvalidOutputLabel  = errors.New("invalid processor output label")
	ErrDuplicateDest       = errors.New("two edges have the same destination")
	ErrIncompleteNodeInput = errors.New("incomplete node input")
)

type Error struct {
	// what operation trigger this error
	Trgr string
	// the original error
	Err error
}

func (e *Error) Error() string {
	return fmt.Sprintf("workflow %s: %v", e.Trgr, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

type DataError struct {
	// attached data
	Data any
	// the original error
	Err error
}

func (e *DataError) Error() string {
	return fmt.Sprintf("workflow: %v (%#v)", e.Err, e.Data)
}

func (e *DataError) Unwrap() error {
	return e.Err
}

func ErrWithData(err error, data any) error {
	return &DataError{data, err}
}
