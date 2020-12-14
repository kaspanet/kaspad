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
	onUTXOsChangedChan := make(chan *appmessage.UTXOsChangedNotificationMessage, 1000)
	err := kaspad.rpcClient.RegisterForUTXOsChangedNotifications([]string{miningAddress1}, func(
		notification *appmessage.UTXOsChangedNotificationMessage) {

		onUTXOsChangedChan <- notification
	})
	if err != nil {
		t.Fatalf("Failed to register for UTXO change notifications: %s", err)
	}

	// Mine some blocks
	const blockAmountToMine = 100
	for i := 0; i < blockAmountToMine; i++ {
		mineNextBlock(t, kaspad)
	}

	// Collect the UTXO and make sure there's nothing in Removed
	// Note that we expect blockAmountToMine-1 messages because
	// the last block won't be accepted until the next block is
	// mined
	var notificationEntries []*appmessage.UTXOsByAddressesEntry
	for i := 0; i < blockAmountToMine-1; i++ {
		notification := <-onUTXOsChangedChan
		if len(notification.Removed) > 0 {
			t.Fatalf("Unexpectedly received that a UTXO has been removed")
		}
		for _, added := range notification.Added {
			notificationEntries = append(notificationEntries, added)
		}
	}
	for _, notificationEntry := range notificationEntries {
		t.Logf("outpoint: %s:%d",
			notificationEntry.Outpoint.TransactionID, notificationEntry.Outpoint.Index)
	}

	// Get all the UTXOs and make sure the response is equivalent
	// to the data collected via notifications
	utxosByAddressesResponse, err := kaspad.rpcClient.GetUTXOsByAddresses([]string{miningAddress1})
	if err != nil {
		t.Fatalf("Failed to get UTXOs: %s", err)
	}
	if len(notificationEntries) != len(utxosByAddressesResponse.Entries) {
		t.Fatalf("Unexpected amount of UTXOs. Want: %d, got: %d",
			len(notificationEntries), len(utxosByAddressesResponse.Entries))
	}
	for _, notificationEntry := range notificationEntries {
		var foundResponseEntry *appmessage.UTXOsByAddressesEntry
		for _, responseEntry := range utxosByAddressesResponse.Entries {
			if *notificationEntry.Outpoint == *responseEntry.Outpoint {
				foundResponseEntry = responseEntry
				break
			}
		}
		if foundResponseEntry == nil {
			t.Fatalf("Missing entry in UTXOs response: %s:%d",
				notificationEntry.Outpoint.TransactionID, notificationEntry.Outpoint.Index)
		}
		if *notificationEntry.UTXOEntry != *foundResponseEntry.UTXOEntry {
			t.Fatalf("Unexpected UTXOEntry for outpoint %s:%d. Want: %+v, got: %+v",
				notificationEntry.Outpoint.TransactionID, notificationEntry.Outpoint.Index,
				notificationEntry.UTXOEntry, foundResponseEntry.UTXOEntry)
		}
	}
}
