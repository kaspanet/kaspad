package main

import (
	"fmt"
	"github.com/daglabs/btcd/apiserver/server"
	"github.com/daglabs/btcd/logger"
	"github.com/daglabs/btcd/signal"
	"github.com/daglabs/btcd/util/panics"
)

func main() {
	defer panics.HandlePanic(log, logger.BackendLog)

	cfg, err := parseConfig()
	if err != nil {
		panic(fmt.Errorf("Error parsing command-line arguments: %s", err))
	}

	client, err := connectToServer(cfg)
	if err != nil {
		panic(fmt.Errorf("Error connecting to servers: %s", err))
	}
	shutdownServer := server.Start(cfg.HTTPListen)
	defer func() {
		shutdownServer()
		disconnectFromNode(client)
	}()

	interrupt := signal.InterruptListener()
	<-interrupt
}

func disconnectFromNode(client *apiServerClient) {
	log.Infof("Disconnecting client")
	client.Disconnect()
}
