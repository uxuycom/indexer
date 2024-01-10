package xyerrors

import (
	"errors"
	"gopkg.in/go-playground/assert.v1"
	"testing"
)

func TestWrapErrors(t *testing.T) {
	err := errors.New("test errors")

	verifiedErr := ErrDataVerifiedFailed.WrapCause(err)
	assert.Equal(t, errors.Is(verifiedErr, ErrDataVerifiedFailed), true)

	internalErr := ErrInternal.WrapCause(NewInsError(-19, "test errors 2"))
	assert.Equal(t, errors.Is(internalErr, ErrInternal), true)
}
