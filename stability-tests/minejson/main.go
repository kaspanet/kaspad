package main

import (
<<<<<<< Updated upstream
	"github.com/zoomy-network/zoomyd/domain/consensus"
	"github.com/zoomy-network/zoomyd/stability-tests/common"
	"github.com/zoomy-network/zoomyd/stability-tests/common/mine"
	"github.com/zoomy-network/zoomyd/stability-tests/common/rpc"
	"github.com/zoomy-network/zoomyd/util/panics"
	"github.com/zoomy-network/zoomyd/util/profiling"
=======
>>>>>>> Stashed changes
	"github.com/pkg/errors"
	"github.com/zoomy-network/zoomyd/domain/consensus"
	"github.com/zoomy-network/zoomyd/stability-tests/common"
	"github.com/zoomy-network/zoomyd/stability-tests/common/mine"
	"github.com/zoomy-network/zoomyd/stability-tests/common/rpc"
	"github.com/zoomy-network/zoomyd/util/panics"
	"github.com/zoomy-network/zoomyd/util/profiling"
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
	rpcClient, err := rpc.ConnectToRPC(&cfg.Config, cfg.NetParams())
	if err != nil {
		panic(errors.Wrap(err, "error connecting to JSON-RPC server"))
	}
	defer rpcClient.Disconnect()

	dataDir, err := common.TempDir("minejson")
	if err != nil {
		panic(err)
	}

	consensusConfig := consensus.Config{Params: *cfg.NetParams()}

	err = mine.FromFile(cfg.DAGFile, &consensusConfig, rpcClient, dataDir)
	if err != nil {
		panic(errors.Wrap(err, "error in mine.FromFile"))
	}
}
