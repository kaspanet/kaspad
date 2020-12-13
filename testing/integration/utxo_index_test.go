package integration

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"testing"
)

func TestUTXOIndex(t *testing.T) {
	// Setup a single kaspad instance
	harnessParams := &harnessParams{
		p2pAddress:              p2pAddress1,
		rpcAddress:              rpcAddress1,
		miningAddress:           miningAddress1,
		miningAddressPrivateKey: miningAddress1PrivateKey,
		utxoIndex:               true,
	}
	kaspad, teardown := setupHarness(t, harnessParams)
	defer teardown()

	// skip the first block because it's paying to genesis script,
	// which contains no outputs
	mineNextBlock(t, kaspad)

	// Register for UTXO changes
	err := kaspad.rpcClient.RegisterForUTXOsChangedNotifications([]string{miningAddress1}, func(
		notification *appmessage.UTXOsChangedNotificationMessage) {

		for _, removed := range notification.Removed {
			t.Logf("REMOVED! Address: %s, outpoint: %s:%d", removed.Address,
				removed.Outpoint.TransactionID, removed.Outpoint.Index)
		}
		for _, added := range notification.Added {
			t.Logf("ADDED! Address: %s, outpoint: %s:%d, utxoEntry: %d:%s:%d:%t", added.Address,
				added.Outpoint.TransactionID, added.Outpoint.Index,
				added.UTXOEntry.Amount, added.UTXOEntry.ScriptPubKey, added.UTXOEntry.BlockBlueScore, added.UTXOEntry.IsCoinbase)
		}
	})
	if err != nil {
		t.Fatalf("Failed to register for UTXO change notifications: %s", err)
	}

	// Mine some blocks
	for i := 0; i < 100; i++ {
		mineNextBlock(t, kaspad)
	}
}
