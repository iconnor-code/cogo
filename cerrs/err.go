// Package cerrs
package cerrs

import (
	"errors"
	"fmt"
	"runtime"
)

type CerrCode int

const InternalErrCode CerrCode = 5000

var InternalError = NewWithCode(InternalErrCode, "internal error")

type CError struct {
	code  CerrCode
	msg   string
	track string
	cause error
}

func (e *CError) Error() string {
	if e.cause == nil {
		return fmt.Sprintf("%s:%d:%s", e.track, e.code, e.msg)
	}
	return fmt.Sprintf("%s:%d:%s\n%s", e.track, e.code, e.msg, e.cause)
}

func (e *CError) GetCode() CerrCode {
	return e.code
}

func New(msg string) error {
	return &CError{
		track: caller(),
		msg:   msg,
		cause: nil,
	}
}

func NewWithCode(code CerrCode, msg string) error {
	return &CError{
		code:  code,
		msg:   msg,
		track: caller(),
		cause: nil,
	}
}

func Wrap(err error, msg ...string) error {
	cerr := &CError{
		track: caller(),
		cause: err,
	}
	if len(msg) > 0 {
		cerr.msg = msg[0]
	}
	return cerr
}

func WrapWithCode(err error, code CerrCode, msg ...string) error {
	cerr := &CError{
		code:  code,
		track: caller(),
		cause: err,
	}
	if len(msg) > 0 {
		cerr.msg = msg[0]
	}
	return cerr
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
