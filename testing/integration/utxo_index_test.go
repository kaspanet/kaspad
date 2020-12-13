package integration

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"testing"
)

func TestUTXOIndex(t *testing.T) {
	harnessParams := &harnessParams{
		p2pAddress:              p2pAddress1,
		rpcAddress:              rpcAddress1,
		miningAddress:           miningAddress1,
		miningAddressPrivateKey: miningAddress1PrivateKey,
		utxoIndex:               true,
	}
	kaspad, teardown := setupHarness(t, harnessParams)
	defer teardown()

	err := kaspad.rpcClient.RegisterForUTXOsChangedNotifications([]string{miningAddress1}, func(
		notification *appmessage.UTXOsChangedNotificationMessage) {

		t.Logf("REMOVED! %v", notification.Removed)
		t.Logf("ADDED! %v", notification.Added)
	})
	if err != nil {
		t.Fatalf("Failed to register for UTXO change notifications: %s", err)
	}

	mineNextBlock(t, kaspad)
}
