package judgeserver

import (
	"fmt"
)

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
