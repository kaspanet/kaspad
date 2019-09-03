package utils

type HandlerError struct {
	ErrorCode    int
	ErrorMessage string
}

func (hErr *HandlerError) Error() string {
	return hErr.ErrorMessage
}

func NewHandlerError(code int, message string) *HandlerError {
	return &HandlerError{
		ErrorCode:    code,
		ErrorMessage: message,
	}
}
