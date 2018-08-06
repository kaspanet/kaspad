// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"math/rand"
	"reflect"
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

// TestChainView ensures all of the exported functionality of chain views works
// as intended with the exception of some special cases which are handled in
// other tests.
func TestChainView(t *testing.T) {
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
testLoop:
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

		// Ensure all expected nodes are contained in the active view.
		for _, node := range test.contains {
			if !test.view.Contains(node) {
				t.Errorf("%s: expected %v in active view",
					test.name, node)
				continue testLoop
			}
		}

		// Ensure all nodes from side chain view are NOT contained in
		// the active view.
		for _, node := range test.noContains {
			if test.view.Contains(node) {
				t.Errorf("%s: unexpected %v in active view",
					test.name, node)
				continue testLoop
			}
		}

		// Ensure all nodes contained in the view return the expected
		// next node.
		for i, node := range test.contains {
			// Final node expects nil for the next node.
			var expected *blockNode
			if i < len(test.contains)-1 {
				expected = test.contains[i+1]
			}
			if next := test.view.Next(node); next != expected {
				t.Errorf("%s: unexpected next node -- got %v, "+
					"want %v", test.name, next, expected)
				continue testLoop
			}
		}

		// Ensure nodes that are not contained in the view do not
		// produce a successor node.
		for _, node := range test.noContains {
			if next := test.view.Next(node); next != nil {
				t.Errorf("%s: unexpected next node -- got %v, "+
					"want nil", test.name, next)
				continue testLoop
			}
		}

		// Ensure all nodes contained in the view can be retrieved by
		// height.
		for _, wantNode := range test.contains {
			node := test.view.NodeByHeight(wantNode.height)
			if node != wantNode {
				t.Errorf("%s: unexpected node for height %d -- "+
					"got %v, want %v", test.name,
					wantNode.height, node, wantNode)
				continue testLoop
			}
		}

		// Ensure the block locator for the tip of the active view
		// consists of the expected hashes.
		locator := test.view.BlockLocator(test.view.tip())
		if !reflect.DeepEqual(locator, test.locator) {
			t.Errorf("%s: unexpected locator -- got %v, want %v",
				test.name, locator, test.locator)
			continue
		}
	}
}

// TestChainViewSetTip ensures changing the tip works as intended including
// capacity changes.
func TestChainViewSetTip(t *testing.T) {
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
			tips:     []*blockNode{tip(branch0Nodes), nil},
			contains: [][]*blockNode{branch0Nodes, nil},
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
		for i, tip := range test.tips {
			// Ensure the view tip is the expected node.
			test.view.SetTip(tip)
			if test.view.SelectedTip() != tip { // TODO: (Stas) This is wrong. Modified only to satisfy compilation.
				t.Errorf("%s: unexpected view tip -- got %v, "+
					"want %v", test.name, test.view.Tips(),
					tip)
				continue testLoop
			}

			// Ensure all expected nodes are contained in the view.
			for _, node := range test.contains[i] {
				if !test.view.Contains(node) {
					t.Errorf("%s: expected %v in active view",
						test.name, node)
					continue testLoop
				}
			}

		}
	}
}

// TestChainViewNil ensures that creating and accessing a nil chain view behaves
// as expected.
func TestChainViewNil(t *testing.T) {
	view := newVirtualBlock(nil, dagconfig.MainNetParams.K)

	// Ensure the tips of an uninitialized view do not produce a node.
	if tips := view.Tips(); len(tips) > 0 {
		t.Fatalf("Tip: unexpected tips -- got %v, want nothing", tips)
	}

	// Ensure attempting to get a node for a height that does not exist does
	// not produce a node.
	if node := view.NodeByHeight(10); node != nil {
		t.Fatalf("NodeByHeight: unexpected node -- got %v, want nil", node)
	}

	// Ensure an uninitialized view does not report it contains nodes.
	fakeNode := chainedNodes(nil, 1)[0]
	if view.Contains(fakeNode) {
		t.Fatalf("Contains: view claims it contains node %v", fakeNode)
	}

	// Ensure the next node for a node that does not exist does not produce
	// a node.
	if next := view.Next(nil); next != nil {
		t.Fatalf("Next: unexpected next node -- got %v, want nil", next)
	}

	// Ensure the next node for a node that exists does not produce a node.
	if next := view.Next(fakeNode); next != nil {
		t.Fatalf("Next: unexpected next node -- got %v, want nil", next)
	}

	// Ensure attempting to get a block locator for the tip doesn't produce
	// one since the tip is nil.
	if locator := view.BlockLocator(nil); locator != nil {
		t.Fatalf("BlockLocator: unexpected locator -- got %v, want nil",
			locator)
	}

	// Ensure attempting to get a block locator for a node that exists still
	// works as intended.
	branchNodes := chainedNodes(nil, 50)
	wantLocator := locatorHashes(branchNodes, 49, 48, 47, 46, 45, 44, 43,
		42, 41, 40, 39, 38, 36, 32, 24, 8, 0)
	locator := view.BlockLocator(tstTip(branchNodes))
	if !reflect.DeepEqual(locator, wantLocator) {
		t.Fatalf("BlockLocator: unexpected locator -- got %v, want %v",
			locator, wantLocator)
	}
}
