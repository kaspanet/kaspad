package blockdag

import (
	"testing"

	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/dagconfig/daghash"
)

// TestBlockHeap tests pushing, popping, and determining the length of the heap.
func TestBlockHeap(t *testing.T) {
	block0Header := dagconfig.MainNetParams.GenesisBlock.Header
	block0 := newBlockNode(&block0Header, newSet(), dagconfig.MainNetParams.K)

	block100000Header := Block100000.Header
	block100000 := newBlockNode(&block100000Header, setFromSlice(block0), dagconfig.MainNetParams.K)

	block0smallHash := newBlockNode(&block0Header, newSet(), dagconfig.MainNetParams.K)
	block0smallHash.hash = daghash.Hash{}

	tests := []struct {
		name           string
		toPush         []*blockNode
		expectedLength int
		expectedPop    *blockNode
		direction      HeapDirection
	}{
		{
			name:           "empty heap must have length 0",
			toPush:         []*blockNode{},
			expectedLength: 0,
			expectedPop:    nil,
			direction:      HeapDirectionDown,
		},
		{
			name:           "heap with one push must have length 1",
			toPush:         []*blockNode{block0},
			expectedLength: 1,
			expectedPop:    nil,
			direction:      HeapDirectionDown,
		},
		{
			name:           "heap with one push and one pop",
			toPush:         []*blockNode{block0},
			expectedLength: 0,
			expectedPop:    block0,
			direction:      HeapDirectionDown,
		},
		{
			name:           "push two blocks with different heights, heap shouldn't have to rebalance",
			toPush:         []*blockNode{block100000, block0},
			expectedLength: 1,
			expectedPop:    block100000,
			direction:      HeapDirectionDown,
		},
		{
			name:           "push two blocks with different heights at HeapDirectionUp, heap shouldn't have to rebalance",
			toPush:         []*blockNode{block100000, block0},
			expectedLength: 1,
			expectedPop:    block0,
			direction:      HeapDirectionUp,
		},
		{
			name:           "push two blocks with different heights, heap must rebalance",
			toPush:         []*blockNode{block0, block100000},
			expectedLength: 1,
			expectedPop:    block100000,
			direction:      HeapDirectionDown,
		},
		{
			name:           "push two blocks with equal heights but different hashes, heap shouldn't have to rebalance",
			toPush:         []*blockNode{block0, block0smallHash},
			expectedLength: 1,
			expectedPop:    block0,
			direction:      HeapDirectionDown,
		},
		{
			name:           "push two blocks with equal heights but different hashes, heap must rebalance",
			toPush:         []*blockNode{block0smallHash, block0},
			expectedLength: 1,
			expectedPop:    block0,
			direction:      HeapDirectionDown,
		},
	}

	for _, test := range tests {
		heap := NewHeap(test.direction)
		for _, block := range test.toPush {
			heap.Push(block)
		}

		var poppedBlock *blockNode
		if test.expectedPop != nil {
			poppedBlock = heap.pop()
		}
		if heap.Len() != test.expectedLength {
			t.Errorf("unexpected heap length in test \"%s\". "+
				"Expected: %v, got: %v", test.name, test.expectedLength, heap.Len())
		}
		if poppedBlock != test.expectedPop {
			t.Errorf("unexpected popped block in test \"%s\". "+
				"Expected: %v, got: %v", test.name, test.expectedPop, poppedBlock)
		}
	}
}
