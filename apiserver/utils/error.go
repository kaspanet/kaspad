package utils

import "net/http"

// HandlerError is an error returned from
// a rest route handler or a middleware.
type HandlerError struct {
	Code          int `json:"errorCode"`
	Message       string
	ClientMessage string `json:"errorMessage"`
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
