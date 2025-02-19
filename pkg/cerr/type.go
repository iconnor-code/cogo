package cerr

import (
	"fmt"

	"github.com/pkg/errors"
)

// 自定义错误
type CustomError struct {
	Code int
	Msg  string
	Err  error
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("code: %d, msg: %s", e.Code, e.Msg)
}

func (e *CustomError) WithError(err error) error {
	e.Err = err
	return e
}

func NewCustomError(code int, msg string, err error) *CustomError {
	if err != nil {
		err = errors.WithStack(err)
	}
	return &CustomError{
		Code: code,
		Msg:  msg,
		Err:  err,
	}
}

func NewClientError(msg string, err error) *CustomError {
	return NewCustomError(ERR_CODE_CLIENT, msg, err)
}

func NewInternalError(msg string, err error) *CustomError {
	return NewCustomError(ERR_CODE_INTERNAL, msg, err)
}

func NewExternalError(msg string, err error) *CustomError {
	return NewCustomError(ERR_CODE_EXTRA, msg, err)
}

// 对错误进行包装，添加堆栈信息
type stackTracer interface {
	StackTrace() errors.StackTrace
}

func New(msg string) error {
	return errors.New(msg)
}

func WithStack(err error) error {
	if _, ok := err.(stackTracer); ok {
		return err
	}
	return errors.WithStack(err)
}