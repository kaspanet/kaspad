package main

import (
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/kaspanet/kaspad/stability-tests/common"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/kaspanet/kaspad/util/profiling"
	"testing"
)

const mempoolSizeLimit = 10_000

func TestMempoolLimits(t *testing.T) {
	defer panics.HandlePanic(log, "mempool-limits-main", nil)
	err := parseConfig()
	if err != nil {
		t.Fatalf("error in parseConfig: %s", err)
	}
	defer backendLog.Close()
	common.UseLogger(backendLog, log.Level())

	cfg := activeConfig()
	if cfg.Profile != "" {
		profiling.Start(cfg.Profile, log)
	}

	payAddressKeyPair := decodePayAddressKeyPair(t)
	payToPayAddressScript := buildPayToPayAddressScript(t)
	rpcClient := buildRPCClient(t)

	// Create enough funds for the test
	fundingTransactions := generateFundingCoinbaseTransactions(t, rpcClient)

	// Fill up the mempool to the brim
	submitAnAmountOfTransactionsToTheMempool(t, rpcClient, payAddressKeyPair,
		payToPayAddressScript, fundingTransactions, mempoolSizeLimit, false)
	verifyMempoolSizeEqualTo(t, rpcClient, mempoolSizeLimit)

	// Add some more transactions to the mempool. We expect the
	// mempool to either not grow or even to shrink, since an eviction
	// may also remove any dependant (chained) transactions.
	// Note that we pass ignoreOrphanRejects: true because we
	// expect some of the submitted transactions to depend on
	// transactions that had been evicted from the mempool
	submitAnAmountOfTransactionsToTheMempool(t, rpcClient, payAddressKeyPair,
		payToPayAddressScript, fundingTransactions, 1000, true)
	verifyMempoolSizeEqualToOrLessThan(t, rpcClient, mempoolSizeLimit)

	// Empty mempool out by continuously adding blocks to the DAG
	emptyOutMempool(t, rpcClient)

	log.Infof("mempool-limits passed")
}

func buildRPCClient(t *testing.T) *rpcclient.RPCClient {
	client, err := rpcclient.NewRPCClient(activeConfig().KaspadRPCAddress)
	if err != nil {
		t.Fatalf("error connecting to %s: %s", activeConfig().KaspadRPCAddress, err)
	}
	return client
}

func verifyMempoolSizeEqualTo(t *testing.T, rpcClient *rpcclient.RPCClient, expectedMempoolSize int) {
	getInfoResponse, err := rpcClient.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo: %s", err)
	}
	if getInfoResponse.MempoolSize != uint64(expectedMempoolSize) {
		t.Fatalf("Unexpected mempool size. Want: %d, got: %d",
			expectedMempoolSize, getInfoResponse.MempoolSize)
	}
}

func verifyMempoolSizeEqualToOrLessThan(t *testing.T, rpcClient *rpcclient.RPCClient, expectedMaxMempoolSize int) {
	getInfoResponse, err := rpcClient.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo: %s", err)
	}
	if getInfoResponse.MempoolSize > uint64(expectedMaxMempoolSize) {
		t.Fatalf("Unexpected mempool size. Want: %d, got: %d",
			expectedMaxMempoolSize, getInfoResponse.MempoolSize)
	}
}

func emptyOutMempool(t *testing.T, rpcClient *rpcclient.RPCClient) {
	log.Infof("Adding blocks until mempool shrinks to 0 transactions")
	getInfoResponse, err := rpcClient.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo: %s", err)
	}
	currentMempoolSize := getInfoResponse.MempoolSize
	for currentMempoolSize > 0 {
		mineBlockAndGetCoinbaseTransaction(t, rpcClient)
		getInfoResponse, err := rpcClient.GetInfo()
		if err != nil {
			t.Fatalf("GetInfo: %s", err)
		}
		if getInfoResponse.MempoolSize == currentMempoolSize {
			t.Fatalf("Mempool did not shrink after a block was added to the DAG")
		}
		log.Infof("Mempool shrank from %d transactions to %d transactions",
			currentMempoolSize, getInfoResponse.MempoolSize)
		currentMempoolSize = getInfoResponse.MempoolSize
	}
}
