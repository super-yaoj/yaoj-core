package run

import "fmt"

type Error struct {
	// what operation trigger this error
	Trgr string
	// the original error
	Err error
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %v", e.Trgr, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}
