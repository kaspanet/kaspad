package ghostdagmanager

import "github.com/kaspanet/kaspad/domain/consensus/model"

func (gm *ghostdagManager) findSelectedParent(parentHashes []*model.DomainHash) *model.DomainHash {
	var selectedParent *model.DomainHash
	for _, hash := range parentHashes {
		if selectedParent == nil || gm.less(selectedParent, hash) {
			selectedParent = hash
		}
	}
	return selectedParent
}

func (gm *ghostdagManager) less(blockA, blockB *model.DomainHash) bool {
	blockAGHOSTDAGData := gm.ghostdagDataStore.Get(gm.databaseContext, blockA)
	blockBGHOSTDAGData := gm.ghostdagDataStore.Get(gm.databaseContext, blockB)
	chosenSelectedParent := gm.ChooseSelectedParent(blockA, blockAGHOSTDAGData, blockB, blockBGHOSTDAGData)
	return chosenSelectedParent == blockB
}

func (gm *ghostdagManager) ChooseSelectedParent(
	blockHashA *model.DomainHash, blockAGHOSTDAGData *model.BlockGHOSTDAGData,
	blockHashB *model.DomainHash, blockBGHOSTDAGData *model.BlockGHOSTDAGData) *model.DomainHash {

	blockABlueScore := blockAGHOSTDAGData.BlueScore
	blockBBlueScore := blockBGHOSTDAGData.BlueScore
	if blockABlueScore == blockBBlueScore {
		if hashesLess(blockHashA, blockHashB) {
			return blockHashB
		}
		return blockHashA
	}
	if blockABlueScore < blockBBlueScore {
		return blockHashB
	}
	return blockHashA
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
