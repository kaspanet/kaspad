package integration

import (
	"testing"
	"time"

	"github.com/kaspanet/kaspad/wire"

	clientpkg "github.com/kaspanet/kaspad/rpc/client"
)

func TestIntegrationBasicSync(t *testing.T) {
	app1, app2, client1, client2, teardown := setup(t)
	defer teardown()

	connect(t, app1, app2, client1, client2)

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

	onBlockAddedChan := make(chan *wire.BlockHeader)
	client2.onBlockAdded = func(header *wire.BlockHeader) {
		onBlockAddedChan <- header
	}

	err = client1.SubmitBlock(block, nil)
	if err != nil {
		t.Fatalf("Error submitting block: %s", err)
	}

	var header *wire.BlockHeader
	select {
	case header = <-onBlockAddedChan:
	case <-time.After(defaultTimeout):
		t.Fatalf("Timeout waiting for block added notification")
	}

	if !header.BlockHash().IsEqual(block.Hash()) {
		t.Errorf("Expected block with hash '%s', but got '%s'", block.Hash(), header.BlockHash())
	}
}
