package integration

import (
	"testing"

	"github.com/kaspanet/kaspad/app/appmessage"
)

func setOnBlockAddedHandler(t *testing.T, harness *appHarness, handler func(header *appmessage.BlockHeader)) {
	err := harness.rpcClient.registerForBlockAddedNotifications()
	if err != nil {
		t.Fatalf("Error from NotifyBlocks: %s", err)
	}
	harness.rpcClient.onBlockAdded = handler
}
