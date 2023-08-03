package errors

import (
	"fmt"
)

type ErrFatal struct {
	msg string
}

func (e *ErrFatal) Error() string {
	return e.msg
}

func NewErrFatal(msg string, vars ...interface{}) error {
	return &ErrFatal{
		msg: fmt.Sprintf(msg, vars...),
	}
}

type ErrSeedingCluster struct {
	err error
}

func (e *ErrSeedingCluster) Error() string {
	return e.err.Error()
}

func (e *ErrSeedingCluster) Unwrap() error {
	return e.err
}

func NewErrSeedingCluster(err error, name string) error {
	return &ErrSeedingCluster{
		err: fmt.Errorf("error seeding the cluster with custom template: %s: %w", name, err),
	}
}
