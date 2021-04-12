package mempoollimits

import (
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/kaspanet/kaspad/stability-tests/common"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/kaspanet/kaspad/util/profiling"
	"os"
	"testing"
)

const (
	mempoolSizeLimit        = 1_000_000
	overfillMempoolByAmount = 1_000
)

func TestMempoolLimits(t *testing.T) {
	if os.Getenv("RUN_STABILITY_TESTS") == "" {
		t.Skip()
	}

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

	// Make sure that the mempool size is exactly the limit
	mempoolSize := getMempoolSize(t, rpcClient)
	if mempoolSize != mempoolSizeLimit {
		t.Fatalf("Unexpected mempool size. Want: %d, got: %d",
			mempoolSizeLimit, mempoolSize)
	}

	// Add some more transactions to the mempool. We expect the
	// mempool to either not grow or even to shrink, since an eviction
	// may also remove any dependant (chained) transactions.
	// Note that we pass ignoreOrphanRejects: true because we
	// expect some of the submitted transactions to depend on
	// transactions that had been evicted from the mempool
	submitAnAmountOfTransactionsToTheMempool(t, rpcClient, payAddressKeyPair,
		payToPayAddressScript, fundingTransactions, overfillMempoolByAmount, true)

	// Make sure that the mempool size is the limit or smaller
	mempoolSize = getMempoolSize(t, rpcClient)
	if mempoolSize > mempoolSizeLimit {
		t.Fatalf("Unexpected mempool size. Want at most: %d, got: %d",
			mempoolSizeLimit, mempoolSize)
	}

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

func getMempoolSize(t *testing.T, rpcClient *rpcclient.RPCClient) uint64 {
	getInfoResponse, err := rpcClient.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo: %+v", err)
	}
	return getInfoResponse.MempoolSize
}

func emptyOutMempool(t *testing.T, rpcClient *rpcclient.RPCClient) {
	log.Infof("Adding blocks until mempool shrinks to 0 transactions")
	getInfoResponse, err := rpcClient.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo: %+v", err)
	}
	currentMempoolSize := getInfoResponse.MempoolSize
	for currentMempoolSize > 0 {
		mineBlockAndGetCoinbaseTransaction(t, rpcClient)
		getInfoResponse, err := rpcClient.GetInfo()
		if err != nil {
			t.Fatalf("GetInfo: %+v", err)
		}
		if getInfoResponse.MempoolSize == currentMempoolSize {
			t.Fatalf("Mempool did not shrink after a block was added to the DAG")
		}
		log.Infof("Mempool shrank from %d transactions to %d transactions",
			currentMempoolSize, getInfoResponse.MempoolSize)
		currentMempoolSize = getInfoResponse.MempoolSize
	}
}
