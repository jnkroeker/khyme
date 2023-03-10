package web

import "errors"

type shutdownError struct {
	Message string
}

func NewShutdownError(message string) error {
	return &shutdownError{message}
}

func (se *shutdownError) Error() string {
	return se.Message
}

// errors.As is a type comparison
// it looks to see if the value stored
// inside the error is of a give type,
// in this case shutdownError
func IsShutdown(err error) bool {
	var se *shutdownError
	return errors.As(err, &se)
}
