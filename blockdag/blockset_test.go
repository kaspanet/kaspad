package blockdag

import (
	"reflect"
	"testing"

	"github.com/daglabs/btcd/util/daghash"
)

func TestHashes(t *testing.T) {
	bs := setFromSlice(
		&blockNode{
			hash: &daghash.Hash{3},
		},
		&blockNode{
			hash: &daghash.Hash{1},
		},
		&blockNode{
			hash: &daghash.Hash{0},
		},
		&blockNode{
			hash: &daghash.Hash{2},
		},
	)

	expected := []*daghash.Hash{
		{0},
		{1},
		{2},
		{3},
	}

	hashes := bs.hashes()
	if !daghash.AreEqual(hashes, expected) {
		t.Errorf("TestHashes: hashes order is %s but expected %s", hashes, expected)
	}
}
func TestBlockSetHighest(t *testing.T) {
	node1 := &blockNode{hash: &daghash.Hash{10}, height: 1}
	node2a := &blockNode{hash: &daghash.Hash{20}, height: 2}
	node2b := &blockNode{hash: &daghash.Hash{21}, height: 2}
	node3 := &blockNode{hash: &daghash.Hash{30}, height: 3}

	tests := []struct {
		name            string
		set             blockSet
		expectedHighest *blockNode
	}{
		{
			name:            "empty set",
			set:             setFromSlice(),
			expectedHighest: nil,
		},
		{
			name:            "set with one member",
			set:             setFromSlice(node1),
			expectedHighest: node1,
		},
		{
			name:            "same-height highest members in set",
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
			t.Errorf("blockSet.highest: unexpected value in test '%s'. "+
				"Expected: %v, got: %v", test.name, test.expectedHighest, highest)
		}
	}
}

func TestBlockSetSubtract(t *testing.T) {
	node1 := &blockNode{hash: &daghash.Hash{10}}
	node2 := &blockNode{hash: &daghash.Hash{20}}
	node3 := &blockNode{hash: &daghash.Hash{30}}

	tests := []struct {
		name           string
		setA           blockSet
		setB           blockSet
		expectedResult blockSet
	}{
		{
			name:           "both sets empty",
			setA:           setFromSlice(),
			setB:           setFromSlice(),
			expectedResult: setFromSlice(),
		},
		{
			name:           "subtract an empty set",
			setA:           setFromSlice(node1),
			setB:           setFromSlice(),
			expectedResult: setFromSlice(node1),
		},
		{
			name:           "subtract from empty set",
			setA:           setFromSlice(),
			setB:           setFromSlice(node1),
			expectedResult: setFromSlice(),
		},
		{
			name:           "subtract unrelated set",
			setA:           setFromSlice(node1),
			setB:           setFromSlice(node2),
			expectedResult: setFromSlice(node1),
		},
		{
			name:           "typical case",
			setA:           setFromSlice(node1, node2),
			setB:           setFromSlice(node2, node3),
			expectedResult: setFromSlice(node1),
		},
	}

	for _, test := range tests {
		result := test.setA.subtract(test.setB)
		if !reflect.DeepEqual(result, test.expectedResult) {
			t.Errorf("blockSet.subtract: unexpected result in test '%s'. "+
				"Expected: %v, got: %v", test.name, test.expectedResult, result)
		}
	}
}

func TestBlockSetAddSet(t *testing.T) {
	node1 := &blockNode{hash: &daghash.Hash{10}}
	node2 := &blockNode{hash: &daghash.Hash{20}}
	node3 := &blockNode{hash: &daghash.Hash{30}}

	tests := []struct {
		name           string
		setA           blockSet
		setB           blockSet
		expectedResult blockSet
	}{
		{
			name:           "both sets empty",
			setA:           setFromSlice(),
			setB:           setFromSlice(),
			expectedResult: setFromSlice(),
		},
		{
			name:           "add an empty set",
			setA:           setFromSlice(node1),
			setB:           setFromSlice(),
			expectedResult: setFromSlice(node1),
		},
		{
			name:           "add to empty set",
			setA:           setFromSlice(),
			setB:           setFromSlice(node1),
			expectedResult: setFromSlice(node1),
		},
		{
			name:           "add already added member",
			setA:           setFromSlice(node1, node2),
			setB:           setFromSlice(node1),
			expectedResult: setFromSlice(node1, node2),
		},
		{
			name:           "typical case",
			setA:           setFromSlice(node1, node2),
			setB:           setFromSlice(node2, node3),
			expectedResult: setFromSlice(node1, node2, node3),
		},
	}

	for _, test := range tests {
		test.setA.addSet(test.setB)
		if !reflect.DeepEqual(test.setA, test.expectedResult) {
			t.Errorf("blockSet.addSet: unexpected result in test '%s'. "+
				"Expected: %v, got: %v", test.name, test.expectedResult, test.setA)
		}
	}
}

func TestBlockSetAddSlice(t *testing.T) {
	node1 := &blockNode{hash: &daghash.Hash{10}}
	node2 := &blockNode{hash: &daghash.Hash{20}}
	node3 := &blockNode{hash: &daghash.Hash{30}}

	tests := []struct {
		name           string
		set            blockSet
		slice          []*blockNode
		expectedResult blockSet
	}{
		{
			name:           "add empty slice to empty set",
			set:            setFromSlice(),
			slice:          []*blockNode{},
			expectedResult: setFromSlice(),
		},
		{
			name:           "add an empty slice",
			set:            setFromSlice(node1),
			slice:          []*blockNode{},
			expectedResult: setFromSlice(node1),
		},
		{
			name:           "add to empty set",
			set:            setFromSlice(),
			slice:          []*blockNode{node1},
			expectedResult: setFromSlice(node1),
		},
		{
			name:           "add already added member",
			set:            setFromSlice(node1, node2),
			slice:          []*blockNode{node1},
			expectedResult: setFromSlice(node1, node2),
		},
		{
			name:           "typical case",
			set:            setFromSlice(node1, node2),
			slice:          []*blockNode{node2, node3},
			expectedResult: setFromSlice(node1, node2, node3),
		},
	}

	for _, test := range tests {
		test.set.addSlice(test.slice)
		if !reflect.DeepEqual(test.set, test.expectedResult) {
			t.Errorf("blockSet.addSlice: unexpected result in test '%s'. "+
				"Expected: %v, got: %v", test.name, test.expectedResult, test.set)
		}
	}
}

func TestBlockSetUnion(t *testing.T) {
	node1 := &blockNode{hash: &daghash.Hash{10}}
	node2 := &blockNode{hash: &daghash.Hash{20}}
	node3 := &blockNode{hash: &daghash.Hash{30}}

	tests := []struct {
		name           string
		setA           blockSet
		setB           blockSet
		expectedResult blockSet
	}{
		{
			name:           "both sets empty",
			setA:           setFromSlice(),
			setB:           setFromSlice(),
			expectedResult: setFromSlice(),
		},
		{
			name:           "union against an empty set",
			setA:           setFromSlice(node1),
			setB:           setFromSlice(),
			expectedResult: setFromSlice(node1),
		},
		{
			name:           "union from an empty set",
			setA:           setFromSlice(),
			setB:           setFromSlice(node1),
			expectedResult: setFromSlice(node1),
		},
		{
			name:           "union with subset",
			setA:           setFromSlice(node1, node2),
			setB:           setFromSlice(node1),
			expectedResult: setFromSlice(node1, node2),
		},
		{
			name:           "typical case",
			setA:           setFromSlice(node1, node2),
			setB:           setFromSlice(node2, node3),
			expectedResult: setFromSlice(node1, node2, node3),
		},
	}

	for _, test := range tests {
		result := test.setA.union(test.setB)
		if !reflect.DeepEqual(result, test.expectedResult) {
			t.Errorf("blockSet.union: unexpected result in test '%s'. "+
				"Expected: %v, got: %v", test.name, test.expectedResult, result)
		}
	}
}

func TestBlockSetHashesEqual(t *testing.T) {
	node1 := &blockNode{hash: &daghash.Hash{10}}
	node2 := &blockNode{hash: &daghash.Hash{20}}

	tests := []struct {
		name           string
		set            blockSet
		hashes         []*daghash.Hash
		expectedResult bool
	}{
		{
			name:           "empty set, no hashes",
			set:            setFromSlice(),
			hashes:         []*daghash.Hash{},
			expectedResult: true,
		},
		{
			name:           "empty set, one hash",
			set:            setFromSlice(),
			hashes:         []*daghash.Hash{node1.hash},
			expectedResult: false,
		},
		{
			name:           "set and hashes of different length",
			set:            setFromSlice(node1, node2),
			hashes:         []*daghash.Hash{node1.hash},
			expectedResult: false,
		},
		{
			name:           "set equal to hashes",
			set:            setFromSlice(node1, node2),
			hashes:         []*daghash.Hash{node1.hash, node2.hash},
			expectedResult: true,
		},
		{
			name:           "set equal to hashes, different order",
			set:            setFromSlice(node1, node2),
			hashes:         []*daghash.Hash{node2.hash, node1.hash},
			expectedResult: true,
		},
	}

	for _, test := range tests {
		result := test.set.hashesEqual(test.hashes)
		if result != test.expectedResult {
			t.Errorf("blockSet.hashesEqual: unexpected result in test '%s'. "+
				"Expected: %t, got: %t", test.name, test.expectedResult, result)
		}
	}
}
