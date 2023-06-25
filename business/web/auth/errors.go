package auth

import "fmt"

// AuthError is used to pass an error during the request
// through the application with auth specific context
type AuthError struct {
	msg string
}

// NewAuthError creates and AuthError for the provided message
func NewAuthError(format string, args ...any) error {
	return &AuthError{
		msg: fmt.Sprintf(format, args...),
	}
}

// Error implements the error interface.
// It uses the default message of the wrapped error.
// This is what will be shown in the service's logs.
func (ae *AuthError) Error() string {
	return ae.msg
}
