package errors

import "fmt"

type FatalError struct {
	msg string
}

func (e FatalError) Error() string {
	return e.msg
}

func NewFatalError(msg string) FatalError {
	return FatalError{
		msg: msg,
	}
}

func NewFatalErrorf(msg string, vars ...interface{}) FatalError {
	return FatalError{
		msg: fmt.Sprintf(msg, vars...),
	}
}
