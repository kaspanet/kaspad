package utils

import (
	"fmt"
	"github.com/jinzhu/gorm"
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

// NewHandlerErrorFromDBErrors takes a slice of database errors and a prefix, and
// returns an HandlerError with error code http.StatusInternalServerError with
// all of the database errors formatted to one string with the given prefix
func NewHandlerErrorFromDBErrors(prefix string, dbErrors []error) *HandlerError {
	return NewInternalServerHandlerError(NewErrorFromDBErrors(prefix, dbErrors).Error())
}

// IsDBRecordNotFoundError returns true if the given dbErrors contains only a RecordNotFound error
func IsDBRecordNotFoundError(dbErrors []error) bool {
	return len(dbErrors) == 1 && gorm.IsRecordNotFoundError(dbErrors[0])
}

// HasDBError returns true if the given dbErrors contain any errors that aren't RecordNotFound
func HasDBError(dbErrors []error) bool {
	return !IsDBRecordNotFoundError(dbErrors) && len(dbErrors) > 0
}
