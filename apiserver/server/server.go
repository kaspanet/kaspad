package server

import (
	"context"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

const gracefulShutdownTimeout = 30 * time.Second

func mainHandler(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintf(w, "API Server is running")
	if err != nil {
		panic(err)
	}
}

// Start starts the HTTP REST server and returns a
// function to gracefully shutdown it.
func Start(listenAddr string) func() {
	router := mux.NewRouter()
	router.Use(addRequestMetaDataMiddleware)
	router.Use(recoveryMiddleware)
	router.Use(loggingMiddleware)
	router.HandleFunc("/", mainHandler)
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
			log.Errorf("Error shutting down http httpServer: %s", err)
		}
	}
}
