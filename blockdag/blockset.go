package blockdag

import (
	"sort"
	"strings"

	"github.com/daglabs/btcd/dagconfig/daghash"
)

// blockSet implements a basic unsorted set of blocks
type blockSet map[daghash.Hash]*blockNode

// newSet creates a new, empty BlockSet
func newSet() blockSet {
	return map[daghash.Hash]*blockNode{}
}

// setFromSlice converts a slice of blocks into an unordered set represented as map
func setFromSlice(blocks ...*blockNode) blockSet {
	set := newSet()
	for _, block := range blocks {
		set[block.hash] = block
	}
	return set
}

// maxHeight returns the height of the highest block in the block set
func (bs blockSet) maxHeight() int32 {
	var maxHeight int32
	for _, node := range bs {
		if maxHeight < node.height {
			maxHeight = node.height
		}
	}
	return maxHeight
}

func (bs blockSet) highest() *blockNode {
	var highest *blockNode
	for _, node := range bs {
		if highest == nil || highest.height < node.height || daghash.Less(&node.hash, &highest.hash) {
			highest = node
		}
	}
	return highest
}

// add adds a block to this BlockSet
func (bs blockSet) add(block *blockNode) {
	bs[block.hash] = block
}

// remove removes a block from this BlockSet, if exists
// Does nothing if this set does not contain the block
func (bs blockSet) remove(block *blockNode) {
	delete(bs, block.hash)
}

// clone clones thie block set
func (bs blockSet) clone() blockSet {
	clone := newSet()
	for _, block := range bs {
		clone.add(block)
	}
	return clone
}

// subtract returns the difference between the BlockSet and another BlockSet
func (bs blockSet) subtract(other blockSet) blockSet {
	diff := newSet()
	for _, block := range bs {
		if !other.contains(block) {
			diff.add(block)
		}
	}
	return diff
}

// addSet adds all blocks in other set to this set
func (bs blockSet) addSet(other blockSet) {
	for _, block := range other {
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
	_, ok := bs[block.hash]
	return ok
}

// hashesEqual returns true if the given hashes are equal to the hashes
// of the blocks in this set.
// NOTE: The given hash slice must not contain duplicates.
func (bs blockSet) hashesEqual(hashes []daghash.Hash) bool {
	if len(hashes) != len(bs) {
		return false
	}

	for _, hash := range hashes {
		if _, wasFound := bs[hash]; !wasFound {
			return false
		}
	}

	return true
}

// hashes returns the hashes of the blocks in this set.
func (bs blockSet) hashes() []daghash.Hash {
	hashes := make([]daghash.Hash, 0, len(bs))
	for hash := range bs {
		hashes = append(hashes, hash)
	}
	sort.Slice(hashes, func(i, j int) bool {
		return daghash.Less(&hashes[i], &hashes[j])
	})
	return hashes
}

func (bs blockSet) String() string {
	nodeStrs := make([]string, 0, len(bs))
	for _, node := range bs {
		nodeStrs = append(nodeStrs, node.String())
	}
	return strings.Join(nodeStrs, ",")
}

// anyChildInSet returns true iff any child of block is contained within this set
func (bs blockSet) anyChildInSet(block *blockNode) bool {
	for _, child := range block.children {
		if bs.contains(child) {
			return true
		}
	}

	return false
}

func (bs blockSet) bluest() *blockNode {
	var bluestNode *blockNode
	var maxScore uint64
	for _, node := range bs {
		if bluestNode == nil ||
			node.blueScore > maxScore ||
			(node.blueScore == maxScore && daghash.Less(&node.hash, &bluestNode.hash)) {
			bluestNode = node
			maxScore = node.blueScore
		}
	}
	return bluestNode
}
