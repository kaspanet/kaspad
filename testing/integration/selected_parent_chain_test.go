package integration

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"testing"
)

func TestVirtualSelectedParentChain(t *testing.T) {
	// Setup a couple of kaspad instances
	kaspad1, kaspad2, _, teardown := standardSetup(t)
	defer teardown()

	// Register to virtual selected parent chain changes
	onVirtualSelectedParentChainChangedChan := make(chan *appmessage.VirtualSelectedParentChainChangedNotificationMessage)
	err := kaspad1.rpcClient.RegisterForVirtualSelectedParentChainChangedNotifications(
		func(notification *appmessage.VirtualSelectedParentChainChangedNotificationMessage) {
			onVirtualSelectedParentChainChangedChan <- notification
		})
	if err != nil {
		t.Fatalf("Failed to register for virtual selected parent chain change notifications: %s", err)
	}

	// In kaspad1, mine a chain over the genesis and make sure
	// each chain changed notifications contains only one entry
	// in `added` and nothing in `removed`
	const blockAmountToMine = 10
	for i := 0; i < blockAmountToMine; i++ {
		minedBlock := mineNextBlock(t, kaspad1)
		notification := <-onVirtualSelectedParentChainChangedChan
		if len(notification.RemovedChainBlockHashes) > 0 {
			t.Fatalf("RemovedChainBlockHashes is unexpectedly not empty")
		}
		if len(notification.AddedChainBlocks) != 1 {
			t.Fatalf("Unexpected length of AddedChainBlocks. Want: %d, got: %d",
				1, len(notification.AddedChainBlocks))
		}

		minedBlockHash := consensushashing.BlockHash(minedBlock)
		minedBlockHashHex := hex.EncodeToString(minedBlockHash[:])
		if minedBlockHashHex != notification.AddedChainBlocks[0].Hash {
			t.Fatalf("Unexpected block hash in AddedChainBlocks. Want: %s, got: %s",
				minedBlockHashHex, notification.AddedChainBlocks[0].Hash)
		}
	}

	// In kaspad2, mine a different chain of `blockAmountToMine`
	// blocks over the genesis
	for i := 0; i < blockAmountToMine; i++ {
		mineNextBlock(t, kaspad2)
	}

	// Connect the two kaspads
	connect(t, kaspad1, kaspad2)

	// In kaspad2, mine another block. This should trigger sync
	// between the two nodes
	mineNextBlock(t, kaspad2)

	// For the first `blockAmountToMine - 1` blocks we don't expect
	// the chain to change at all
	for i := 0; i < blockAmountToMine-1; i++ {
		notification := <-onVirtualSelectedParentChainChangedChan
		if len(notification.RemovedChainBlockHashes) > 0 {
			t.Fatalf("RemovedChainBlockHashes is unexpectedly not empty")
		}
		if len(notification.AddedChainBlocks) > 0 {
			t.Fatalf("AddedChainBlocks is unexpectedly not empty")
		}
	}

	// Either the next block could cause a reorg or the one
	// after it
	potentialReorgNotification1 := <-onVirtualSelectedParentChainChangedChan
	potentialReorgNotification2 := <-onVirtualSelectedParentChainChangedChan
	var reorgNotification *appmessage.VirtualSelectedParentChainChangedNotificationMessage
	var nonReorgNotification *appmessage.VirtualSelectedParentChainChangedNotificationMessage
	if len(potentialReorgNotification1.RemovedChainBlockHashes) > 0 {
		reorgNotification = potentialReorgNotification1
		nonReorgNotification = potentialReorgNotification2
	} else {
		reorgNotification = potentialReorgNotification2
		nonReorgNotification = potentialReorgNotification1
	}

	// Make sure that the non-reorg notification has nothing
	// in `removed`
	if len(nonReorgNotification.RemovedChainBlockHashes) > 0 {
		t.Fatalf("nonReorgNotification.RemovedChainBlockHashes is unexpectedly not empty")
	}

	// Make sure that the reorg notification contains exactly
	// `blockAmountToMine` blocks in its `removed`
	if len(reorgNotification.RemovedChainBlockHashes) != blockAmountToMine {
		t.Fatalf("Unexpected length of reorgNotification.RemovedChainBlockHashes. Want: %d, got: %d",
			blockAmountToMine, len(reorgNotification.RemovedChainBlockHashes))
	}
}
