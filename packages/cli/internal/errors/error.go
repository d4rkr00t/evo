package errors

import "fmt"

type EvoError struct {
	code int
	msg  string
	err  error
}

func New(code int, msg string) EvoError {
	return EvoError{
		code: code, msg: msg, err: nil,
	}
}

func Wrap(code int, msg string, err error) EvoError {
	return EvoError{
		code: code, msg: msg, err: err,
	}
}

func (e EvoError) Error() string {
	return fmt.Sprintf("ERROR EVO%04d: %s", e.code, e.msg)
}

func (e EvoError) Unwrap() error {
	return e.err
}
