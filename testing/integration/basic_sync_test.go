package integration

import (
	"testing"
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"

	"github.com/kaspanet/kaspad/app/appmessage"
)

func TestIntegrationBasicSync(t *testing.T) {
	appHarness1, appHarness2, appHarness3, teardown := standardSetup(t)
	defer teardown()

	// Connect nodes in chain: 1 <--> 2 <--> 3
	// So that node 3 doesn't directly get blocks from node 1
	connect(t, appHarness1, appHarness2)
	connect(t, appHarness2, appHarness3)

	app2OnBlockAddedChan := make(chan *appmessage.RPCBlock)
	setOnBlockAddedHandler(t, appHarness2, func(notification *appmessage.BlockAddedNotificationMessage) {
		app2OnBlockAddedChan <- notification.BlockVerboseData.Block
	})

	app3OnBlockAddedChan := make(chan *appmessage.RPCBlock)
	setOnBlockAddedHandler(t, appHarness3, func(notification *appmessage.BlockAddedNotificationMessage) {
		app3OnBlockAddedChan <- notification.BlockVerboseData.Block
	})

	block := mineNextBlock(t, appHarness1)

	var rpcBlock *appmessage.RPCBlock
	select {
	case rpcBlock = <-app2OnBlockAddedChan:
	case <-time.After(defaultTimeout):
		t.Fatalf("Timeout waiting for block added notification on node directly connected to miner")
	}
	domainBlockFromRPC, err := appmessage.RPCBlockToDomainBlock(rpcBlock)
	if err != nil {
		t.Fatalf("Could not convert RPC block: %s", err)
	}
	rpcBlockHash := consensushashing.BlockHash(domainBlockFromRPC)

	blockHash := consensushashing.BlockHash(block)
	if !rpcBlockHash.Equal(blockHash) {
		t.Errorf("Expected block with hash '%s', but got '%s'", blockHash, rpcBlockHash)
	}

	select {
	case rpcBlock = <-app3OnBlockAddedChan:
	case <-time.After(defaultTimeout):
		t.Fatalf("Timeout waiting for block added notification on node indirectly connected to miner")
	}
	domainBlockFromRPC, err = appmessage.RPCBlockToDomainBlock(rpcBlock)
	if err != nil {
		t.Fatalf("Could not convert RPC block: %s", err)
	}
	rpcBlockHash = consensushashing.BlockHash(domainBlockFromRPC)

	blockHash = consensushashing.BlockHash(block)
	if !rpcBlockHash.Equal(blockHash) {
		t.Errorf("Expected block with hash '%s', but got '%s'", blockHash, rpcBlockHash)
	}
}
