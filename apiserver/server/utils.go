package server

import (
	"fmt"
	"net/http"
)

type handlerError struct {
	ErrorCode    int
	ErrorMessage string
}

func (hErr *handlerError) Error() string {
	return hErr.ErrorMessage
}

func newHandleError(code int, message string) *handlerError {
	return &handlerError{
		ErrorCode:    code,
		ErrorMessage: message,
	}
}

func sendErr(ctx *apiServerContext, w http.ResponseWriter, hErr *handlerError) {
	errMsg := fmt.Sprintf("got error: %s", hErr)
	ctx.warnf(errMsg)
	w.WriteHeader(hErr.ErrorCode)
	sendJSONResponse(w, hErr)
}
