package errors

import (
	"errors"
	"testing"
)

func TestNewFatalError(t *testing.T) {
	err := NewErrFatal("test fatal error")

	var fatal *ErrFatal
	if !errors.As(err, &fatal) {
		t.Error("An ErrFatal was expected")
	}

	nonFatal := errors.New("nonFatal error")
	if errors.As(nonFatal, &fatal) {
		t.Error("error expected")
	}
}
