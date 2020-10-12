package blockdag

import (
	"fmt"
	"github.com/pkg/errors"
	"testing"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util"
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
		block := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{mainChainTip.hash}, nil)
		mainChainTip, ok = dag.index.LookupNode(block.BlockHash())
		if !ok {
			t.Fatalf("Couldn't lookup in blockIndex that was just submitted: %s", block.BlockHash())
		}

		status := dag.index.BlockNodeStatus(mainChainTip)
		if status != statusValid {
			t.Fatalf("Block #%d in main chain expected to have status '%s', but got '%s'",
				i, statusValid, status)
		}
	}

	// Mine another chain of `finality-Interval - 2` blocks
	sideChainTip := dag.genesis
	for i := uint64(0); i < finalityInterval-2; i++ {
		block := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{sideChainTip.hash}, nil)
		sideChainTip, ok = dag.index.LookupNode(block.BlockHash())
		if !ok {
			t.Fatalf("Couldn't lookup in blockIndex that was just submitted: %s", block.BlockHash())
		}

		status := dag.index.BlockNodeStatus(sideChainTip)
		if status != statusUTXOPendingVerification {
			t.Fatalf("Block #%d in side-chain expected to have status '%s', but got '%s'",
				i, statusUTXOPendingVerification, status)
		}
	}

	// Add two more blocks in the side-chain until it becomes the selected chain
	for i := uint64(0); i < 2; i++ {
		block := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{sideChainTip.hash}, nil)
		sideChainTip, ok = dag.index.LookupNode(block.BlockHash())
		if !ok {
			t.Fatalf("Couldn't lookup in blockIndex that was just submitted: %s", block.BlockHash())
		}
	}

	// Make sure that now the sideChainTip is valid and selectedTip
	status := dag.index.BlockNodeStatus(sideChainTip)
	if status != statusValid {
		t.Fatalf("Overtaking block in side-chain expected to have status '%s', but got '%s'",
			statusValid, status)
	}
	if dag.selectedTip() != sideChainTip {
		t.Fatalf("Overtaking block in side-chain is not selectedTip")
	}

	// Add two more blocks to main chain, to move finality point to first non-genesis block in mainChain
	for i := uint64(0); i < 2; i++ {
		block := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{mainChainTip.hash}, nil)
		mainChainTip, ok = dag.index.LookupNode(block.BlockHash())
		if !ok {
			t.Fatalf("Couldn't lookup in blockIndex that was just submitted: %s", block.BlockHash())
		}
	}

	if dag.virtual.finalityPoint() == dag.genesis {
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
		block := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{sideChainTip.hash}, nil)
		sideChainTip, ok = dag.index.LookupNode(block.BlockHash())
		if !ok {
			t.Fatalf("Couldn't lookup in blockIndex that was just submitted: %s", block.BlockHash())
		}
	}

	// Check that sideChainTip is the bluest tip now
	if dag.tips.bluest() != sideChainTip {
		t.Fatalf("sideChainTip is not the bluest tip when it is expected to be")
	}

	status = dag.index.BlockNodeStatus(sideChainTip)
	if status != statusUTXOPendingVerification {
		t.Fatalf("Finality violating block expected to have status '%s', but got '%s'",
			statusUTXOPendingVerification, status)
	}

	// Make sure that a finlality conflict notification was sent
	select {
	case <-notificationChan:
	default:
		t.Fatalf("No finality violation notification was sent")
	}
}

func TestBoundedMergeDepth(t *testing.T) {
	// Set finalityInterval to 50 blocks, so that test runs quickly
	dagConfig := dagconfig.SimnetParams
	dagConfig.FinalityDuration = 50 * dagConfig.TargetTimePerBlock
	cfg := Config{DAGParams: &dagConfig}
	if int(dagConfig.K) >= int(dagConfig.FinalityDuration) {
		t.Fatal("K must be smaller than finality duration for this test to run")
	}

	checkViolatingMergeDepth := func(dag *BlockDAG, parents []*daghash.Hash) (*appmessage.MsgBlock, bool) {
		b, err := PrepareBlockForTest(dag, parents, nil)
		if err != nil {
			t.Fatalf("error in PrepareBlockForTest: %+v", err)
		}
		_, _, err = dag.ProcessBlock(util.NewBlock(b), BFNoPoWCheck)
		if err == nil {
			return b, false
		} else if errors.Cause(err).(RuleError).ErrorCode == ErrViolatingBoundedMergeDepth {
			return b, true
		} else {
			t.Fatalf("expected err: %v, found err: %v", ErrViolatingBoundedMergeDepth, err)
			select {} // fo some reason go doesn't recognize that t.Fatalf never returns
		}
	}
	processBlock := func(dag *BlockDAG, block *appmessage.MsgBlock, name string) {
		isOrphan, isDelayed, err := dag.ProcessBlock(util.NewBlock(block), BFNoPoWCheck)
		if err != nil {
			t.Fatalf("%s got unexpected error from ProcessBlock: %+v", name, err)
		}
		if isOrphan || isDelayed {
			t.Fatalf("%s is unexpectadly orphan: %t or delayed: %t", name, isOrphan, isDelayed)
		}
	}

	getStatus := func(dag *BlockDAG, blockHash *daghash.Hash) blockStatus {
		node, ok := dag.index.LookupNode(blockHash)
		if !ok {
			t.Fatalf("Couldn't lookup in blockIndex that was just submitted: %s", blockHash)
		}
		return dag.index.BlockNodeStatus(node)
	}

	dagBuild, teardownFunc1, err := DAGSetup("BoundedMergeTest", true, cfg)
	if err != nil {
		t.Fatalf("Failed to setup dag instance: %v", err)
	}
	dagReal, teardownFunc2, err := DAGSetup("BoundedMergeReal", true, cfg)
	if err != nil {
		t.Fatalf("Failed to setup dag instance: %v", err)
	}
	defer teardownFunc2()
	finalityInterval := int(dagBuild.FinalityInterval())

	// Create a block on top on genesis
	block1 := PrepareAndProcessBlockForTest(t, dagBuild, []*daghash.Hash{dagBuild.genesis.hash}, nil)

	// Create a chain
	selectedChain := make([]*appmessage.MsgBlock, 0, finalityInterval+2)
	parent := block1.BlockHash()
	// Make sure this is always bigger than `blocksChain2` so it will stay the selected chain
	for i := 0; i < finalityInterval+2; i++ {
		block := PrepareAndProcessBlockForTest(t, dagBuild, []*daghash.Hash{parent}, nil)
		selectedChain = append(selectedChain, block)
		parent = block.BlockHash()
	}

	// Create another chain
	blocksChain2 := make([]*appmessage.MsgBlock, 0, finalityInterval+1)
	parent = block1.BlockHash()
	for i := 0; i < finalityInterval+1; i++ {
		block := PrepareAndProcessBlockForTest(t, dagBuild, []*daghash.Hash{parent}, nil)
		blocksChain2 = append(blocksChain2, block)
		parent = block.BlockHash()
	}

	// Teardown and assign nil to make sure we use the right DAG from here on.
	teardownFunc1()
	dagBuild = nil

	// Now test against the real DAG
	// submit block1
	processBlock(dagReal, block1, "block1")

	// submit chain1
	for i, block := range selectedChain {
		processBlock(dagReal, block, fmt.Sprintf("selectedChain block No %d", i))
	}

	// submit chain2
	for i, block := range blocksChain2 {
		processBlock(dagReal, block, fmt.Sprintf("blocksChain2 block No %d", i))
	}

	// submit a block pointing at tip(chain1) and on block2
	mergeDepthViolatingBlockBottom, isViolatingMergeDepth := checkViolatingMergeDepth(dagReal, []*daghash.Hash{blocksChain2[0].BlockHash(), selectedChain[len(selectedChain)-1].BlockHash()})
	if !isViolatingMergeDepth {
		t.Fatalf("expected mergeDepthViolatingBlockBottom to violate merge depth")
	}

	mergeDepthViolatingTop, isViolatingMergeDepth := checkViolatingMergeDepth(dagReal, []*daghash.Hash{blocksChain2[len(blocksChain2)-1].BlockHash(), selectedChain[len(selectedChain)-1].BlockHash()})
	if !isViolatingMergeDepth {
		t.Fatalf("expected mergeDepthViolatingTop to violate merge depth")
	}

	// the location of the parents in the slices need to be both `-X` so the `selectedChain` one will have higher blueScore (it's a chain longer by 1)
	kosherizingBlock, isViolatingMergeDepth := checkViolatingMergeDepth(dagReal, []*daghash.Hash{blocksChain2[len(blocksChain2)-3].BlockHash(), selectedChain[len(selectedChain)-3].BlockHash()})
	if isViolatingMergeDepth {
		t.Fatalf("expected blueKosherizingBlock to not violate merge depth")
	}
	// Make sure it's actually blue
	found := false
	for _, blue := range dagReal.VirtualBlueHashes() {
		if blue.IsEqual(kosherizingBlock.BlockHash()) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected kosherizingBlock to be blue by the virtual")
	}

	pointAtBlueKosherizing, isViolatingMergeDepth := checkViolatingMergeDepth(dagReal, []*daghash.Hash{kosherizingBlock.BlockHash(), selectedChain[len(selectedChain)-1].BlockHash()})
	if isViolatingMergeDepth {
		t.Fatalf("expected selectedTip to not violate merge depth")
	}
	if !dagReal.SelectedTipHash().IsEqual(pointAtBlueKosherizing.BlockHash()) {
		t.Fatalf("expected %s to be the selectedTip but found %s instead", pointAtBlueKosherizing.BlockHash(), dagReal.SelectedTipHash())
	}

	// Now let's make the kosherizing block red and try to merge again
	tip := selectedChain[len(selectedChain)-1].BlockHash()
	// we use k-1 because `kosherizingBlock` points at tip-2, so 2+k-1 = k+1 anticone.
	for i := 0; i < int(dagReal.Params.K)-1; i++ {
		block := PrepareAndProcessBlockForTest(t, dagReal, []*daghash.Hash{tip}, nil)
		tip = block.BlockHash()
	}

	if !dagReal.SelectedTipHash().IsEqual(tip) {
		t.Fatalf("expected %s to be the selectedTip but found %s instead", tip, dagReal.SelectedTipHash())
	}

	// Make sure it's actually red by the virtual.
	found = false
	for _, blue := range dagReal.VirtualBlueHashes() {
		if blue.IsEqual(kosherizingBlock.BlockHash()) {
			found = true
			break
		}
	}
	if found {
		t.Fatalf("expected kosherizingBlock to be red by the virtual")
	}

	pointAtRedKosherizing, isViolatingMergeDepth := checkViolatingMergeDepth(dagReal, []*daghash.Hash{kosherizingBlock.BlockHash(), tip})
	if !isViolatingMergeDepth {
		t.Fatalf("expected selectedTipRedKosherize to violate merge depth")
	}

	// Now `pointAtBlueKosherizing` itself is actually still blue, so we can still point at that even though we can't point at kosherizing directly anymore
	transitiveBlueKosherizing, isViolatingMergeDepth := checkViolatingMergeDepth(dagReal, []*daghash.Hash{pointAtBlueKosherizing.BlockHash(), tip})
	if isViolatingMergeDepth {
		t.Fatalf("expected transitiveBlueKosherizing to not violate merge depth")
	}
	if !dagReal.SelectedTipHash().IsEqual(transitiveBlueKosherizing.BlockHash()) {
		t.Fatalf("expected %s to be the selectedTip but found %s instead", transitiveBlueKosherizing.BlockHash(), dagReal.SelectedTipHash())
	}

	// Lets validate the status of all the interesting blocks
	if getStatus(dagReal, pointAtBlueKosherizing.BlockHash()) != statusValid {
		t.Fatalf("pointAtBlueKosherizing expected status '%s' but got '%s'", statusValid, getStatus(dagReal, pointAtBlueKosherizing.BlockHash()))
	}
	if getStatus(dagReal, pointAtRedKosherizing.BlockHash()) != statusValidateFailed {
		t.Fatalf("pointAtRedKosherizing expected status '%s' but got '%s'", statusValidateFailed, getStatus(dagReal, pointAtRedKosherizing.BlockHash()))
	}
	if getStatus(dagReal, transitiveBlueKosherizing.BlockHash()) != statusValid {
		t.Fatalf("transitiveBlueKosherizing expected status '%s' but got '%s'", statusValid, getStatus(dagReal, transitiveBlueKosherizing.BlockHash()))
	}
	if getStatus(dagReal, mergeDepthViolatingBlockBottom.BlockHash()) != statusValidateFailed {
		t.Fatalf("mergeDepthViolatingBlockBottom expected status '%s' but got '%s'", statusValidateFailed, getStatus(dagReal, mergeDepthViolatingBlockBottom.BlockHash()))
	}
	if getStatus(dagReal, mergeDepthViolatingTop.BlockHash()) != statusValidateFailed {
		t.Fatalf("mergeDepthViolatingTop expected status '%s' but got '%s'", statusValidateFailed, getStatus(dagReal, mergeDepthViolatingTop.BlockHash()))
	}
	if getStatus(dagReal, kosherizingBlock.BlockHash()) != statusUTXOPendingVerification {
		t.Fatalf("kosherizingBlock expected status '%s' but got '%s'", statusUTXOPendingVerification, getStatus(dagReal, kosherizingBlock.BlockHash()))
	}
	for i, b := range blocksChain2 {
		if getStatus(dagReal, b.BlockHash()) != statusUTXOPendingVerification {
			t.Fatalf("blocksChain2[%d] expected status '%s' but got '%s'", i, statusUTXOPendingVerification, getStatus(dagReal, b.BlockHash()))
		}
	}
	for i, b := range selectedChain {
		if getStatus(dagReal, b.BlockHash()) != statusValid {
			t.Fatalf("selectedChain[%d] expected status '%s' but got '%s'", i, statusValid, getStatus(dagReal, b.BlockHash()))
		}
	}
}
