package main

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kaspanet/kaspad/stability-tests/common"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/pkg/errors"
)

const (
	syncerRPCAddress = "localhost:9000"
	syncedRPCAddress = "localhost:9100"
)

func startNodes() (teardown func(), err error) {
	const (
		syncerListen = "localhost:9001"
		syncedListen = "localhost:9101"
	)

	log.Infof("Starting nodes")
	syncerDataDir, err := common.TempDir("kaspad-datadir-syncer")
	if err != nil {
		panic(errors.Wrapf(err, "error in Tempdir"))
	}
	log.Infof("SYNCER datadir: %s", syncerDataDir)

	syncedDataDir, err := common.TempDir("kaspad-datadir-synced")
	if err != nil {
		panic(errors.Wrapf(err, "error in Tempdir"))
	}
	log.Infof("SYNCED datadir: %s", syncedDataDir)

	syncerCmd, err := common.StartCmd("KASPAD-SYNCER",
		"kaspad",
		common.NetworkCliArgumentFromNetParams(activeConfig().NetParams()),
		"--datadir", syncerDataDir,
		"--logdir", syncerDataDir,
		"--rpclisten", syncerRPCAddress,
		"--listen", syncerListen,
		"--loglevel", "debug",
	)
	if err != nil {
		return nil, err
	}

	syncedCmd, err := common.StartCmd("KASPAD-SYNCED",
		"kaspad",
		common.NetworkCliArgumentFromNetParams(activeConfig().NetParams()),
		"--datadir", syncedDataDir,
		"--logdir", syncedDataDir,
		"--rpclisten", syncedRPCAddress,
		"--listen", syncedListen,
		"--connect", syncerListen,
		"--loglevel", "debug",
	)
	if err != nil {
		return nil, err
	}

	shutdown := uint64(0)

	processesStoppedWg := sync.WaitGroup{}
	processesStoppedWg.Add(2)
	spawn("startNodes-syncerCmd.Wait", func() {
		err := syncerCmd.Wait()
		if err != nil {
			if atomic.LoadUint64(&shutdown) == 0 {
				panics.Exit(log, fmt.Sprintf("syncerCmd closed unexpectedly: %s. See logs at: %s", err, syncerDataDir))
			}
			if !strings.Contains(err.Error(), "signal: killed") {
				panics.Exit(log, fmt.Sprintf("syncerCmd closed with an error: %s. See logs at: %s", err, syncerDataDir))
			}
		}
		processesStoppedWg.Done()
	})

	spawn("startNodes-syncedCmd.Wait", func() {
		err = syncedCmd.Wait()
		if err != nil {
			if atomic.LoadUint64(&shutdown) == 0 {
				panics.Exit(log, fmt.Sprintf("syncedCmd closed unexpectedly: %s. See logs at: %s", err, syncedDataDir))
			}
			if !strings.Contains(err.Error(), "signal: killed") {
				panics.Exit(log, fmt.Sprintf("syncedCmd closed with an error: %s. See logs at: %s", err, syncedDataDir))
			}
		}
		processesStoppedWg.Done()
	})

	// We let the nodes initialize and connect to each other
	log.Infof("Waiting for nodes to start...")
	const initTime = 2 * time.Second
	time.Sleep(initTime)

	return func() {
		atomic.StoreUint64(&shutdown, 1)
		killWithSigterm(syncerCmd, "syncerCmd")
		killWithSigterm(syncedCmd, "syncedCmd")

		processesStoppedChan := make(chan struct{})
		spawn("startNodes-processStoppedWg.Wait", func() {
			processesStoppedWg.Wait()
			processesStoppedChan <- struct{}{}
		})

		const timeout = 10 * time.Second
		select {
		case <-processesStoppedChan:
		case <-time.After(timeout):
			panics.Exit(log, fmt.Sprintf("Processes couldn't be closed after %s", timeout))
		}
	}, nil
}
