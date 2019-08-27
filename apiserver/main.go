package main

import (
	"fmt"
	"github.com/daglabs/btcd/signal"
	"github.com/daglabs/btcd/util/panics"
)

func main() {
	defer panics.HandlePanic(log, backendLog)

	cfg, err := parseConfig()
	if err != nil {
		panic(fmt.Errorf("Error parsing command-line arguments: %s", err))
	}

	client, err := connectToServer(cfg)
	if err != nil {
		panic(fmt.Errorf("Error connecting to servers: %s", err))
	}
	defer disconnect(client)

	interrupt := signal.InterruptListener()
	<-interrupt
}

func disconnect(client *apiServerClient) {
	log.Infof("Disconnecting client")
	client.Disconnect()
}
