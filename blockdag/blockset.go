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

// setFromSlice converts a slice of blocks into an unordered set represented as map
func setFromSlice(blocks ...*blockNode) blockSet {
	set := newSet()
	for _, block := range blocks {
		set.add(block)
	}
	return set
}

// add adds a block to this BlockSet
func (bs blockSet) add(block *blockNode) {
	bs[block] = struct{}{}
}

// remove removes a block from this BlockSet, if exists
// Does nothing if this set does not contain the block
func (bs blockSet) remove(block *blockNode) {
	delete(bs, block)
}

// clone clones thie block set
func (bs blockSet) clone() blockSet {
	clone := newSet()
	for block := range bs {
		clone.add(block)
	}
	return clone
}

// subtract returns the difference between the BlockSet and another BlockSet
func (bs blockSet) subtract(other blockSet) blockSet {
	diff := newSet()
	for block := range bs {
		if !other.contains(block) {
			diff.add(block)
		}
	}
	return diff
}

// addSet adds all blocks in other set to this set
func (bs blockSet) addSet(other blockSet) {
	for block := range other {
		bs.add(block)
	}
}

// addSlice adds provided slice to this set
func (bs blockSet) addSlice(slice []*blockNode) {
	for _, block := range slice {
		bs.add(block)
	}
}

// union returns a BlockSet that contains all blocks included in this set,
// the other set, or both
func (bs blockSet) union(other blockSet) blockSet {
	union := bs.clone()

	union.addSet(other)

	return union
}

// contains returns true iff this set contains block
func (bs blockSet) contains(block *blockNode) bool {
	_, ok := bs[block]
	return ok
}

// containsHash returns true iff this set contains a block hash
func (bs blockSet) containsHash(hash *daghash.Hash) bool {
	for block := range bs {
		if block.hash.IsEqual(hash) {
			return true
		}
	}
	return false
}

// hashesEqual returns true if the given hashes are equal to the hashes
// of the blocks in this set.
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

// hashes returns the hashes of the blocks in this set.
func (bs blockSet) hashes() []*daghash.Hash {
	hashes := make([]*daghash.Hash, 0, len(bs))
	for block := range bs {
		hashes = append(hashes, block.hash)
	}
	daghash.Sort(hashes)
	return hashes
}

func (bs blockSet) String() string {
	blockStrs := make([]string, 0, len(bs))
	for block := range bs {
		blockStrs = append(blockStrs, block.String())
	}
	return strings.Join(blockStrs, ",")
}

// anyChildInSet returns true iff any child of block is contained within this set
func (bs blockSet) anyChildInSet(block *blockNode) bool {
	for child := range block.children {
		if bs.contains(child) {
			return true
		}
	}

	return false
}

func (bs blockSet) bluest() *blockNode {
	var bluestBlock *blockNode
	var maxScore uint64
	for block := range bs {
		if bluestBlock == nil ||
			block.blueScore > maxScore ||
			(block.blueScore == maxScore && daghash.Less(block.hash, bluestBlock.hash)) {
			bluestBlock = block
			maxScore = block.blueScore
		}
	}
	return bluestBlock
}
