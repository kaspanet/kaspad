package blockdag

import (
	"reflect"
	"testing"

	"github.com/kaspanet/kaspad/util/daghash"
)

func TestHashes(t *testing.T) {
	bs := blockSetFromSlice(
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
			setA:           blockSetFromSlice(),
			setB:           blockSetFromSlice(),
			expectedResult: blockSetFromSlice(),
		},
		{
			name:           "subtract an empty set",
			setA:           blockSetFromSlice(node1),
			setB:           blockSetFromSlice(),
			expectedResult: blockSetFromSlice(node1),
		},
		{
			name:           "subtract from empty set",
			setA:           blockSetFromSlice(),
			setB:           blockSetFromSlice(node1),
			expectedResult: blockSetFromSlice(),
		},
		{
			name:           "subtract unrelated set",
			setA:           blockSetFromSlice(node1),
			setB:           blockSetFromSlice(node2),
			expectedResult: blockSetFromSlice(node1),
		},
		{
			name:           "typical case",
			setA:           blockSetFromSlice(node1, node2),
			setB:           blockSetFromSlice(node2, node3),
			expectedResult: blockSetFromSlice(node1),
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
			setA:           blockSetFromSlice(),
			setB:           blockSetFromSlice(),
			expectedResult: blockSetFromSlice(),
		},
		{
			name:           "add an empty set",
			setA:           blockSetFromSlice(node1),
			setB:           blockSetFromSlice(),
			expectedResult: blockSetFromSlice(node1),
		},
		{
			name:           "add to empty set",
			setA:           blockSetFromSlice(),
			setB:           blockSetFromSlice(node1),
			expectedResult: blockSetFromSlice(node1),
		},
		{
			name:           "add already added member",
			setA:           blockSetFromSlice(node1, node2),
			setB:           blockSetFromSlice(node1),
			expectedResult: blockSetFromSlice(node1, node2),
		},
		{
			name:           "typical case",
			setA:           blockSetFromSlice(node1, node2),
			setB:           blockSetFromSlice(node2, node3),
			expectedResult: blockSetFromSlice(node1, node2, node3),
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
			set:            blockSetFromSlice(),
			slice:          []*blockNode{},
			expectedResult: blockSetFromSlice(),
		},
		{
			name:           "add an empty slice",
			set:            blockSetFromSlice(node1),
			slice:          []*blockNode{},
			expectedResult: blockSetFromSlice(node1),
		},
		{
			name:           "add to empty set",
			set:            blockSetFromSlice(),
			slice:          []*blockNode{node1},
			expectedResult: blockSetFromSlice(node1),
		},
		{
			name:           "add already added member",
			set:            blockSetFromSlice(node1, node2),
			slice:          []*blockNode{node1},
			expectedResult: blockSetFromSlice(node1, node2),
		},
		{
			name:           "typical case",
			set:            blockSetFromSlice(node1, node2),
			slice:          []*blockNode{node2, node3},
			expectedResult: blockSetFromSlice(node1, node2, node3),
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
			setA:           blockSetFromSlice(),
			setB:           blockSetFromSlice(),
			expectedResult: blockSetFromSlice(),
		},
		{
			name:           "union against an empty set",
			setA:           blockSetFromSlice(node1),
			setB:           blockSetFromSlice(),
			expectedResult: blockSetFromSlice(node1),
		},
		{
			name:           "union from an empty set",
			setA:           blockSetFromSlice(),
			setB:           blockSetFromSlice(node1),
			expectedResult: blockSetFromSlice(node1),
		},
		{
			name:           "union with subset",
			setA:           blockSetFromSlice(node1, node2),
			setB:           blockSetFromSlice(node1),
			expectedResult: blockSetFromSlice(node1, node2),
		},
		{
			name:           "typical case",
			setA:           blockSetFromSlice(node1, node2),
			setB:           blockSetFromSlice(node2, node3),
			expectedResult: blockSetFromSlice(node1, node2, node3),
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
