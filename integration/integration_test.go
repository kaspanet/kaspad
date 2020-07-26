package integration

import (
	"testing"
	"time"

	"github.com/kaspanet/kaspad/wire"

	clientpkg "github.com/kaspanet/kaspad/rpc/client"
)

func TestIntegrationBasicSync(t *testing.T) {
	app1, app2, app3, client1, client2, client3, teardown := setup(t)
	defer teardown()

	// Connect nodes in chain: 1 <--> 2 <--> 3
	// So that node 3 doesn't directly get blocks from node 1
	connect(t, app1, app2, client1, client2, p2pAddress1)
	connect(t, app2, app3, client2, client3, p2pAddress2)

	blockTemplate, err := client1.GetBlockTemplate(testAddress1, "")
	if err != nil {
		t.Fatalf("Error getting block template: %+v", err)
	}

	block, err := clientpkg.ConvertGetBlockTemplateResultToBlock(blockTemplate)
	if err != nil {
		t.Fatalf("Error parsing blockTemplate: %s", err)
	}

	solveBlock(t, block)

	err = client2.NotifyBlocks()
	if err != nil {
		t.Fatalf("Error from NotifyBlocks: %+v", err)
	}
	app2OnBlockAddedChan := make(chan *wire.BlockHeader)
	client2.onBlockAdded = func(header *wire.BlockHeader) {
		app2OnBlockAddedChan <- header
	}

	err = client3.NotifyBlocks()
	if err != nil {
		t.Fatalf("Error from NotifyBlocks: %+v", err)
	}
	app3OnBlockAddedChan := make(chan *wire.BlockHeader)
	client3.onBlockAdded = func(header *wire.BlockHeader) {
		app3OnBlockAddedChan <- header
	}

	err = client1.SubmitBlock(block, nil)
	if err != nil {
		t.Fatalf("Error submitting block: %s", err)
	}

	var header *wire.BlockHeader
	select {
	case header = <-app2OnBlockAddedChan:
	case <-time.After(defaultTimeout):
		t.Fatalf("Timeout waiting for block added notification on node directly connected to miner")
	}

	if !header.BlockHash().IsEqual(block.Hash()) {
		t.Errorf("Expected block with hash '%s', but got '%s'", block.Hash(), header.BlockHash())
	}

	select {
	case header = <-app3OnBlockAddedChan:
	case <-time.After(defaultTimeout):
		t.Fatalf("Timeout waiting for block added notification on node indirectly connected to miner")
	}

	if !header.BlockHash().IsEqual(block.Hash()) {
		t.Errorf("Expected block with hash '%s', but got '%s'", block.Hash(), header.BlockHash())
	}
}
