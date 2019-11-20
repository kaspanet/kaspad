package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/daglabs/btcd/signal"
	"github.com/daglabs/btcd/util/panics"
)

func main() {
	defer panics.HandlePanic(log, backendLog, nil)
	cfg, err := parseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing command-line arguments: %s", err)
		os.Exit(1)
	}

	if cfg.Verbose {
		enableRPCLogging()
	}

	connManager, err := newConnectionManager(cfg)
	if err != nil {
		panic(errors.Errorf("Error initializing connection manager: %s", err))
	}
	defer connManager.close()

	spawn(func() {
		err = mineLoop(connManager)
		if err != nil {
			panic(errors.Errorf("Error in main loop: %s", err))
		}
	})

	interrupt := signal.InterruptListener()
	<-interrupt
}

func disconnect(clients []*simulatorClient) {
	for _, client := range clients {
		client.Disconnect()
	}
}
