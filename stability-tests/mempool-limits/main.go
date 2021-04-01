package main

import (
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/kaspanet/kaspad/stability-tests/common"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/kaspanet/kaspad/util/profiling"
	"github.com/pkg/errors"
	"os"
)

const mempoolSizeLimit = 10_000

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
			backendLog.Close()
			os.Exit(1)
		}
	}()

	rpcClient := buildRPCClient()

	// Create enough funds for the test
	generateFundingCoinbaseTransactions(rpcClient)

	// Fill up the mempool to the brim
	submitAnAmountOfTransactionsToTheMempool(rpcClient, mempoolSizeLimit)
	verifyMempoolSize(rpcClient, mempoolSizeLimit)

	// Add some more transactions to the mempool. We expect the
	// transactions not to be rejected but the mempool not to
	// grow
	submitAnAmountOfTransactionsToTheMempool(rpcClient, 1000)
	verifyMempoolSize(rpcClient, mempoolSizeLimit)

	log.Infof("mempool-limits passed")
}

func buildRPCClient() *rpcclient.RPCClient {
	client, err := rpcclient.NewRPCClient(activeConfig().KaspadRPCAddress)
	if err != nil {
		panic(errors.Wrapf(err, "error connecting to %s", activeConfig().KaspadRPCAddress))
	}
	return client
}

func verifyMempoolSize(rpcClient *rpcclient.RPCClient, expectedMempoolSize int) {
	getInfoResponse, err := rpcClient.GetInfo()
	if err != nil {
		panic(err)
	}
	if getInfoResponse.MempoolSize != uint64(expectedMempoolSize) {
		panic(errors.Errorf("Unexpected mempool size. Want: %d, got: %d",
			expectedMempoolSize, getInfoResponse.MempoolSize))
	}
}
