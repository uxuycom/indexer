// Copyright (c) 2023-2024 The UXUY Developer Team
// License:
// MIT License

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE

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
