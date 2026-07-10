// Package cerrs
package cerrs

import (
	"errors"
	"fmt"
	"runtime"
)

type CerrCode int

type Kind uint8

const (
	SuccessCode    CerrCode = 0
	UnknownErrCode CerrCode = 5000

	KindInternal Kind = iota
	KindInvalidArgument
	KindUnauthenticated
	KindPermissionDenied
	KindNotFound
	KindAlreadyExists
	KindFailedPrecondition
	KindUnavailable
)

type CError struct {
	code  CerrCode
	kind  Kind
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

func (e *CError) Kind() Kind { return e.kind }

func (e *CError) PublicMessage() string { return e.msg }

func (e *CError) Unwrap() error {
	return e.cause
}

func New(msg string) error {
	return &CError{
		code:  UnknownErrCode,
		kind:  KindInternal,
		track: caller(),
		msg:   msg,
		cause: nil,
	}
}

func NewWithCode(code CerrCode, msg string) error {
	return &CError{
		code:  code,
		kind:  kindForLegacyCode(code),
		msg:   msg,
		track: caller(),
		cause: nil,
	}
}

func InvalidArgument(msg string) error    { return NewKind(KindInvalidArgument, msg) }
func Unauthenticated(msg string) error    { return NewKind(KindUnauthenticated, msg) }
func PermissionDenied(msg string) error   { return NewKind(KindPermissionDenied, msg) }
func NotFound(msg string) error           { return NewKind(KindNotFound, msg) }
func AlreadyExists(msg string) error      { return NewKind(KindAlreadyExists, msg) }
func FailedPrecondition(msg string) error { return NewKind(KindFailedPrecondition, msg) }
func Unavailable(msg string) error        { return NewKind(KindUnavailable, msg) }

func NewKind(kind Kind, msg string) error {
	return &CError{code: codeForKind(kind), kind: kind, msg: msg, track: caller()}
}

func Wrap(err error, msg ...string) error {
	cerr := &CError{
		code:  UnknownErrCode,
		kind:  KindInternal,
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
		kind:  kindForLegacyCode(code),
		track: caller(),
		cause: err,
	}
	if len(msg) > 0 {
		cerr.msg = msg[0]
	}
	return cerr
}

func WrapKind(err error, kind Kind, msg string) error {
	return &CError{code: codeForKind(kind), kind: kind, msg: msg, track: caller(), cause: err}
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

func kindForLegacyCode(code CerrCode) Kind {
	switch code {
	case 4000:
		return KindInvalidArgument
	case 4010:
		return KindUnauthenticated
	case 4030:
		return KindPermissionDenied
	case 4040:
		return KindNotFound
	case 4090:
		return KindAlreadyExists
	case 4120:
		return KindFailedPrecondition
	case 5030:
		return KindUnavailable
	default:
		return KindInternal
	}
}

func codeForKind(kind Kind) CerrCode {
	switch kind {
	case KindInvalidArgument:
		return 4000
	case KindUnauthenticated:
		return 4010
	case KindPermissionDenied:
		return 4030
	case KindNotFound:
		return 4040
	case KindAlreadyExists:
		return 4090
	case KindFailedPrecondition:
		return 4120
	case KindUnavailable:
		return 5030
	default:
		return UnknownErrCode
	}
}
