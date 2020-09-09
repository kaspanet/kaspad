// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"reflect"
	"testing"

	"github.com/kaspanet/kaspad/util/daghash"

	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util"
)

func buildNode(t *testing.T, dag *BlockDAG, parents blockSet) *blockNode {
	block, err := PrepareBlockForTest(dag, parents.hashes(), nil)
	if err != nil {
		t.Fatalf("error in PrepareBlockForTest: %s", err)
	}
	utilBlock := util.NewBlock(block)
	isOrphan, isDelayed, err := dag.ProcessBlock(utilBlock, BFNoPoWCheck)
	if err != nil {
		t.Fatalf("unexpected error in ProcessBlock: %s", err)
	}
	if isDelayed {
		t.Fatalf("block is too far in the future")
	}
	if isOrphan {
		t.Fatalf("block was unexpectedly orphan")
	}
	return nodeByMsgBlock(t, dag, block)
}

// TestTips ensures that tips are updated as expected.
func TestTips(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	params := dagconfig.SimnetParams
	params.K = 1
	dag, teardownFunc, err := DAGSetup("TestTips", true, Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("TestTips: Failed to setup DAG instance: %s", err)
	}
	defer teardownFunc()

	resetExtraNonceForTest()

	// Create a DAG as follows:
	// 0 <- 1 <- 2
	//  \
	//   <- 3 <- 5
	//  \    X
	//   <- 4 <- 6
	node0 := dag.genesis
	node1 := buildNode(t, dag, blockSetFromSlice(node0))
	node2 := buildNode(t, dag, blockSetFromSlice(node1))
	node3 := buildNode(t, dag, blockSetFromSlice(node0))
	node4 := buildNode(t, dag, blockSetFromSlice(node0))
	node5 := buildNode(t, dag, blockSetFromSlice(node3, node4))
	node6 := buildNode(t, dag, blockSetFromSlice(node3, node4))

	// Given an empty VirtualBlock, each of the following test cases will:
	// Set its tips to tipsToSet
	// Add to it all the tips in tipsToAdd, one after the other
	// Call .Tips() on it and compare the result to expectedTips
	// Call .selectedTip() on it and compare the result to expectedSelectedParent
	tests := []struct {
		name                   string
		tipsToSet              []*blockNode
		tipsToAdd              []*blockNode
		expectedTips           blockSet
		expectedSelectedParent *blockNode
	}{
		{
			name:                   "virtual with genesis tip",
			tipsToSet:              []*blockNode{node0},
			tipsToAdd:              []*blockNode{},
			expectedTips:           blockSetFromSlice(node0),
			expectedSelectedParent: node0,
		},
		{
			name:                   "virtual with genesis tip, add child of genesis",
			tipsToSet:              []*blockNode{node0},
			tipsToAdd:              []*blockNode{node1},
			expectedTips:           blockSetFromSlice(node1),
			expectedSelectedParent: node1,
		},
		{
			name:                   "virtual with genesis, add a full DAG",
			tipsToSet:              []*blockNode{node0},
			tipsToAdd:              []*blockNode{node1, node2, node3, node4, node5, node6},
			expectedTips:           blockSetFromSlice(node2, node5, node6),
			expectedSelectedParent: node6,
		},
	}

	for _, test := range tests {
		// Set the tips. This will be the initial state
		_, _, err := dag.setTips(blockSetFromSlice(test.tipsToSet...))
		if err != nil {
			t.Fatalf("%s: Error setting tips: %+v", test.name, err)
		}

		// Add all blockNodes in tipsToAdd in order
		for _, tipToAdd := range test.tipsToAdd {
			addNodeAsChildToParents(tipToAdd)
			_, _, err := dag.addTip(tipToAdd)
			if err != nil {
				t.Fatalf("%s: Error adding tip: %+v", test.name, err)
			}
		}

		// Ensure that the dag's tips are now equal to expectedTips
		resultTips := dag.tips
		if !reflect.DeepEqual(resultTips, test.expectedTips) {
			t.Errorf("%s: unexpected tips. "+
				"Expected: %v, got: %v.", test.name, test.expectedTips, resultTips)
		}

		// Ensure that the virtual block's selectedParent is now equal to expectedSelectedParent
		resultSelectedTip := dag.virtual.selectedParent
		if !reflect.DeepEqual(resultSelectedTip, test.expectedSelectedParent) {
			t.Errorf("%s: unexpected selected tip. "+
				"Expected: %v, got: %v.", test.name, test.expectedSelectedParent, resultSelectedTip)
		}
	}
}

func TestSelectedPath(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	params := dagconfig.SimnetParams
	params.K = 1
	dag, teardownFunc, err := DAGSetup("TestSelectedPath", true, Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("TestSelectedPath: Failed to setup DAG instance: %s", err)
	}
	defer teardownFunc()

	initialPath := blockSetFromSlice(dag.genesis)
	tip := dag.genesis
	for i := 0; i < 5; i++ {
		tipBlock := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{tip.hash}, nil)

		var ok bool
		tip, ok = dag.index.LookupNode(tipBlock.BlockHash())
		if !ok {
			t.Fatalf("Couldn't lookup node that was just added")
		}

		initialPath.add(tip)
	}
	initialTip := tip

	firstPath := initialPath.clone()
	for i := 0; i < 5; i++ {
		tip = buildNode(t, dag, blockSetFromSlice(tip))
		firstPath.add(tip)
		_, _, err = dag.addTip(tip)
		if err != nil {
			t.Fatalf("Error adding tip: %+v", err)
		}
	}
	// For now we don't have any DAG, just chain, the selected path should include all the blocks on the chain.
	if !reflect.DeepEqual(dag.virtual.selectedParentChainSet, firstPath) {
		t.Fatalf("TestSelectedPath: selectedPathSet doesn't include the expected values. got %v, want %v",
			dag.virtual.selectedParent, firstPath)
	}
	// We expect that selectedParentChainSlice should have all the blocks we've added so far
	wantLen := 11
	gotLen := len(dag.virtual.selectedParentChainSlice)
	if wantLen != gotLen {
		t.Fatalf("TestSelectedPath: selectedParentChainSlice doesn't have the expected length. got %d, want %d",
			gotLen, wantLen)
	}

	secondPath := initialPath.clone()
	tip = initialTip
	for i := 0; i < 100; i++ {
		tip = buildNode(t, dag, blockSetFromSlice(tip))
		secondPath.add(tip)
		_, _, err = dag.addTip(tip)
		if err != nil {
			t.Fatalf("Error adding tip: %+v", err)
		}
	}
	// Because we added a chain that is much longer than the previous chain, the selected path should be re-organized.
	if !reflect.DeepEqual(dag.virtual.selectedParentChainSet, secondPath) {
		t.Fatalf("TestSelectedPath: selectedPathSet didn't handle the re-org as expected. got %v, want %v",
			dag.virtual.selectedParent, firstPath)
	}
	// We expect that selectedParentChainSlice should have all the blocks we've added so far except the old chain
	wantLen = 106
	gotLen = len(dag.virtual.selectedParentChainSlice)
	if wantLen != gotLen {
		t.Fatalf("TestSelectedPath: selectedParentChainSlice doesn't have"+
			"the expected length, possibly because it didn't handle the re-org as expected. got %d, want %d", gotLen, wantLen)
	}

	tip = initialTip
	for i := 0; i < 3; i++ {
		tip = buildNode(t, dag, blockSetFromSlice(tip))
		_, _, err = dag.addTip(tip)
		if err != nil {
			t.Fatalf("Error adding tip: %+v", err)
		}
	}
	// Because we added a very short chain, the selected path should not be affected.
	if !reflect.DeepEqual(dag.virtual.selectedParentChainSet, secondPath) {
		t.Fatalf("TestSelectedPath: selectedPathSet did an unexpected re-org. got %v, want %v",
			dag.virtual.selectedParent, firstPath)
	}
	// We expect that selectedParentChainSlice not to change
	wantLen = 106
	gotLen = len(dag.virtual.selectedParentChainSlice)
	if wantLen != gotLen {
		t.Fatalf("TestSelectedPath: selectedParentChainSlice doesn't"+
			"have the expected length, possibly due to unexpected did an unexpected re-org. got %d, want %d", gotLen, wantLen)
	}

	// We call updateSelectedParentSet manually without updating the tips, to check if it panics
	virtual2 := newVirtualBlock(dag, nil)
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("updateSelectedParentSet didn't panic")
		}
	}()
	virtual2.updateSelectedParentSet(buildNode(t, dag, blockSetFromSlice()))
}

// TestChainUpdates makes sure the chainUpdates from setTips are correct:
// It creates two chains: a main-chain to be removed and a side-chain to be added
// The main-chain has to be longer then the side-chain, so that the natural selected tip of the DAG is the one
// from the main chain.
// Then dag.setTip is called with the tip of the side-chain to artificially re-org the DAG, and verify
// the chainUpdates return value is correct.
func TestChainUpdates(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	params := dagconfig.SimnetParams
	params.K = 1
	dag, teardownFunc, err := DAGSetup("TestChainUpdates", true, Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("TestChainUpdates: Failed to setup DAG instance: %s", err)
	}
	defer teardownFunc()

	genesis := dag.genesis

	// Create the main-chain to be removed
	var toBeRemovedNodes []*blockNode
	toBeRemovedTip := genesis
	for i := 0; i < 9; i++ {
		toBeRemovedTip = buildNode(t, dag, blockSetFromSlice(toBeRemovedTip))
		toBeRemovedNodes = append(toBeRemovedNodes, toBeRemovedTip)
	}

	// Create the side-chain to be added
	var toBeAddedNodes []*blockNode
	toBeAddedTip := genesis
	for i := 0; i < 8; i++ {
		toBeAddedTip = buildNode(t, dag, blockSetFromSlice(toBeAddedTip))
		toBeAddedNodes = append(toBeAddedNodes, toBeAddedTip)
	}

	err = resolveNodeStatusForTest(toBeAddedTip)
	if err != nil {
		t.Fatalf("Error resolving status of toBeAddedTip: %+v", err)
	}

	// Set the virtual tip to be the tip of the toBeAdded side-chain
	_, chainUpdates, err := dag.setTips(blockSetFromSlice(toBeAddedTip))
	if err != nil {
		t.Fatalf("Error setting tips: %+v", err)
	}

	// Make sure that the removed blocks are as expected (in reverse order)
	if len(chainUpdates.removedChainBlockHashes) != len(toBeRemovedNodes) {
		t.Fatalf("TestChainUpdates: wrong removed amount. "+
			"Got: %d, want: %d", len(chainUpdates.removedChainBlockHashes), len(toBeRemovedNodes))
	}
	for i, removedHash := range chainUpdates.removedChainBlockHashes {
		correspondingRemovedNode := toBeRemovedNodes[len(toBeRemovedNodes)-1-i]
		if !removedHash.IsEqual(correspondingRemovedNode.hash) {
			t.Fatalf("TestChainUpdates: wrong removed hash. "+
				"Got: %s, want: %s", removedHash, correspondingRemovedNode.hash)
		}
	}

	// Make sure that the added blocks are as expected (in forward order)
	if len(chainUpdates.addedChainBlockHashes) != len(toBeAddedNodes) {
		t.Fatalf("TestChainUpdates: wrong added amount. "+
			"Got: %d, want: %d", len(chainUpdates.addedChainBlockHashes), len(toBeAddedNodes))
	}
	for i, addedHash := range chainUpdates.addedChainBlockHashes {
		correspondingAddedNode := toBeAddedNodes[i]
		if !addedHash.IsEqual(correspondingAddedNode.hash) {
			t.Fatalf("TestChainUpdates: wrong added hash. "+
				"Got: %s, want: %s", addedHash, correspondingAddedNode.hash)
		}
	}
}
