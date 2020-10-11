package ghostdagmanager

import "github.com/kaspanet/kaspad/domain/consensus/model"

func (gm *GHOSTDAGManager) bluest(blockHashes []*model.DomainHash) *model.DomainHash {
	var bluestBlockHash *model.DomainHash
	for _, hash := range blockHashes {
		if bluestBlockHash == nil || gm.less(bluestBlockHash, hash) {
			bluestBlockHash = hash
		}
	}
	return bluestBlockHash
}

func (gm *GHOSTDAGManager) less(blockA, blockB *model.DomainHash) bool {
	blockABlueScore := gm.ghostdagDataStore.Get(gm.databaseContext, blockA).BlueScore
	blockBBlueScore := gm.ghostdagDataStore.Get(gm.databaseContext, blockB).BlueScore
	if blockABlueScore == blockBBlueScore {
		return hashesLess(blockA, blockB)
	}
	return blockABlueScore < blockBBlueScore
}

func hashesLess(a, b *model.DomainHash) bool {
	// We compare the hashes backwards because Hash is stored as a little endian byte array.
	for i := len(a) - 1; i >= 0; i-- {
		switch {
		case a[i] < b[i]:
			return true
		case a[i] > b[i]:
			return false
		}
	}
	return false
}
