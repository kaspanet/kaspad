package utils

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
