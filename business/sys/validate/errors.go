package validate

import (
	"encoding/json"
	"errors"
)

// naming convention for error variables is starting with ERR
var ErrInvalidID = errors.New("ID is not in the correct form")

// specific to errors encountered with web requests
type ErrorResponse struct {
	Error  string `json:"error"`
	Fields string `json:"fields,omitempty"`
}

type RequestError struct {
	Err    error
	Status int
	Fields error
}

func NewRequestError(err error, status int) error {
	return &RequestError{err, status, nil}
}

// pointer semantics for plain error interface return
func (err *RequestError) Error() string {
	return err.Err.Error()
}

type FieldError struct {
	Field string `json:"field"`
	Err   string `json:"error"`
}

type FieldErrors []FieldError

// any time we use slices, we use value semantics
func (fe FieldErrors) Error() string {
	d, err := json.Marshal(fe)
	if err != nil {
		return err.Error()
	}
	return string(d)
}

func Cause(err error) error {
	root := err
	for {
		if err = errors.Unwrap(root); err == nil {
			return root
		}
		root = err
	}
}
