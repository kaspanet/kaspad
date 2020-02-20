package blockdag

import (
	"strings"

	"github.com/kaspanet/kaspad/util/daghash"
)

// blockSet implements a basic unsorted set of blocks
type blockSet map[*blockNode]struct{}

// newBlockSet creates a new, empty BlockSet
func newBlockSet() blockSet {
	return map[*blockNode]struct{}{}
}

// blockSetFromSlice converts a slice of blockNodes into an unordered set represented as map
func blockSetFromSlice(nodes ...*blockNode) blockSet {
	set := newBlockSet()
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
	clone := newBlockSet()
	for node := range bs {
		clone.add(node)
	}
	return clone
}

// subtract returns the difference between the BlockSet and another BlockSet
func (bs blockSet) subtract(other blockSet) blockSet {
	diff := newBlockSet()
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
