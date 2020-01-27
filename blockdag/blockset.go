package blockdag

import (
	"strings"

	"github.com/kaspanet/kaspad/util/daghash"
)

// blockSet implements a basic unsorted set of blocks
type blockSet map[*blockNode]struct{}

// newSet creates a new, empty BlockSet
func newSet() blockSet {
	return map[*blockNode]struct{}{}
}

// setFromSlice converts a slice of blockNodes into an unordered set represented as map
func setFromSlice(nodes ...*blockNode) blockSet {
	set := newSet()
	for _, node := range nodes {
		set.add(node)
	}
	return set
}

// add adds a blockNode to this BlockSet
func (bs blockSet) add(node *blockNode) {
	bs[node] = struct{}{}
}

// remove removes a blockNode from this BlockSet, if exists
// Does nothing if this set does not contain the blockNode
func (bs blockSet) remove(node *blockNode) {
	delete(bs, node)
}

// clone clones thie block set
func (bs blockSet) clone() blockSet {
	clone := newSet()
	for node := range bs {
		clone.add(node)
	}
	return clone
}

// subtract returns the difference between the BlockSet and another BlockSet
func (bs blockSet) subtract(other blockSet) blockSet {
	diff := newSet()
	for node := range bs {
		if !other.contains(node) {
			diff.add(node)
		}
	}
	return diff
}

// addSet adds all blockNodes in other set to this set
func (bs blockSet) addSet(other blockSet) {
	for node := range other {
		bs.add(node)
	}
}

// addSlice adds provided slice to this set
func (bs blockSet) addSlice(slice []*blockNode) {
	for _, node := range slice {
		bs.add(node)
	}
}

// union returns a BlockSet that contains all blockNodes included in this set,
// the other set, or both
func (bs blockSet) union(other blockSet) blockSet {
	union := bs.clone()

	union.addSet(other)

	return union
}

// contains returns true iff this set contains node
func (bs blockSet) contains(node *blockNode) bool {
	_, ok := bs[node]
	return ok
}

// containsHash returns true iff this set contains a block hash
func (bs blockSet) containsHash(hash *daghash.Hash) bool {
	for node := range bs {
		if node.hash.IsEqual(hash) {
			return true
		}
	}
	return false
}

// hashesEqual returns true if the given hashes are equal to the hashes
// of the blockNodes in this set.
// NOTE: The given hash slice must not contain duplicates.
func (bs blockSet) hashesEqual(hashes []*daghash.Hash) bool {
	if len(hashes) != len(bs) {
		return false
	}

	for _, hash := range hashes {
		if contains := bs.containsHash(hash); !contains {
			return false
		}
	}

	return true
}

// hashes returns the hashes of the blockNodes in this set.
func (bs blockSet) hashes() []*daghash.Hash {
	hashes := make([]*daghash.Hash, 0, len(bs))
	for node := range bs {
		hashes = append(hashes, node.hash)
	}
	daghash.Sort(hashes)
	return hashes
}

func (bs blockSet) String() string {
	nodeStrs := make([]string, 0, len(bs))
	for node := range bs {
		nodeStrs = append(nodeStrs, node.String())
	}
	return strings.Join(nodeStrs, ",")
}

// anyChildInSet returns true iff any child of node is contained within this set
func (bs blockSet) anyChildInSet(node *blockNode) bool {
	for child := range node.children {
		if bs.contains(child) {
			return true
		}
	}

	return false
}

func (bs blockSet) bluest() *blockNode {
	var bluestNode *blockNode
	var maxScore uint64
	for node := range bs {
		if bluestNode == nil ||
			node.blueScore > maxScore ||
			(node.blueScore == maxScore && daghash.Less(node.hash, bluestNode.hash)) {
			bluestNode = node
			maxScore = node.blueScore
		}
	}
	return bluestNode
}
