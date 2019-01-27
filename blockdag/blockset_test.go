package blockdag

import (
	"testing"

	"github.com/daglabs/btcd/dagconfig/daghash"
)

func TestHashes(t *testing.T) {
	bs := setFromSlice(
		&blockNode{
			hash: daghash.Hash{3},
		},
		&blockNode{
			hash: daghash.Hash{1},
		},
		&blockNode{
			hash: daghash.Hash{0},
		},
		&blockNode{
			hash: daghash.Hash{2},
		},
	)

	expected := []daghash.Hash{
		{0},
		{1},
		{2},
		{3},
	}

	if !daghash.AreEqual(bs.hashes(), expected) {
		t.Errorf("TestHashes: hashes are not ordered as expected")
	}
}
func TestBlockSetHighest(t *testing.T) {
	node1 := &blockNode{hash: daghash.Hash{10}, height: 1}
	node2a := &blockNode{hash: daghash.Hash{20}, height: 2}
	node2b := &blockNode{hash: daghash.Hash{21}, height: 2}
	node3 := &blockNode{hash: daghash.Hash{30}, height: 3}

	tests := []struct {
		name            string
		set             blockSet
		expectedHighest *blockNode
	}{
		{
			name:            "empty set returns nil",
			set:             setFromSlice(),
			expectedHighest: nil,
		},
		{
			name:            "set with one member returns that member",
			set:             setFromSlice(node1),
			expectedHighest: node1,
		},
		{
			name:            "same-height highest decided by node hash",
			set:             setFromSlice(node2b, node1, node2a),
			expectedHighest: node2a,
		},
		{
			name:            "typical set",
			set:             setFromSlice(node2b, node3, node1, node2a),
			expectedHighest: node3,
		},
	}

	for _, test := range tests {
		highest := test.set.highest()
		if highest != test.expectedHighest {
			t.Errorf("blockSet.highest: unexpected value in test '%s'. " +
				"Expected: %v, got: %v", test.name, test.expectedHighest, highest)
		}
	}
}
