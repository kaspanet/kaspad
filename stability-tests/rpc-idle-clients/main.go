package main

import (
	"time"

<<<<<<< Updated upstream
	"github.com/zoomy-network/zoomyd/stability-tests/common"
	"github.com/zoomy-network/zoomyd/stability-tests/common/rpc"
	"github.com/zoomy-network/zoomyd/util/panics"
	"github.com/zoomy-network/zoomyd/util/profiling"
=======
>>>>>>> Stashed changes
	"github.com/pkg/errors"
	"github.com/zoomy-network/zoomyd/stability-tests/common"
	"github.com/zoomy-network/zoomyd/stability-tests/common/rpc"
	"github.com/zoomy-network/zoomyd/util/panics"
	"github.com/zoomy-network/zoomyd/util/profiling"
)

func main() {
	defer panics.HandlePanic(log, "rpc-idle-clients-main", nil)
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

	numRPCClients := cfg.NumClients
	clients := make([]*rpc.Client, numRPCClients)
	for i := uint32(0); i < numRPCClients; i++ {
		rpcClient, err := rpc.ConnectToRPC(&cfg.Config, cfg.NetParams())
		if err != nil {
			panic(errors.Wrap(err, "error connecting to RPC server"))
		}
		clients[i] = rpcClient
	}

	const testDuration = 30 * time.Second
	select {
	case <-time.After(testDuration):
	}
	for _, client := range clients {
		client.Close()
	}
}
