package externalapi

// BlockLevelParents represent the parents within a single super-block level
// See https://github.com/kaspanet/research/issues/3 for further details
type BlockLevelParents []*DomainHash

// Equal returns true if this BlockLevelParents is equal to `other`
func (sl BlockLevelParents) Equal(other BlockLevelParents) bool {
	if len(sl) != len(other) {
		return false
	}
	for _, thisHash := range sl {
		found := false
		for _, otherHash := range other {
			if thisHash.Equal(otherHash) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Clone creates a clone of this BlockLevelParents
func (sl BlockLevelParents) Clone() BlockLevelParents {
	return CloneHashes(sl)
}

// Contains returns true if this BlockLevelParents contains the given blockHash
func (sl BlockLevelParents) Contains(blockHash *DomainHash) bool {
	for _, blockLevelParent := range sl {
		if blockLevelParent.Equal(blockHash) {
			return true
		}
	}
	return false
}

// ParentsEqual returns true if all the BlockLevelParents in `a` and `b` are
// equal pairwise
func ParentsEqual(a, b []BlockLevelParents) bool {
	if len(a) != len(b) {
		return false
	}
	for i, blockLevelParents := range a {
		if !blockLevelParents.Equal(b[i]) {
			return false
		}
	}
	return true
}

// CloneParents creates a clone of the given BlockLevelParents slice
func CloneParents(parents []BlockLevelParents) []BlockLevelParents {
	clone := make([]BlockLevelParents, len(parents))
	for i, blockLevelParents := range parents {
		clone[i] = blockLevelParents.Clone()
	}
	return clone
}
