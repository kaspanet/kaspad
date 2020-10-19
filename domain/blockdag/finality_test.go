package blockdag

import (
	"github.com/kaspanet/kaspad/domain/blocknode"
	"testing"

	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util/daghash"
)

func TestFinality(t *testing.T) {
	// Set finalityInterval to 50 blocks, so that test runs quickly
	dagConfig := dagconfig.SimnetParams
	dagConfig.FinalityDuration = 50 * dagConfig.TargetTimePerBlock

	dag, teardownFunc, err := DAGSetup("finality", true, Config{
		DAGParams: &dagConfig,
	})
	if err != nil {
		t.Fatalf("Failed to setup dag instance: %v", err)
	}
	defer teardownFunc()

	// Build a chain of `finalityInterval - 1` blocks
	finalityInterval := dag.FinalityInterval()
	mainChainTip := dag.genesis
	var ok bool
	for i := uint64(0); i < finalityInterval-1; i++ {
		block := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{mainChainTip.Hash}, nil)
		mainChainTip, ok = dag.Index.LookupNode(block.BlockHash())
		if !ok {
			t.Fatalf("Couldn't lookup in blockIndex that was just submitted: %s", block.BlockHash())
		}

		status := dag.Index.BlockNodeStatus(mainChainTip)
		if status != blocknode.StatusValid {
			t.Fatalf("Block #%d in main chain expected to have status '%s', but got '%s'",
				i, blocknode.StatusValid, status)
		}
	}

	// Mine another chain of `finality-Interval - 2` blocks
	sideChainTip := dag.genesis
	for i := uint64(0); i < finalityInterval-2; i++ {
		block := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{sideChainTip.Hash}, nil)
		sideChainTip, ok = dag.Index.LookupNode(block.BlockHash())
		if !ok {
			t.Fatalf("Couldn't lookup in blockIndex that was just submitted: %s", block.BlockHash())
		}

		status := dag.Index.BlockNodeStatus(sideChainTip)
		if status != blocknode.StatusUTXOPendingVerification {
			t.Fatalf("Block #%d in side-chain expected to have status '%s', but got '%s'",
				i, blocknode.StatusUTXOPendingVerification, status)
		}
	}

	// Add two more blocks in the side-chain until it becomes the selected chain
	for i := uint64(0); i < 2; i++ {
		block := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{sideChainTip.Hash}, nil)
		sideChainTip, ok = dag.Index.LookupNode(block.BlockHash())
		if !ok {
			t.Fatalf("Couldn't lookup in blockIndex that was just submitted: %s", block.BlockHash())
		}
	}

	// Make sure that now the sideChainTip is valid and selectedTip
	status := dag.Index.BlockNodeStatus(sideChainTip)
	if status != blocknode.StatusValid {
		t.Fatalf("Overtaking block in side-chain expected to have status '%s', but got '%s'",
			blocknode.StatusValid, status)
	}
	if dag.selectedTip() != sideChainTip {
		t.Fatalf("Overtaking block in side-chain is not selectedTip")
	}

	// Add two more blocks to main chain, to move finality point to first non-genesis block in mainChain
	for i := uint64(0); i < 2; i++ {
		block := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{mainChainTip.Hash}, nil)
		mainChainTip, ok = dag.Index.LookupNode(block.BlockHash())
		if !ok {
			t.Fatalf("Couldn't lookup in blockIndex that was just submitted: %s", block.BlockHash())
		}
	}

	if dag.finalityPoint(dag.virtual.Node) == dag.genesis {
		t.Fatalf("virtual's finalityPoint is still genesis after adding finalityInterval + 1 blocks to the main chain")
	}

	// Subscribe to finality conflict notifications
	notificationChan := make(chan struct{}, 1)
	dag.Subscribe(func(notification *Notification) {
		if notification.Type == NTFinalityConflict {
			notificationChan <- struct{}{}
		}
	})

	// Add two more blocks to the side chain, so that it violates finality and gets status UTXOPendingVerification even
	// though it is the block with the highest blue score.
	for i := uint64(0); i < 2; i++ {
		block := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{sideChainTip.Hash}, nil)
		sideChainTip, ok = dag.Index.LookupNode(block.BlockHash())
		if !ok {
			t.Fatalf("Couldn't lookup in blockIndex that was just submitted: %s", block.BlockHash())
		}
	}

	// Check that sideChainTip is the bluest tip now
	if dag.tips.Bluest() != sideChainTip {
		t.Fatalf("sideChainTip is not the bluest tip when it is expected to be")
	}

	status = dag.Index.BlockNodeStatus(sideChainTip)
	if status != blocknode.StatusUTXOPendingVerification {
		t.Fatalf("Finality violating block expected to have status '%s', but got '%s'",
			blocknode.StatusUTXOPendingVerification, status)
	}

	// Make sure that a finlality conflict notification was sent
	select {
	case <-notificationChan:
	default:
		t.Fatalf("No finality violation notification was sent")
	}
}
