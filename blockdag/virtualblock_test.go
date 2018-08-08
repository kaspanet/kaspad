// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"math/rand"
	"testing"

	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/wire"
)

// testNoncePrng provides a deterministic prng for the nonce in generated fake
// nodes.  The ensures that the node have unique hashes.
var testNoncePrng = rand.New(rand.NewSource(0))

// chainedNodes returns the specified number of nodes constructed such that each
// subsequent node points to the previous one to create a chain.  The first node
// will point to the passed parent which can be nil if desired.
func chainedNodes(parents blockSet, numNodes int) []*blockNode {
	nodes := make([]*blockNode, numNodes)
	tips := parents
	for i := 0; i < numNodes; i++ {
		// This is invalid, but all that is needed is enough to get the
		// synthetic tests to work.
		header := wire.BlockHeader{Nonce: testNoncePrng.Uint32()}
		header.PrevBlocks = tips.hashes()
		nodes[i] = newBlockNode(&header, tips, dagconfig.SimNetParams.K)
		tips = setFromSlice(nodes[i])
	}
	return nodes
}

// tstTip is a convenience function to grab the tip of a chain of block nodes
// created via chainedNodes.
func tstTip(nodes []*blockNode) *blockNode {
	return nodes[len(nodes)-1]
}

// locatorHashes is a convenience function that returns the hashes for all of
// the passed indexes of the provided nodes.  It is used to construct expected
// block locators in the tests.
func locatorHashes(nodes []*blockNode, indexes ...int) BlockLocator {
	hashes := make(BlockLocator, 0, len(indexes))
	for _, idx := range indexes {
		hashes = append(hashes, &nodes[idx].hash)
	}
	return hashes
}

// zipLocators is a convenience function that returns a single block locator
// given a variable number of them and is used in the tests.
func zipLocators(locators ...BlockLocator) BlockLocator {
	var hashes BlockLocator
	for _, locator := range locators {
		hashes = append(hashes, locator...)
	}
	return hashes
}

// TestVirtualBlock ensures all of the exported functionality of the virtual block
// works as intended with the exception of some special cases which are handled in
// other tests.
func TestVirtualBlock(t *testing.T) {
	// Construct a synthetic block index consisting of the following
	// structure.
	// 0 -> 1 -> 2  -> 3  -> 4
	//       \-> 2a -> 3a -> 4a  -> 5a -> 6a -> 7a -> ... -> 26a
	//             \-> 3a'-> 4a' -> 5a'
	branch0Nodes := chainedNodes(nil, 5)
	branch1Nodes := chainedNodes(setFromSlice(branch0Nodes[1]), 25)
	branch2Nodes := chainedNodes(setFromSlice(branch1Nodes[0]), 3)

	tip := tstTip
	tests := []struct {
		name       string
		view       *virtualBlock // active view
		genesis    *blockNode    // expected genesis block of active view
		tip        *blockNode    // expected tip of active view
		side       *virtualBlock // side chain view
		sideTip    *blockNode    // expected tip of side chain view
		fork       *blockNode    // expected fork node
		contains   []*blockNode  // expected nodes in active view
		noContains []*blockNode  // expected nodes NOT in active view
		equal      *virtualBlock // view expected equal to active view
		unequal    *virtualBlock // view expected NOT equal to active
		locator    BlockLocator  // expected locator for active view tip
	}{
		{
			// Create a view for branch 0 as the active chain and
			// another view for branch 1 as the side chain.
			name:       "chain0-chain1",
			view:       newVirtualBlock(setFromSlice(tip(branch0Nodes)), dagconfig.MainNetParams.K),
			genesis:    branch0Nodes[0],
			tip:        tip(branch0Nodes),
			side:       newVirtualBlock(setFromSlice(tip(branch1Nodes)), dagconfig.MainNetParams.K),
			sideTip:    tip(branch1Nodes),
			fork:       branch0Nodes[1],
			contains:   branch0Nodes,
			noContains: branch1Nodes,
			equal:      newVirtualBlock(setFromSlice(tip(branch0Nodes)), dagconfig.MainNetParams.K),
			unequal:    newVirtualBlock(setFromSlice(tip(branch1Nodes)), dagconfig.MainNetParams.K),
			locator:    locatorHashes(branch0Nodes, 4, 3, 2, 1, 0),
		},
		{
			// Create a view for branch 1 as the active chain and
			// another view for branch 2 as the side chain.
			name:       "chain1-chain2",
			view:       newVirtualBlock(setFromSlice(tip(branch1Nodes)), dagconfig.MainNetParams.K),
			genesis:    branch0Nodes[0],
			tip:        tip(branch1Nodes),
			side:       newVirtualBlock(setFromSlice(tip(branch2Nodes)), dagconfig.MainNetParams.K),
			sideTip:    tip(branch2Nodes),
			fork:       branch1Nodes[0],
			contains:   branch1Nodes,
			noContains: branch2Nodes,
			equal:      newVirtualBlock(setFromSlice(tip(branch1Nodes)), dagconfig.MainNetParams.K),
			unequal:    newVirtualBlock(setFromSlice(tip(branch2Nodes)), dagconfig.MainNetParams.K),
			locator: zipLocators(
				locatorHashes(branch1Nodes, 24, 23, 22, 21, 20,
					19, 18, 17, 16, 15, 14, 13, 11, 7),
				locatorHashes(branch0Nodes, 1, 0)),
		},
		{
			// Create a view for branch 2 as the active chain and
			// another view for branch 0 as the side chain.
			name:       "chain2-chain0",
			view:       newVirtualBlock(setFromSlice(tip(branch2Nodes)), dagconfig.MainNetParams.K),
			genesis:    branch0Nodes[0],
			tip:        tip(branch2Nodes),
			side:       newVirtualBlock(setFromSlice(tip(branch0Nodes)), dagconfig.MainNetParams.K),
			sideTip:    tip(branch0Nodes),
			fork:       branch0Nodes[1],
			contains:   branch2Nodes,
			noContains: branch0Nodes[2:],
			equal:      newVirtualBlock(setFromSlice(tip(branch2Nodes)), dagconfig.MainNetParams.K),
			unequal:    newVirtualBlock(setFromSlice(tip(branch0Nodes)), dagconfig.MainNetParams.K),
			locator: zipLocators(
				locatorHashes(branch2Nodes, 2, 1, 0),
				locatorHashes(branch1Nodes, 0),
				locatorHashes(branch0Nodes, 1, 0)),
		},
	}
	for _, test := range tests {
		// Ensure the active and side chain tips are the expected nodes.
		if test.view.SelectedTip() != test.tip {
			t.Errorf("%s: unexpected active view tip -- got %v, "+
				"want %v", test.name, test.view.Tips(), test.tip)
			continue
		}
		if test.side.SelectedTip() != test.sideTip {
			t.Errorf("%s: unexpected active view tip -- got %v, "+
				"want %v", test.name, test.side.Tips(),
				test.sideTip)
			continue
		}
	}
}

// TestVirtualBlockSetTips ensures changing the tips works as intended including
// capacity changes.
func TestVirtualBlockSetTips(t *testing.T) {
	// Construct a synthetic block index consisting of the following
	// structure.
	// 0 -> 1 -> 2  -> 3  -> 4
	//       \-> 2a -> 3a -> 4a  -> 5a -> 6a -> 7a -> ... -> 26a
	branch0Nodes := chainedNodes(newSet(), 5)
	branch1Nodes := chainedNodes(setFromSlice(branch0Nodes[1]), 25)

	tip := tstTip
	tests := []struct {
		name     string
		view     *virtualBlock  // active view
		tips     []*blockNode   // tips to set
		contains [][]*blockNode // expected nodes in view for each tip
	}{
		{
			// Create an empty view and set the tip to increasingly
			// longer chains.
			name:     "increasing",
			view:     newVirtualBlock(nil, dagconfig.MainNetParams.K),
			tips:     []*blockNode{tip(branch0Nodes), tip(branch1Nodes)},
			contains: [][]*blockNode{branch0Nodes, branch1Nodes},
		},
		{
			// Create a view with a longer chain and set the tip to
			// increasingly shorter chains.
			name:     "decreasing",
			view:     newVirtualBlock(setFromSlice(tip(branch1Nodes)), dagconfig.MainNetParams.K),
			tips:     []*blockNode{tip(branch0Nodes)},
			contains: [][]*blockNode{branch0Nodes},
		},
		{
			// Create a view with a shorter chain and set the tip to
			// a longer chain followed by setting it back to the
			// shorter chain.
			name:     "small-large-small",
			view:     newVirtualBlock(setFromSlice(tip(branch0Nodes)), dagconfig.MainNetParams.K),
			tips:     []*blockNode{tip(branch1Nodes), tip(branch0Nodes)},
			contains: [][]*blockNode{branch1Nodes, branch0Nodes},
		},
		{
			// Create a view with a longer chain and set the tip to
			// a smaller chain followed by setting it back to the
			// longer chain.
			name:     "large-small-large",
			view:     newVirtualBlock(setFromSlice(tip(branch1Nodes)), dagconfig.MainNetParams.K),
			tips:     []*blockNode{tip(branch0Nodes), tip(branch1Nodes)},
			contains: [][]*blockNode{branch0Nodes, branch1Nodes},
		},
	}

testLoop:
	for _, test := range tests {
		for _, tip := range test.tips {
			// Ensure the view tip is the expected node.
			test.view.SetTips(setFromSlice(tip))
			if test.view.SelectedTip() != tip {
				t.Errorf("%s: unexpected view tip -- got %v, "+
					"want %v", test.name, test.view.Tips(),
					tip)
				continue testLoop
			}
		}
	}
}

// TestVirtualBlockNil ensures that creating and accessing a nil virtualBlock behaves
// as expected.
func TestVirtualBlockNil(t *testing.T) {
	view := newVirtualBlock(nil, dagconfig.MainNetParams.K)

	// Ensure the tips of an uninitialized view do not produce a node.
	if tips := view.Tips(); len(tips) > 0 {
		t.Fatalf("Tip: unexpected tips -- got %v, want nothing", tips)
	}
}
