package common

type wrapError struct {
	msg string
	err error
}

func (e *wrapError) Error() string {
	return e.msg
}

func (e *wrapError) Unwrap() error {
	return e.err
}

func Errorf(msg string, err error) error {
	return &wrapError{
		msg: msg + ": " + err.Error(),
		err: err,
	}
}
