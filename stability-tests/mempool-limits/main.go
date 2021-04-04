package main

import (
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/kaspanet/kaspad/stability-tests/common"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/kaspanet/kaspad/util/profiling"
	"github.com/pkg/errors"
	"os"
)

const mempoolSizeLimit = 1_000_000

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
	submitAnAmountOfTransactionsToTheMempool(rpcClient, mempoolSizeLimit, false)
	verifyMempoolSizeEqualTo(rpcClient, mempoolSizeLimit)

	// Add some more transactions to the mempool. We expect the
	// mempool to either not grow or even to shrink, since an eviction
	// may also remove any dependant (chained) transactions.
	// Note that we pass ignoreOrphanRejects: true because we
	// expect some of the submitted transactions to depend on
	// transactions that had been evicted from the mempool
	submitAnAmountOfTransactionsToTheMempool(rpcClient, 1000, true)
	verifyMempoolSizeEqualToOrLessThan(rpcClient, mempoolSizeLimit)

	// Empty mempool out by continuously adding blocks to the DAG
	emptyOutMempool(rpcClient)

	log.Infof("mempool-limits passed")
}

func buildRPCClient() *rpcclient.RPCClient {
	client, err := rpcclient.NewRPCClient(activeConfig().KaspadRPCAddress)
	if err != nil {
		panic(errors.Wrapf(err, "error connecting to %s", activeConfig().KaspadRPCAddress))
	}
	return client
}

func verifyMempoolSizeEqualTo(rpcClient *rpcclient.RPCClient, expectedMempoolSize int) {
	getInfoResponse, err := rpcClient.GetInfo()
	if err != nil {
		panic(err)
	}
	if getInfoResponse.MempoolSize != uint64(expectedMempoolSize) {
		panic(errors.Errorf("Unexpected mempool size. Want: %d, got: %d",
			expectedMempoolSize, getInfoResponse.MempoolSize))
	}
}

func verifyMempoolSizeEqualToOrLessThan(rpcClient *rpcclient.RPCClient, expectedMaxMempoolSize int) {
	getInfoResponse, err := rpcClient.GetInfo()
	if err != nil {
		panic(err)
	}
	if getInfoResponse.MempoolSize > uint64(expectedMaxMempoolSize) {
		panic(errors.Errorf("Unexpected mempool size. Want: %d, got: %d",
			expectedMaxMempoolSize, getInfoResponse.MempoolSize))
	}
}

func emptyOutMempool(rpcClient *rpcclient.RPCClient) {
	log.Infof("Adding blocks until mempool shrinks to 0 transactions")
	getInfoResponse, err := rpcClient.GetInfo()
	if err != nil {
		panic(err)
	}
	currentMempoolSize := getInfoResponse.MempoolSize
	for currentMempoolSize > 0 {
		mineBlockAndGetCoinbaseTransaction(rpcClient)
		getInfoResponse, err := rpcClient.GetInfo()
		if err != nil {
			panic(err)
		}
		if getInfoResponse.MempoolSize == currentMempoolSize {
			panic(errors.Errorf("Mempool did not shrink after a block was added to the DAG"))
		}
		log.Infof("Mempool shrank from %d transactions to %d transactions",
			currentMempoolSize, getInfoResponse.MempoolSize)
		currentMempoolSize = getInfoResponse.MempoolSize
	}
}
