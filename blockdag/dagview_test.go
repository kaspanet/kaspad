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

// TestChainViewNil ensures that creating and accessing a nil chain view behaves
// as expected.
func TestChainViewNil(t *testing.T) {
	// Ensure two unininitialized views are considered equal.
	view := newDAGView(nil)
	if !view.Equals(newDAGView(nil)) {
		t.Fatal("uninitialized nil views unequal")
	}

	// Ensure the genesis of an uninitialized view does not produce a node.
	if genesis := view.Genesis(); genesis != nil {
		t.Fatalf("Genesis: unexpected genesis -- got %v, want nil",
			genesis)
	}

	// Ensure the tips of an uninitialized view do not produce a node.
	if tips := view.Tips(); len(tips) > 0 {
		t.Fatalf("Tip: unexpected tips -- got %v, want nothing", tips)
	}

	// Ensure the height of an uninitialized view is the expected value.
	if height := view.Height(); height != -1 {
		t.Fatalf("Height: unexpected height -- got %d, want -1", height)
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

	// Ensure attempting to find a fork point with a node that doesn't exist
	// doesn't produce a node.
	if fork := view.FindFork(nil); fork != nil {
		t.Fatalf("FindFork: unexpected fork -- got %v, want nil", fork)
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
