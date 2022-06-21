package errors

import "fmt"

type FatalError struct {
	msg string
}

func (e *FatalError) Error() string {
	return e.msg
}

func NewFatalError(msg string, vars ...interface{}) error {
	return &FatalError{
		msg: fmt.Sprintf(msg, vars...),
	}
}
