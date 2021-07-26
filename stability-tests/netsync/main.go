package main

import (
	"sync/atomic"

	"github.com/kaspanet/kaspad/stability-tests/common"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/kaspanet/kaspad/util/profiling"
	"github.com/pkg/errors"
)

func main() {
	defer panics.HandlePanic(log, "netsync-main", nil)
	err := parseConfig()
	if err != nil {
		panic(errors.Wrap(err, "error in parseConfig"))
	}
	defer backendLog.Close()
	common.UseLogger(backendLog, log.Level())
	cfg := activeConfig()
	if cfg.Profile != "" {
		profiling.Start(cfg.Profile, log)
	}

	shutdown := uint64(0)

	syncerClient, syncerTeardown, err := setupSyncer()
	if err != nil {
		panic(errors.Wrap(err, "error in setupSyncer"))
	}
	syncerClient.SetOnErrorHandler(func(err error) {
		if atomic.LoadUint64(&shutdown) == 0 {
			log.Debugf("received error from SYNCER: %s", err)
		}
	})
	defer func() {
		syncerClient.Disconnect()
		syncerTeardown()
	}()

	syncedClient, syncedTeardown, err := setupSyncee()
	if err != nil {
		panic(errors.Wrap(err, "error in setupSyncee"))
	}
	syncedClient.SetOnErrorHandler(func(err error) {
		if atomic.LoadUint64(&shutdown) == 0 {
			log.Debugf("received error from SYNCEE: %s", err)
		}
	})
	defer func() {
		syncedClient.Disconnect()
		syncedTeardown()
	}()

	err = checkSyncRate(syncerClient, syncedClient)
	if err != nil {
		panic(errors.Wrap(err, "error in checkSyncRate"))
	}

	err = checkResolveVirtual(syncerClient, syncedClient)
	if err != nil {
		panic(errors.Wrap(err, "error in checkResolveVirtual"))
	}

	atomic.StoreUint64(&shutdown, 1)
}
