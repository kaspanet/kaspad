package blocknode

import (
	"reflect"
	"testing"

	"github.com/kaspanet/kaspad/util/daghash"
)

func TestHashes(t *testing.T) {
	bs := SetFromSlice(
		&Node{
			Hash: &daghash.Hash{3},
		},
		&Node{
			Hash: &daghash.Hash{1},
		},
		&Node{
			Hash: &daghash.Hash{0},
		},
		&Node{
			Hash: &daghash.Hash{2},
		},
	)

	expected := []*daghash.Hash{
		{0},
		{1},
		{2},
		{3},
	}

	hashes := bs.Hashes()
	if !daghash.AreEqual(hashes, expected) {
		t.Errorf("TestHashes: Hashes order is %s but expected %s", hashes, expected)
	}
}

func TestBlockSetSubtract(t *testing.T) {
	node1 := &Node{Hash: &daghash.Hash{10}}
	node2 := &Node{Hash: &daghash.Hash{20}}
	node3 := &Node{Hash: &daghash.Hash{30}}

	tests := []struct {
		name           string
		setA           Set
		setB           Set
		expectedResult Set
	}{
		{
			name:           "both sets empty",
			setA:           SetFromSlice(),
			setB:           SetFromSlice(),
			expectedResult: SetFromSlice(),
		},
		{
			name:           "Subtract an empty set",
			setA:           SetFromSlice(node1),
			setB:           SetFromSlice(),
			expectedResult: SetFromSlice(node1),
		},
		{
			name:           "Subtract from empty set",
			setA:           SetFromSlice(),
			setB:           SetFromSlice(node1),
			expectedResult: SetFromSlice(),
		},
		{
			name:           "Subtract unrelated set",
			setA:           SetFromSlice(node1),
			setB:           SetFromSlice(node2),
			expectedResult: SetFromSlice(node1),
		},
		{
			name:           "typical case",
			setA:           SetFromSlice(node1, node2),
			setB:           SetFromSlice(node2, node3),
			expectedResult: SetFromSlice(node1),
		},
	}

	for _, test := range tests {
		result := test.setA.Subtract(test.setB)
		if !reflect.DeepEqual(result, test.expectedResult) {
			t.Errorf("Set.Subtract: unexpected result in test '%s'. "+
				"Expected: %v, got: %v", test.name, test.expectedResult, result)
		}
	}
}

func TestBlockSetAddSet(t *testing.T) {
	node1 := &Node{Hash: &daghash.Hash{10}}
	node2 := &Node{Hash: &daghash.Hash{20}}
	node3 := &Node{Hash: &daghash.Hash{30}}

	tests := []struct {
		name           string
		setA           Set
		setB           Set
		expectedResult Set
	}{
		{
			name:           "both sets empty",
			setA:           SetFromSlice(),
			setB:           SetFromSlice(),
			expectedResult: SetFromSlice(),
		},
		{
			name:           "Add an empty set",
			setA:           SetFromSlice(node1),
			setB:           SetFromSlice(),
			expectedResult: SetFromSlice(node1),
		},
		{
			name:           "Add to empty set",
			setA:           SetFromSlice(),
			setB:           SetFromSlice(node1),
			expectedResult: SetFromSlice(node1),
		},
		{
			name:           "Add already added member",
			setA:           SetFromSlice(node1, node2),
			setB:           SetFromSlice(node1),
			expectedResult: SetFromSlice(node1, node2),
		},
		{
			name:           "typical case",
			setA:           SetFromSlice(node1, node2),
			setB:           SetFromSlice(node2, node3),
			expectedResult: SetFromSlice(node1, node2, node3),
		},
	}

	for _, test := range tests {
		test.setA.addSet(test.setB)
		if !reflect.DeepEqual(test.setA, test.expectedResult) {
			t.Errorf("Set.addSet: unexpected result in test '%s'. "+
				"Expected: %v, got: %v", test.name, test.expectedResult, test.setA)
		}
	}
}

func TestBlockSetAddSlice(t *testing.T) {
	node1 := &Node{Hash: &daghash.Hash{10}}
	node2 := &Node{Hash: &daghash.Hash{20}}
	node3 := &Node{Hash: &daghash.Hash{30}}

	tests := []struct {
		name           string
		set            Set
		slice          []*Node
		expectedResult Set
	}{
		{
			name:           "Add empty slice to empty set",
			set:            SetFromSlice(),
			slice:          []*Node{},
			expectedResult: SetFromSlice(),
		},
		{
			name:           "Add an empty slice",
			set:            SetFromSlice(node1),
			slice:          []*Node{},
			expectedResult: SetFromSlice(node1),
		},
		{
			name:           "Add to empty set",
			set:            SetFromSlice(),
			slice:          []*Node{node1},
			expectedResult: SetFromSlice(node1),
		},
		{
			name:           "Add already added member",
			set:            SetFromSlice(node1, node2),
			slice:          []*Node{node1},
			expectedResult: SetFromSlice(node1, node2),
		},
		{
			name:           "typical case",
			set:            SetFromSlice(node1, node2),
			slice:          []*Node{node2, node3},
			expectedResult: SetFromSlice(node1, node2, node3),
		},
	}

	for _, test := range tests {
		test.set.addSlice(test.slice)
		if !reflect.DeepEqual(test.set, test.expectedResult) {
			t.Errorf("Set.addSlice: unexpected result in test '%s'. "+
				"Expected: %v, got: %v", test.name, test.expectedResult, test.set)
		}
	}
}

func TestBlockSetUnion(t *testing.T) {
	node1 := &Node{Hash: &daghash.Hash{10}}
	node2 := &Node{Hash: &daghash.Hash{20}}
	node3 := &Node{Hash: &daghash.Hash{30}}

	tests := []struct {
		name           string
		setA           Set
		setB           Set
		expectedResult Set
	}{
		{
			name:           "both sets empty",
			setA:           SetFromSlice(),
			setB:           SetFromSlice(),
			expectedResult: SetFromSlice(),
		},
		{
			name:           "union against an empty set",
			setA:           SetFromSlice(node1),
			setB:           SetFromSlice(),
			expectedResult: SetFromSlice(node1),
		},
		{
			name:           "union from an empty set",
			setA:           SetFromSlice(),
			setB:           SetFromSlice(node1),
			expectedResult: SetFromSlice(node1),
		},
		{
			name:           "union with subset",
			setA:           SetFromSlice(node1, node2),
			setB:           SetFromSlice(node1),
			expectedResult: SetFromSlice(node1, node2),
		},
		{
			name:           "typical case",
			setA:           SetFromSlice(node1, node2),
			setB:           SetFromSlice(node2, node3),
			expectedResult: SetFromSlice(node1, node2, node3),
		},
	}

	for _, test := range tests {
		result := test.setA.union(test.setB)
		if !reflect.DeepEqual(result, test.expectedResult) {
			t.Errorf("Set.union: unexpected result in test '%s'. "+
				"Expected: %v, got: %v", test.name, test.expectedResult, result)
		}
	}
}

func TestBlockSetAreAllIn(t *testing.T) {
	node1 := &Node{Hash: &daghash.Hash{10}}
	node2 := &Node{Hash: &daghash.Hash{20}}
	node3 := &Node{Hash: &daghash.Hash{30}}

	tests := []struct {
		name           string
		set            Set
		other          Set
		expectedResult bool
	}{
		{
			name:           "two empty sets",
			set:            SetFromSlice(),
			other:          SetFromSlice(),
			expectedResult: true,
		},
		{
			name:           "set empty, other full",
			set:            SetFromSlice(),
			other:          SetFromSlice(node1, node2, node3),
			expectedResult: true,
		},
		{
			name:           "set full, other empty",
			set:            SetFromSlice(node1, node2, node3),
			other:          SetFromSlice(),
			expectedResult: false,
		},
		{
			name:           "same node in both",
			set:            SetFromSlice(node1),
			other:          SetFromSlice(node1),
			expectedResult: true,
		},
		{
			name:           "different node in both",
			set:            SetFromSlice(node1),
			other:          SetFromSlice(node2),
			expectedResult: false,
		},
		{
			name:           "set is subset of other",
			set:            SetFromSlice(node1, node2),
			other:          SetFromSlice(node2, node1, node3),
			expectedResult: true,
		},
		{
			name:           "other is subset of set",
			set:            SetFromSlice(node2, node1, node3),
			other:          SetFromSlice(node1, node2),
			expectedResult: false,
		},
	}

	for _, test := range tests {
		result := test.set.AreAllIn(test.other)

		if result != test.expectedResult {
			t.Errorf("Set.AreAllIn: unexpected result in test '%s'. "+
				"Expected: '%t', got: '%t'", test.name, test.expectedResult, result)
		}
	}
}
