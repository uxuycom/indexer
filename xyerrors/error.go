package xyerrors

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidData        = NewInsError(-100, "invalid data")
	ErrDataVerifiedFailed = NewInsError(-102, "data verified failed")
	ErrInternal           = NewInsError(-500, "internal error")
)

type InsError struct {
	code  int
	msg   string
	cause error
}

func NewInsError(code int, msg string) *InsError {
	return &InsError{
		code: code,
		msg:  msg,
	}
}

func Wrap(e error, code int, message string) *InsError {
	return &InsError{
		msg:   message,
		code:  code,
		cause: e,
	}
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

func (err *InsError) Error() string {
	causeMsg := ""
	if err.cause != nil {
		causeMsg = err.cause.Error()
	}
	return fmt.Sprintf("%s:%d:%s", err.msg, err.code, causeMsg)
}

func (err *InsError) Message() string {
	return err.msg
}

func (err *InsError) Code() int {
	return err.code
}

func (err *InsError) Cause(cause error) error {
	return err.cause
}

func (err *InsError) WrapCause(cause error) *InsError {
	err.cause = cause
	return err
}
