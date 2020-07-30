package integration

import (
	"testing"

	"github.com/kaspanet/kaspad/wire"
)

func setOnBlockAddedHandler(t *testing.T, harness *appHarness, handler func(header *wire.BlockHeader)) {
	err := harness.rpcClient.NotifyBlocks()
	if err != nil {
		t.Fatalf("Error from NotifyBlocks: %s", err)
	}
	harness.rpcClient.onBlockAdded = handler
}
