package errs

import "runtime/debug"

type StackError struct {
	err   error
	stack []byte
}

func WithStack(err error) error {
	if err == nil {
		return nil
	}
	return StackError{err: err, stack: debug.Stack()}
}

func (e StackError) Error() string {
	return e.err.Error()
}

func (e StackError) Unwrap() error {
	return e.err
}

func (e StackError) Stack() []byte {
	return e.stack
}
