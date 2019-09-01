package server

import (
	"fmt"
	"net/http"
)

var nextRequestID uint64 = 1

func addRequestMetadataMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rCtx := newAPIServerContext(r.Context()).setRequestID(nextRequestID)
		r.WithContext(rCtx)
		nextRequestID++
		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := newAPIServerContext(r.Context())
		ctx.infof("Method: %s URI: %s", r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func recoveryMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := newAPIServerContext(r.Context())
		var errStr string
		defer func() {
			recoveryErr := recover()
			if recoveryErr != nil {
				switch t := recoveryErr.(type) {
				case string:
					errStr = t
				case error:
					errStr = t.Error()
				default:
					errStr = "unknown error"
				}
				ctx.errorf("got error: %s", errStr)
				http.Error(w, fmt.Sprintf("got error in request %d: %s", ctx.requestID(), errStr), http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, r)
	})
}
