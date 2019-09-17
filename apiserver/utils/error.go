package utils

import (
	"fmt"
	"github.com/jinzhu/gorm"
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

// IsDBRecordNotFoundError returns true if the given dbResult is a RecordNotFound error
func IsDBRecordNotFoundError(dbResult *gorm.DB) bool {
	return dbResult.RecordNotFound() && len(dbResult.GetErrors()) == 1
}

// IsDBError returns true if the given dbResult is an error that isn't RecordNotFound
func IsDBError(dbResult *gorm.DB) bool {
	return !IsDBRecordNotFoundError(dbResult) && len(dbResult.GetErrors()) > 0
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
