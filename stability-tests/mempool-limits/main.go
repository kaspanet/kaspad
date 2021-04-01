package main

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/kaspanet/kaspad/stability-tests/common"
	"github.com/kaspanet/kaspad/stability-tests/common/mine"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/kaspanet/kaspad/util/profiling"
	"github.com/pkg/errors"
	"os"
)

func main() {
	defer panics.HandlePanic(log, "mempool-limits-main", nil)
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

	defer func() {
		if err := recover(); err != nil {
			log.Criticalf("mempool-limits failed: %s", err)
			os.Exit(1)
		}
	}()

	rpcClient := buildRPCClient()
	fillUpMempool(rpcClient)

	log.Infof("mempool-limits passed")
}

func buildRPCClient() *rpcclient.RPCClient {
	client, err := rpcclient.NewRPCClient(activeConfig().KaspadRPCAddress)
	if err != nil {
		panic(errors.Wrapf(err, "error connecting to %s", activeConfig().KaspadRPCAddress))
	}
	return client
}

func fillUpMempool(rpcClient *rpcclient.RPCClient) {
	transactionsToGenerate := 1_000_000
	maxTransactionsInBlock := 1_000
	fundingBlocksToGenerate := transactionsToGenerate / maxTransactionsInBlock

	payAddress := "kaspasim:qzpj2cfa9m40w9m2cmr8pvfuqpp32mzzwsuw6ukhfduqpp32mzzws59e8fapc"
	for i := 0; i < fundingBlocksToGenerate; i++ {
		getBlockTemplateResponse, err := rpcClient.GetBlockTemplate(payAddress)
		if err != nil {
			panic(err)
		}
		templateBlock, err := appmessage.RPCBlockToDomainBlock(getBlockTemplateResponse.Block)
		if err != nil {
			panic(err)
		}
		mine.SolveBlock(templateBlock)
		_, err = rpcClient.SubmitBlock(templateBlock)
		if err != nil {
			panic(err)
		}
	}
}
