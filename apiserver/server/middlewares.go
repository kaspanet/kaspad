package server

import (
	"net/http"
	"runtime/debug"
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
		defer func() {
			recoveryErr := recover()
			if recoveryErr != nil {
				log.Criticalf("Fatal error: %s", recoveryErr)
				log.Criticalf("Stack trace: %s", debug.Stack())
				sendErr(ctx, w, newHandleError(http.StatusInternalServerError, "A server error occurred."))
			}
		}()
		h.ServeHTTP(w, r)
	})
}

func setJSONMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		h.ServeHTTP(w, r)
	})
}
