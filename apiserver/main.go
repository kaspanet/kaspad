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

	db, err := connectToDB(cfg)
	if err != nil {
		panic(fmt.Errorf("Error connecting to database: %s", err))
	}
	defer func() {
		err := db.Close()
		if err != nil {
			panic(fmt.Errorf("Error closing the database: %s", err))
		}
	}()

	client, err := connectToServer(cfg)
	if err != nil {
		panic(fmt.Errorf("Error connecting to servers: %s", err))
	}
	shutdownServer := server.Start(cfg.HTTPListen)
	defer func() {
		shutdownServer()
		disconnectFromNode(client)
	}()

	doneChan := make(chan struct{}, 1)
	spawn(func() {
		err := blockLoop(client, db, doneChan)
		if err != nil {
			panic(err)
		}
	})

	interrupt := signal.InterruptListener()
	<-interrupt

	// Gracefully stop blockLoop
	doneChan <- struct{}{}
}

func disconnectFromNode(client *apiServerClient) {
	log.Infof("Disconnecting client")
	client.Disconnect()
}
