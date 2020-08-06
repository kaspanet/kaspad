package integration

import (
	"testing"

	"github.com/kaspanet/kaspad/domainmessage"
)

func setOnBlockAddedHandler(t *testing.T, harness *appHarness, handler func(header *domainmessage.BlockHeader)) {
	err := harness.rpcClient.NotifyBlocks()
	if err != nil {
		t.Fatalf("Error from NotifyBlocks: %s", err)
	}
	harness.rpcClient.onBlockAdded = handler
}
