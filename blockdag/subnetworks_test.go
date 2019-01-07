package blockdag

import (
	"encoding/binary"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/wire"
	"testing"
	"time"
)

// TestSubNetworkRegistry tests the full sub-network registry flow. In this test:
// 1. We create a sub-network registry transaction and add its block to the DAG
// 2. Add 2*finalityInterval so that the sub-network registry transaction becomes final
// 3. Attempt to retrieve the gas limit of the now-registered sub-network
func TestSubNetworkRegistry(t *testing.T) {
	params := dagconfig.SimNetParams
	params.K = 1
	dag, teardownFunc, err := DAGSetup("TestFinality", Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()
	blockTime := time.Unix(dag.genesis.timestamp, 0)
	extraNonce := int64(0)

	buildNextBlock := func(parents blockSet, txs []*wire.MsgTx) (*util.Block, error) {
		// We need to change the blockTime to keep all block hashes unique
		blockTime = blockTime.Add(time.Second)

		// We need to change the extraNonce to keep coinbase hashes unique
		extraNonce++

		bh := &wire.BlockHeader{
			Version:      1,
			Bits:         dag.genesis.bits,
			ParentHashes: parents.hashes(),
			Timestamp:    blockTime,
		}
		msgBlock := wire.NewMsgBlock(bh)
		blockHeight := parents.maxHeight() + 1
		coinbaseTx, err := createCoinbaseTx(blockHeight, 1, extraNonce, dag.dagParams)
		if err != nil {
			return nil, err
		}
		_ = msgBlock.AddTransaction(coinbaseTx)

		for _, tx := range txs {
			_ = msgBlock.AddTransaction(tx)
		}

		return util.NewBlock(msgBlock), nil
	}

	addBlockToDAG := func(block *util.Block) (*blockNode, error) {
		dag.dagLock.Lock()
		defer dag.dagLock.Unlock()

		err = dag.maybeAcceptBlock(block, BFNone)
		if err != nil {
			return nil, err
		}

		return dag.index.LookupNode(block.Hash()), nil
	}

	currentNode := dag.genesis

	// Create a block with a valid sub-network registry transaction
	gasLimit := uint64(12345)
	registryTx := wire.NewMsgTx(wire.TxVersion)
	registryTx.SubNetworkID = wire.SubNetworkRegistry
	registryTx.Payload = make([]byte, 8)
	binary.LittleEndian.PutUint64(registryTx.Payload, gasLimit)

	// Add it to the DAG
	registryBlock, err := buildNextBlock(setFromSlice(currentNode), []*wire.MsgTx{registryTx})
	if err != nil {
		t.Fatalf("could not build registry block: %s", err)
	}
	currentNode, err = addBlockToDAG(registryBlock)
	if err != nil {
		t.Fatalf("could not add registry block to DAG: %s", err)
	}

	// Add 2*finalityInterval to ensure the registry transaction is finalized
	for currentNode.blueScore < 2*finalityInterval {
		nextBlock, err := buildNextBlock(setFromSlice(currentNode), []*wire.MsgTx{})
		if err != nil {
			t.Fatalf("could not create block: %s", err)
		}
		currentNode, err = addBlockToDAG(nextBlock)
		if err != nil {
			t.Fatalf("could not add block to DAG: %s", err)
		}
	}

	// Make sure that the sub-network had been successfully registered by trying
	// to retrieve its gas limit.
	mostRecentlyRegisteredSubNetworkID := dag.lastSubNetworkID - 1
	limit, err := dag.GasLimit(mostRecentlyRegisteredSubNetworkID)
	if err != nil {
		t.Fatalf("could not retrieve gas limit: %s", err)
	}
	if limit != gasLimit {
		t.Fatalf("unexpected gas limit. want: %d, got: %d", gasLimit, limit)
	}
}
