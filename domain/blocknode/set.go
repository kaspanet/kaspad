package blocknode

import (
	"strings"

	"github.com/kaspanet/kaspad/util/daghash"
)

// Set implements a basic unsorted set of blocks
type Set map[*Node]struct{}

// NewSet creates a new, empty Set
func NewSet() Set {
	return map[*Node]struct{}{}
}

// SetFromSlice converts a slice of blockNodes into an unordered set represented as map
func SetFromSlice(nodes ...*Node) Set {
	set := NewSet()
	for _, node := range nodes {
		set.Add(node)
	}
	return set
}

// Add adds a blockNode to this Set
func (bs Set) Add(node *Node) {
	bs[node] = struct{}{}
}

// Remove removes a blockNode from this Set, if exists
// Does nothing if this set does not contain the blockNode
func (bs Set) Remove(node *Node) {
	delete(bs, node)
}

// Clone clones thie block set
func (bs Set) Clone() Set {
	clone := make(Set, len(bs))
	for node := range bs {
		clone.Add(node)
	}
	return clone
}

// Subtract returns the difference between the Set and another Set
func (bs Set) Subtract(other Set) Set {
	diff := NewSet()
	for node := range bs {
		if !other.Contains(node) {
			diff.Add(node)
		}
	}
	return diff
}

// addSet adds all blockNodes in other set to this set
func (bs Set) addSet(other Set) {
	for node := range other {
		bs.Add(node)
	}
}

// addSlice adds provided slice to this set
func (bs Set) addSlice(slice []*Node) {
	for _, node := range slice {
		bs.Add(node)
	}
}

// union returns a Set that contains all blockNodes included in this set,
// the other set, or both
func (bs Set) union(other Set) Set {
	union := bs.Clone()

	union.addSet(other)

	return union
}

// Contains returns true if this set contains the given node
func (bs Set) Contains(node *Node) bool {
	_, ok := bs[node]
	return ok
}

// Hashes returns the Hashes of the blockNodes in this set.
func (bs Set) Hashes() []*daghash.Hash {
	hashes := make([]*daghash.Hash, 0, len(bs))
	for node := range bs {
		hashes = append(hashes, node.Hash)
	}
	daghash.Sort(hashes)
	return hashes
}

func (bs Set) String() string {
	nodeStrs := make([]string, 0, len(bs))
	for node := range bs {
		nodeStrs = append(nodeStrs, node.String())
	}
	return strings.Join(nodeStrs, ",")
}

// Bluest returns the node with the maximum bluescore from the set
func (bs Set) Bluest() *Node {
	var bluestNode *Node
	for node := range bs {
		if bluestNode == nil || bluestNode.Less(node) {
			bluestNode = node
		}
	}
	return bluestNode
}

// IsEqual checks if both sets are equal
func (bs Set) IsEqual(other Set) bool {
	if len(bs) != len(other) {
		return false
	}

	for node := range bs {
		if !other.Contains(node) {
			return false
		}
	}

	return true
}

// AreAllIn checks if other set contains all nodes from the current set
func (bs Set) AreAllIn(other Set) bool {
	for node := range bs {
		if !other.Contains(node) {
			return false
		}
	}

	return true
}

// IsOnlyGenesis returns true if the only block in this Set is the genesis block
func (bs Set) IsOnlyGenesis() bool {
	if len(bs) != 1 {
		return false
	}
	for node := range bs {
		if node.IsGenesis() {
			return true
		}
	}

	return false
}
