package main

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/kaspanet/kaspad/version"

	"github.com/pkg/errors"

	_ "net/http/pprof"

	"github.com/kaspanet/kaspad/signal"
	"github.com/kaspanet/kaspad/util/panics"
)

func main() {
	defer panics.HandlePanic(log, nil)
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

	// Enable http profiling server if requested.
	if cfg.Profile != "" {
		spawn(func() {
			listenAddr := net.JoinHostPort("", cfg.Profile)
			log.Infof("Profile server listening on %s", listenAddr)
			profileRedirect := http.RedirectHandler("/debug/pprof", http.StatusSeeOther)
			http.Handle("/", profileRedirect)
			log.Errorf("%s", http.ListenAndServe(listenAddr, nil))
		})
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
