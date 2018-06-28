package blockdag

import (
	"strings"
	"github.com/daglabs/btcd/dagconfig/daghash"
)

// BlockSet implements a basic unsorted set of blocks
type BlockSet map[*blockNode]bool

// NewSet creates a new, empty BlockSet
func NewSet() BlockSet {
	return map[*blockNode]bool{}
}

// SetFromSlice converts a slice of blocks into an unordered set represented as map
func SetFromSlice(blocks ...*blockNode) BlockSet {
	set := NewSet()
	for _, block := range blocks {
		set[block] = true
	}
	return set
}

// ToSlice converts a set of blocks into a slice
func (bs BlockSet) ToSlice() []*blockNode {
	slice := []*blockNode{}

	for block := range bs {
		slice = append(slice, block)
	}

	return slice
}

// Add adds a block to this BlockSet
func (bs BlockSet) Add(block *blockNode) {
	bs[block] = true
}

// Remove removes a block from this BlockSet, if exists
func (bs BlockSet) Remove(block *blockNode) {
	delete(bs, block)
}

// Clone clones thie block set
func (bs BlockSet) Clone() BlockSet {
	clone := NewSet()
	for block := range bs {
		clone.Add(block)
	}
	return clone
}

// Subtract returns the difference between the BlockSet and another BlockSet
func (bs BlockSet) Subtract(other BlockSet) BlockSet {
	diff := NewSet()
	for block := range bs {
		if !other.Contains(block) {
			diff.Add(block)
		}
	}
	return diff
}

// AddSet adds all blocks in other set to this set
func (bs BlockSet) AddSet(other BlockSet) {
	for block := range other {
		bs.Add(block)
	}
}

// AddSlice adds provided slice to this set
func (bs BlockSet) AddSlice(slice []*blockNode) {
	for _, block := range slice {
		bs.Add(block)
	}
}

// Union returns a BlockSet that contains all blocks included in this set,
// the other set, or both
func (bs BlockSet) Union(other BlockSet) BlockSet {
	union := bs.Clone()

	union.AddSet(other)

	return union
}

// Contains returns true iff this set contains block
func (bs BlockSet) Contains(block *blockNode) bool {
	_, ok := bs[block]
	return ok
}

// HashesEqual returns true if the given hashes are equal to the hashes
// of the blocks in this set.
// NOTE: The given hash slice must not contain duplicates.
func (bs BlockSet) HashesEqual(hashes []daghash.Hash) bool {
	if len(hashes) != len(bs) {
		return false
	}

	for _, hash := range hashes {
		wasFound := false
		for node := range bs {
			if hash.IsEqual(&node.hash) {
				wasFound = true
				break
			}
		}

		if !wasFound {
			return false
		}
	}

	return true
}

// First returns the first block in this set or nil if this set is empty.
func (bs BlockSet) First() *blockNode {
	for block := range bs {
		return block
	}

	return nil
}

func (bs BlockSet) String() string {
	ids := []string{}
	for block := range bs {
		ids = append(ids, block.hash.String())
	}
	return strings.Join(ids, ",")
}
