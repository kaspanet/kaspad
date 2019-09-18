package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const gracefulShutdownTimeout = 30 * time.Second

// Start starts the HTTP REST server and returns a
// function to gracefully shutdown it.
func Start(listenAddr string) func() {
	router := mux.NewRouter()
	router.Use(addRequestMetadataMiddleware)
	router.Use(recoveryMiddleware)
	router.Use(loggingMiddleware)
	router.Use(setJSONMiddleware)
	addRoutes(router)
	httpServer := &http.Server{
		Addr:    listenAddr,
		Handler: handlers.CORS()(router),
	}
	spawn(func() {
		log.Errorf("%s", httpServer.ListenAndServe())
	})

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
		defer cancel()
		err := httpServer.Shutdown(ctx)
		if err != nil {
			log.Errorf("Error shutting down HTTP server: %s", err)
		}
	}
}
