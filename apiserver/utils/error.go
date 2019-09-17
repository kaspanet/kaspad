package utils

import (
	"fmt"
	"strings"
)

// HandlerError is an error returned from
// a rest route handler or a middleware.
type HandlerError struct {
	ErrorCode    int
	ErrorMessage string
}

func (hErr *HandlerError) Error() string {
	return hErr.ErrorMessage
}

// NewHandlerError returns a HandlerError with the given code and message.
func NewHandlerError(code int, message string) *HandlerError {
	return &HandlerError{
		ErrorCode:    code,
		ErrorMessage: message,
	}
}

// NewErrorFromDBErrors takes a slice of database errors and a prefix, and
// returns an error with all of the database errors formatted to one string with
// the given prefix
func NewErrorFromDBErrors(prefix string, dbErrors []error) error {
	dbErrorsStrings := make([]string, len(dbErrors))
	for i, dbErr := range dbErrors {
		dbErrorsStrings[i] = fmt.Sprintf("\"%s\"", dbErr)
	}
	return fmt.Errorf("%s [%s]", prefix, strings.Join(dbErrorsStrings, ","))
}
