package main

import (
	"os"
	"sync/atomic"

	"github.com/kaspanet/automation/stability-tests/common"
	"github.com/kaspanet/automation/stability-tests/common/rpc"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/kaspanet/kaspad/util/profiling"
	"github.com/pkg/errors"
)

func main() {
	err := realMain()

	if err != nil {
		log.Criticalf("An error occured: %+v", err)
		backendLog.Close()
		os.Exit(1)
	}
	backendLog.Close()
}

func realMain() error {
	defer panics.HandlePanic(log, "simple-sync-main", nil)

	err := parseConfig()
	if err != nil {
		return errors.Wrap(err, "error in parseConfig")
	}
	common.UseLogger(backendLog, log.Level())
	cfg := activeConfig()
	if cfg.Profile != "" {
		profiling.Start(cfg.Profile, log)
	}

	shutdown := uint64(0)

	teardown, err := startNodes()
	if err != nil {
		return errors.Wrap(err, "error in startNodes")
	}
	defer teardown()

	syncerRPCClient, err := rpc.ConnectToRPC(&rpc.RPCConfig{
		RPCServer: syncerRPCAddress,
	}, activeConfig().NetParams())
	if err != nil {
		return errors.Wrap(err, "error connecting to RPC server")
	}

	syncedRPCClient, err := rpc.ConnectToRPC(&rpc.RPCConfig{
		RPCServer: syncedRPCAddress,
	}, activeConfig().NetParams())
	if err != nil {
		return errors.Wrap(err, "error connecting to RPC server")
	}

	err = syncedRPCClient.RegisterForBlockAddedNotifications()
	if err != nil {
		return errors.Wrap(err, "error registering for blockAdded notifications")
	}

	err = mineLoop(syncerRPCClient, syncedRPCClient)
	if err != nil {
		return errors.Wrap(err, "error in mineLoop")
	}

	atomic.StoreUint64(&shutdown, 1)

	return nil
}
