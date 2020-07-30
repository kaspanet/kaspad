package integration

import (
	"testing"

	"github.com/kaspanet/kaspad/wire"
)

func SetOnBlockAddedHandler(t *testing.T, harness *appHarness, handler func(header *wire.BlockHeader)) {
	err := harness.rpcClient.NotifyBlocks()
	if err != nil {
		t.Fatalf("Error from NotifyBlocks: %+v", err)
	}
	harness.rpcClient.onBlockAdded = handler
}
