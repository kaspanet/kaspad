package utils

import (
	"fmt"
	"net/http"
	"strings"
)

// HandlerError is an error returned from
// a rest route handler or a middleware.
type HandlerError struct {
	Code          int
	Message       string
	ClientMessage string
}

func (hErr *HandlerError) Error() string {
	return hErr.Message
}

// NewHandlerError returns a HandlerError with the given code and message.
func NewHandlerError(code int, message string) *HandlerError {
	return &HandlerError{
		Code:          code,
		Message:       message,
		ClientMessage: message,
	}
}

// NewHandlerErrorWithCustomClientMessage returns a HandlerError with
// the given code, message and client error message.
func NewHandlerErrorWithCustomClientMessage(code int, message, clientMessage string) *HandlerError {
	return &HandlerError{
		Code:          code,
		Message:       message,
		ClientMessage: clientMessage,
	}
}

// NewInternalServerHandlerError returns a HandlerError with
// the given message, and the http.StatusInternalServerError
// status text as client message.
func NewInternalServerHandlerError(message string) *HandlerError {
	return NewHandlerErrorWithCustomClientMessage(http.StatusInternalServerError, message, http.StatusText(http.StatusInternalServerError))
}

// NewHandleErrorFromDBErrors takes a slice of database errors and a prefix, and
// returns an HandlerError with error code http.StatusInternalServerError with
// all of the database errors formatted to one string with the given prefix
func NewHandleErrorFromDBErrors(prefix string, dbErrors []error) *HandlerError {
	dbErrorsStrings := make([]string, len(dbErrors))
	for i, dbErr := range dbErrors {
		dbErrorsStrings[i] = fmt.Sprintf("\"%s\"", dbErr)
	}
	errMsg := fmt.Sprintf("%s [%s]", prefix, strings.Join(dbErrorsStrings, ","))
	return NewInternalServerHandlerError(errMsg)
}
