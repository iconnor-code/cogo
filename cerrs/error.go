// Package cerrs
package cerrs

import (
	"errors"
	"fmt"
	"runtime"
)

type CError struct {
	stack string
	msg   string
	cause error
}

func (e *CError) Error() string {
	if e.cause == nil {
		return fmt.Sprintf("%s:%s", e.stack, e.msg)
	}
	return fmt.Sprintf("%s:%s\n%s", e.stack, e.msg, e.cause)
}

func New(msg string) error {
	return &CError{
		stack: caller(),
		msg:   msg,
		cause: nil,
	}
}

func Wrap(msg string, err error) error {
	return &CError{
		stack: caller(),
		msg:   msg,
		cause: err,
	}
}

func Unwrap(err error) error {
	if err == nil {
		return nil
	}
	if cerr, ok := err.(*CError); ok {
		return cerr.cause
	}
	return err
}

func Is(err, target error) bool {
	if err == nil || target == nil {
		return err == target
	}
	if cerr, ok := err.(*CError); ok {
		return cerr.cause == target
	}
	return errors.Is(err, target)
}

func As(err error, target any) bool {
	return errors.As(err, target)
}

func caller() string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s:%d", file, line)
}
