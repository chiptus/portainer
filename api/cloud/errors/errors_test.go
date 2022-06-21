package errors

import (
	"errors"
	"testing"
)

func TestNewFatalError(t *testing.T) {
	err := NewFatalError("test fatal error")

	var fatal *FatalError
	if !errors.As(err, &fatal) {
		t.Error("A FatalError was expected")
	}

	nonFatal := errors.New("nonFatal error")
	if errors.As(nonFatal, &fatal) {
		t.Error("error expected")
	}
}
