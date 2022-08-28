package judgeserver

import (
	"fmt"
)

var ()

type Error struct {
	// what operation trigger this error
	Trgr string
	// the original error
	Err error
}

func (e *Error) Error() string {
	return fmt.Sprintf("judgeserver %s: %v", e.Trgr, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

type HttpError struct {
	status int
	Err    error
}

func (e *HttpError) Error() string {
	return fmt.Sprintf("judgeserver: %v", e.Err)
}
func (e *HttpError) Unwrap() error {
	return e.Err
}
