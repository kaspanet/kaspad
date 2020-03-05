package main

import (
	"fmt"
	"github.com/kaspanet/kaspad/version"
	"os"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/signal"
	"github.com/kaspanet/kaspad/util/panics"
)

func main() {
	defer panics.HandlePanic(backendLog, nil)
	interrupt := signal.InterruptListener()

	cfg, err := parseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing command-line arguments: %s\n", err)
		os.Exit(1)
	}

	// Show version at startup.
	log.Infof("Version %s", version.Version())

	if cfg.Verbose {
		enableRPCLogging()
	}

	client, err := connectToServer(cfg)
	if err != nil {
		panic(errors.Wrap(err, "Error connecting to the RPC server"))
	}
	defer client.Disconnect()

	doneChan := make(chan struct{})
	spawn(func() {
		err = mineLoop(client, cfg.NumberOfBlocks, cfg.BlockDelay)
		if err != nil {
			panic(errors.Errorf("Error in mine loop: %s", err))
		}
		doneChan <- struct{}{}
	})

	select {
	case <-doneChan:
	case <-interrupt:
	}
}
