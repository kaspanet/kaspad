package httpserverutils

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"runtime/debug"
)

var nextRequestID uint64 = 1

// AddRequestMetadataMiddleware is a middleware that adds some
// metadata to the context of every request.
func AddRequestMetadataMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rCtx := ToServerContext(r.Context()).SetRequestID(nextRequestID)
		r.WithContext(rCtx)
		nextRequestID++
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware is a middleware that writes
// logs for every request.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := ToServerContext(r.Context())
		ctx.Infof("Method: %s URI: %s", r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

// RecoveryMiddleware is a middleware that recovers
// from panics, log it, and sends Internal Server
// Error to the client.
func RecoveryMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := ToServerContext(r.Context())
		defer func() {
			recoveryErr := recover()
			if recoveryErr != nil {
				var recoveryErrAsError error
				if rErr, ok := recoveryErr.(error); ok {
					recoveryErrAsError = rErr
				} else {
					recoveryErrAsError = errors.Errorf("%s", recoveryErr)
				}
				recoveryErrStr := fmt.Sprintf("%s", recoveryErr)
				log.Criticalf("Fatal error: %+v", recoveryErrStr)
				log.Criticalf("Stack trace: %s", debug.Stack())
				SendErr(ctx, w, recoveryErrAsError)
			}
		}()
		h.ServeHTTP(w, r)
	})
}

// SetJSONMiddleware is a middleware that sets the content type of
// every request to be application/json.
func SetJSONMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		h.ServeHTTP(w, r)
	})
}
