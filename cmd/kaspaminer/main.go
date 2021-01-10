package main

import (
	"fmt"
	"os"

	"github.com/kaspanet/kaspad/util"

	"github.com/kaspanet/kaspad/version"

	"github.com/pkg/errors"

	_ "net/http/pprof"

	"github.com/kaspanet/kaspad/infrastructure/os/signal"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/kaspanet/kaspad/util/profiling"
)

func main() {
	defer panics.HandlePanic(log, "MAIN", nil)
	interrupt := signal.InterruptListener()

	cfg, err := parseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing command-line arguments: %s\n", err)
		os.Exit(1)
	}

	// Show version at startup.
	log.Infof("Version %s", version.Version())

	// Enable http profiling server if requested.
	if cfg.Profile != "" {
		profiling.Start(cfg.Profile, log)
	}

	client, err := newMinerClient(cfg)
	if err != nil {
		panic(errors.Wrap(err, "error connecting to the RPC server"))
	}
	defer client.Disconnect()

	miningAddr, err := util.DecodeAddress(cfg.MiningAddr, cfg.ActiveNetParams.Prefix)
	if err != nil {
		panic(errors.Wrap(err, "error decoding mining address"))
	}

	doneChan := make(chan struct{})
	spawn("mineLoop", func() {
		err = mineLoop(client, cfg.NumberOfBlocks, cfg.TargetBlocksPerSecond, cfg.MineWhenNotSynced, miningAddr)
		if err != nil {
			panic(errors.Wrap(err, "error in mine loop"))
		}
		doneChan <- struct{}{}
	})

	select {
	case <-doneChan:
	case <-interrupt:
	}
}
