package server

import (
	"context"
	"fmt"
	"net/http"
)

var nextRequestID uint64 = 1

type contextKey uint32

const (
	contextKeyRequestID contextKey = iota
)

func associateRequestID(r *http.Request) *http.Request {
	ctx := context.WithValue(r.Context(), contextKeyRequestID, nextRequestID)
	nextRequestID++
	return r.WithContext(ctx)
}

func getRequestID(r *http.Request) uint64 {
	id := r.Context().Value(contextKeyRequestID)
	return id.(uint64)
}

func addRequestMetaDataMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = associateRequestID(r)
		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("Request %d: %s %s", getRequestID(r), r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func recoveryMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
				http.Error(w, fmt.Sprintf("got error in request %d: %s", getRequestID(r), errStr), http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, r)
	})
}
