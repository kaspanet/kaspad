package main

import (
	"context"
	"encoding/json"
	"github.com/kaspanet/kaspad/faucet/config"
	"github.com/kaspanet/kaspad/httpserverutils"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const gracefulShutdownTimeout = 30 * time.Second

// startHTTPServer starts the HTTP REST server and returns a
// function to gracefully shutdown it.
func startHTTPServer(listenAddr string) func() {
	router := mux.NewRouter()
	router.Use(httpserverutils.AddRequestMetadataMiddleware)
	router.Use(httpserverutils.RecoveryMiddleware)
	router.Use(httpserverutils.LoggingMiddleware)
	router.Use(httpserverutils.SetJSONMiddleware)
	router.HandleFunc(
		"/request_money",
		httpserverutils.MakeHandler(requestMoneyHandler)).
		Methods("POST")
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

type requestMoneyData struct {
	Address string `json:"address"`
}

func requestMoneyHandler(_ *httpserverutils.ServerContext, r *http.Request, _ map[string]string, _ map[string]string,
	requestBody []byte) (interface{}, error) {
	hErr := validateIPUsage(r)
	if hErr != nil {
		return nil, hErr
	}
	requestData := &requestMoneyData{}
	err := json.Unmarshal(requestBody, requestData)
	if err != nil {
		return nil, httpserverutils.NewHandlerErrorWithCustomClientMessage(http.StatusUnprocessableEntity,
			errors.Wrap(err, "Error unmarshalling request body"),
			"The request body is not json-formatted")
	}
	address, err := util.DecodeAddress(requestData.Address, config.ActiveNetParams().Prefix)
	if err != nil {
		return nil, httpserverutils.NewHandlerErrorWithCustomClientMessage(http.StatusUnprocessableEntity,
			errors.Wrap(err, "Error decoding address"),
			"Error decoding address")
	}
	tx, err := sendToAddress(address)
	if err != nil {
		return nil, err
	}
	hErr = updateIPUsage(r)
	if hErr != nil {
		return nil, hErr
	}
	return tx.TxID().String(), nil
}
