package integration

import (
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"sync"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/app/appmessage"
)

func TestIBD(t *testing.T) {
	const numBlocks = 100

	syncer, syncee, _, teardown := standardSetup(t)
	defer teardown()

	for i := 0; i < numBlocks-1; i++ {
		mineNextBlock(t, syncer)
	}

	blockAddedWG := sync.WaitGroup{}
	blockAddedWG.Add(numBlocks)
	receivedBlocks := 0
	setOnBlockAddedHandler(t, syncee, func(_ *appmessage.BlockAddedNotificationMessage) {
		receivedBlocks++
		blockAddedWG.Done()
	})

	connect(t, syncer, syncee)

	// We expect this to trigger IBD
	mineNextBlock(t, syncer)

	select {
	case <-time.After(defaultTimeout):
		t.Fatalf("Timeout waiting for IBD to finish. Received %d blocks out of %d", receivedBlocks, numBlocks)
	case <-ReceiveFromChanWhenDone(func() { blockAddedWG.Wait() }):
	}

	tip1Hash, err := syncer.rpcClient.GetSelectedTipHash()
	if err != nil {
		t.Fatalf("Error getting tip for syncer")
	}
	tip2Hash, err := syncee.rpcClient.GetSelectedTipHash()
	if err != nil {
		t.Fatalf("Error getting tip for syncee")
	}

	if tip1Hash.SelectedTipHash != tip2Hash.SelectedTipHash {
		t.Errorf("Tips of syncer: '%s' and syncee '%s' are not equal", tip1Hash.SelectedTipHash, tip2Hash.SelectedTipHash)
	}
}

// TestIBDWithPruning checks the IBD from a node with
// already pruned blocks.
func TestIBDWithPruning(t *testing.T) {
	const numBlocks = 100

	overrideDAGParams := dagconfig.SimnetParams

	// This is done to make a pruning depth of 6 blocks
	overrideDAGParams.FinalityDuration = 2 * overrideDAGParams.TargetTimePerBlock
	overrideDAGParams.K = 0
	harnesses, teardown := setupHarnesses(t, []*harnessParams{
		{
			p2pAddress:              p2pAddress1,
			rpcAddress:              rpcAddress1,
			miningAddress:           miningAddress1,
			miningAddressPrivateKey: miningAddress1PrivateKey,
			overrideDAGParams:       &overrideDAGParams,
		},
		{
			p2pAddress:              p2pAddress2,
			rpcAddress:              rpcAddress2,
			miningAddress:           miningAddress2,
			miningAddressPrivateKey: miningAddress2PrivateKey,
			overrideDAGParams:       &overrideDAGParams,
		},
	})
	defer teardown()

	syncer, syncee := harnesses[0], harnesses[1]

	// Let the syncee have two blocks that the syncer
	// doesn't have to test a situation where
	// the block locator will need more than one
	// iteration to find the highest shared chain
	// block.
	mineNextBlock(t, syncee)
	mineNextBlock(t, syncee)

	for i := 0; i < numBlocks-1; i++ {
		mineNextBlock(t, syncer)
	}

	connect(t, syncer, syncee)

	// We expect this to trigger IBD
	mineNextBlock(t, syncer)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	start := time.Now()
	for range ticker.C {
		if time.Since(start) > defaultTimeout {
			t.Fatalf("Timeout waiting for IBD to finish.")
		}

		tip1Hash, err := syncer.rpcClient.GetSelectedTipHash()
		if err != nil {
			t.Fatalf("Error getting tip for syncer")
		}
		tip2Hash, err := syncee.rpcClient.GetSelectedTipHash()
		if err != nil {
			t.Fatalf("Error getting tip for syncee")
		}

		if tip1Hash.SelectedTipHash == tip2Hash.SelectedTipHash {
			break
		}
	}
}
