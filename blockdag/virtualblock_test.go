// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"reflect"
	"testing"

	"github.com/daglabs/btcd/dagconfig/daghash"
)

func buildNodeGenerator(phantomK uint32) func(parents blockSet) *blockNode {
	// For the purposes of these tests, we'll create blockNodes whose hashes are a
	// series of numbers from 0 to n.
	hashCounter := byte(0)
	return func(parents blockSet) *blockNode {
		block := newBlockNode(nil, parents, phantomK)
		block.hash = daghash.Hash{hashCounter}
		hashCounter++

		return block
	}
}

// TestVirtualBlock ensures that VirtualBlock works as expected.
func TestVirtualBlock(t *testing.T) {
	phantomK := uint32(1)
	buildNode := buildNodeGenerator(phantomK)

	// Create a DAG as follows:
	// 0 <- 1 <- 2
	//  \
	//   <- 3 <- 5
	//  \    X
	//   <- 4 <- 6
	node0 := buildNode(setFromSlice())
	node1 := buildNode(setFromSlice(node0))
	node2 := buildNode(setFromSlice(node1))
	node3 := buildNode(setFromSlice(node0))
	node4 := buildNode(setFromSlice(node0))
	node5 := buildNode(setFromSlice(node3, node4))
	node6 := buildNode(setFromSlice(node3, node4))

	// Given an empty VirtualBlock, each of the following test cases will:
	// Set its tips to tipsToSet
	// Add to it all the tips in tipsToAdd, one after the other
	// Call .Tips() on it and compare the result to expectedTips
	// Call .SelectedTip() on it and compare the result to expectedSelectedTip
	tests := []struct {
		name                string
		tipsToSet           []*blockNode
		tipsToAdd           []*blockNode
		expectedTips        blockSet
		expectedSelectedTip *blockNode
	}{
		{
			name:                "empty virtual",
			tipsToSet:           []*blockNode{},
			tipsToAdd:           []*blockNode{},
			expectedTips:        newSet(),
			expectedSelectedTip: nil,
		},
		{
			name:                "virtual with genesis tip",
			tipsToSet:           []*blockNode{node0},
			tipsToAdd:           []*blockNode{},
			expectedTips:        setFromSlice(node0),
			expectedSelectedTip: node0,
		},
		{
			name:                "virtual with genesis tip, add child of genesis",
			tipsToSet:           []*blockNode{node0},
			tipsToAdd:           []*blockNode{node1},
			expectedTips:        setFromSlice(node1),
			expectedSelectedTip: node1,
		},
		{
			name:                "empty virtual, add a full DAG",
			tipsToSet:           []*blockNode{},
			tipsToAdd:           []*blockNode{node0, node1, node2, node3, node4, node5, node6},
			expectedTips:        setFromSlice(node2, node5, node6),
			expectedSelectedTip: node5,
		},
	}

	for _, test := range tests {
		// Create an empty VirtualBlock
		virtual := newVirtualBlock(nil, phantomK)

		// Set the tips. This will be the initial state
		virtual.SetTips(setFromSlice(test.tipsToSet...))

		// Add all blockNodes in tipsToAdd in order
		for _, tipToAdd := range test.tipsToAdd {
			virtual.AddTip(tipToAdd)
		}

		// Ensure that the virtual block's tips are now equal to expectedTips
		resultTips := virtual.tips()
		if !reflect.DeepEqual(resultTips, test.expectedTips) {
			t.Errorf("unexpected tips in test \"%s\". "+
				"Expected: %v, got: %v.", test.name, test.expectedTips, resultTips)
		}

		// Ensure that the virtual block's selectedTip is now equal to expectedSelectedTip
		resultSelectedTip := virtual.SelectedTip()
		if !reflect.DeepEqual(resultSelectedTip, test.expectedSelectedTip) {
			t.Errorf("unexpected selected tip in test \"%s\". "+
				"Expected: %v, got: %v.", test.name, test.expectedSelectedTip, resultSelectedTip)
		}
	}
}

func TestSelectedPath(t *testing.T) {
	phantomK := uint32(1)
	buildNode := buildNodeGenerator(phantomK)

	// Create an empty VirtualBlock
	virtual := newVirtualBlock(nil, phantomK)

	tip := buildNode(setFromSlice())
	virtual.AddTip(tip)
	initialPath := setFromSlice(tip)
	for i := 0; i < 5; i++ {
		tip = buildNode(setFromSlice(tip))
		initialPath.add(tip)
		virtual.AddTip(tip)
	}
	initialTip := tip

	firstPath := initialPath.clone()
	for i := 0; i < 5; i++ {
		tip = buildNode(setFromSlice(tip))
		firstPath.add(tip)
		virtual.AddTip(tip)
	}
	// For now we don't have any DAG, just chain, the selected path should include all the blocks on the chain.
	if !reflect.DeepEqual(virtual.selectedPathSet, firstPath) {
		t.Fatalf("TestSelectedPath: selectedPathSet doesn't include the expected values. got %v, want %v", virtual.selectedParent, firstPath)
	}

	secondPath := initialPath.clone()
	tip = initialTip
	for i := 0; i < 100; i++ {
		tip = buildNode(setFromSlice(tip))
		secondPath.add(tip)
		virtual.AddTip(tip)
	}
	// Because we added a chain that is much longer than the previous chain, the selected path should be re-organized.
	if !reflect.DeepEqual(virtual.selectedPathSet, secondPath) {
		t.Fatalf("TestSelectedPath: selectedPathSet didn't handle the re-org as expected. got %v, want %v", virtual.selectedParent, firstPath)
	}

	tip = initialTip
	for i := 0; i < 3; i++ {
		tip = buildNode(setFromSlice(tip))
		virtual.AddTip(tip)
	}
	// Because we added a very short chain, the selected path should not be affected.
	if !reflect.DeepEqual(virtual.selectedPathSet, secondPath) {
		t.Fatalf("TestSelectedPath: selectedPathSet didn an unexpected re-org. got %v, want %v", virtual.selectedParent, firstPath)
	}
}
