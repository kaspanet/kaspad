package server

import (
	"github.com/daglabs/btcd/apiserver/utils"
	"net/http"
	"runtime/debug"
)

var nextRequestID uint64 = 1

func addRequestMetadataMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rCtx := utils.NewAPIServerContext(r.Context()).SetRequestID(nextRequestID)
		r.WithContext(rCtx)
		nextRequestID++
		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := utils.NewAPIServerContext(r.Context())
		ctx.Infof("Method: %s URI: %s", r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func recoveryMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := utils.NewAPIServerContext(r.Context())
		defer func() {
			recoveryErr := recover()
			if recoveryErr != nil {
				log.Criticalf("Fatal error: %s", recoveryErr)
				log.Criticalf("Stack trace: %s", debug.Stack())
				sendErr(ctx, w, utils.NewHandlerError(http.StatusInternalServerError, "A server error occurred."))
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
