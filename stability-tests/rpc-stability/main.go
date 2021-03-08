package main

import (
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient/grpcclient"
	"github.com/kaspanet/kaspad/stability-tests/common"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/kaspanet/kaspad/util/profiling"
	"github.com/pkg/errors"
)

func main() {
	defer panics.HandlePanic(log, "rpc-stability-main", nil)
	err := parseConfig()
	if err != nil {
		panic(errors.Wrap(err, "error parsing configuration"))
	}
	defer backendLog.Close()
	common.UseLogger(backendLog, log.Level())

	cfg := activeConfig()
	if cfg.Profile != "" {
		profiling.Start(cfg.Profile, log)
	}

	rpcAddress, err := cfg.NetParams().NormalizeRPCServerAddress(cfg.RPCServer)
	if err != nil {
		panic(errors.Wrap(err, "error parsing RPC server address"))
	}
	rpcClient, err := grpcclient.Connect(rpcAddress)
	if err != nil {
		panic(errors.Wrap(err, "error connecting to RPC server"))
	}
	defer rpcClient.Disconnect()

	commandsChan, err := readCommands()
	if err != nil {
		panic(errors.Wrapf(err, "error reading commands from file %s", cfg.CommandsFilePath))
	}

	err = sendCommands(rpcClient, commandsChan)
	if err != nil {
		panic(errors.Wrap(err, "error sending commands"))
	}
}
