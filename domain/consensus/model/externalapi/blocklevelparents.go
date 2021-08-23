package externalapi

type BlockLevelParents []*DomainHash

func (sl BlockLevelParents) Equal(other BlockLevelParents) bool {
	return HashesEqual(sl, other)
}

func (sl BlockLevelParents) Clone() BlockLevelParents {
	return CloneHashes(sl)
}

func BlockLevelParentsEqual(a, b []BlockLevelParents) bool {
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

func CloneBlockLevelParents(blockLevelParents []BlockLevelParents) []BlockLevelParents {
	clone := make([]BlockLevelParents, len(blockLevelParents))
	for i, superblockLevel := range blockLevelParents {
		clone[i] = superblockLevel.Clone()
	}
	return clone
}
