package integration

import (
	"testing"
	"time"

	"github.com/kaspanet/kaspad/domainmessage"
)

func TestIntegrationBasicSync(t *testing.T) {
	appHarness1, appHarness2, appHarness3, teardown := standardSetup(t)
	defer teardown()

	// Connect nodes in chain: 1 <--> 2 <--> 3
	// So that node 3 doesn't directly get blocks from node 1
	connect(t, appHarness1, appHarness2)
	connect(t, appHarness2, appHarness3)

	app2OnBlockAddedChan := make(chan *domainmessage.BlockHeader)
	setOnBlockAddedHandler(t, appHarness2, func(header *domainmessage.BlockHeader) {
		app2OnBlockAddedChan <- header
	})

	app3OnBlockAddedChan := make(chan *domainmessage.BlockHeader)
	setOnBlockAddedHandler(t, appHarness3, func(header *domainmessage.BlockHeader) {
		app3OnBlockAddedChan <- header
	})

	block := mineNextBlock(t, appHarness1)

	var header *domainmessage.BlockHeader
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
