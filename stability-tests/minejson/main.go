package main

import (
	"github.com/kaspanet/automation/stability-tests/common"
	"github.com/kaspanet/automation/stability-tests/common/mine"
	"github.com/kaspanet/automation/stability-tests/common/rpc"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/kaspanet/kaspad/util/profiling"
	"github.com/pkg/errors"
)

func main() {
	defer panics.HandlePanic(log, "minejson-main", nil)
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
	rpcClient, err := rpc.ConnectToRPC(&cfg.RPCConfig, cfg.NetParams())
	if err != nil {
		panic(errors.Wrap(err, "error connecting to JSON-RPC server"))
	}
	defer rpcClient.Disconnect()

	dataDir, err := common.TempDir("minejson")
	if err != nil {
		panic(err)
	}

	err = mine.MineFromFile(cfg.DAGFile, cfg.NetParams(), rpcClient, dataDir)
	if err != nil {
		panic(errors.Wrap(err, "error in MineFromFile"))
	}
}
