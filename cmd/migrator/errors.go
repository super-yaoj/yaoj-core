package main

import (
	"errors"
	"fmt"
)

var (
	ErrUnknownType = errors.New("unknown type")
)

type Error struct {
	// what operation trigger this error
	Trgr string
	// the original error
	Err error
}

func (e *Error) Error() string {
	return fmt.Sprintf("worker.workflow %s: %v", e.Trgr, e.Err)
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
	return fmt.Sprintf("worker.workflow: %v (%#v)", e.Err, e.Data)
}

func (e *DataError) Unwrap() error {
	return e.Err
}

func ErrWithData(err error, data any) error {
	return &DataError{data, err}
}
