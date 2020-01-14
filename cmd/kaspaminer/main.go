package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/cmd/kaspaminer/version"
	"github.com/kaspanet/kaspad/signal"
	"github.com/kaspanet/kaspad/util/panics"
)

func main() {
	defer panics.HandlePanic(log, nil, nil)
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
		err = mineLoop(client, cfg.NumberOfBlocks)
		if err != nil {
			panic(errors.Errorf("Error in main loop: %s", err))
		}
		doneChan <- struct{}{}
	})

	select {
	case <-doneChan:
	case <-interrupt:
	}
}
