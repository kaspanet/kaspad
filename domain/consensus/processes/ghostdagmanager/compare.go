package ghostdagmanager

import "github.com/kaspanet/kaspad/domain/consensus/model"

func (gm *ghostdagManager) findSelectedParent(parentHashes []*model.DomainHash) (*model.DomainHash, error) {
	var selectedParent *model.DomainHash
	for _, hash := range parentHashes {
		if selectedParent == nil {
			selectedParent = hash
			continue
		}
		isHashBiggerThanSelectedParent, err := gm.less(selectedParent, hash)
		if err != nil {
			return nil, err
		}
		if isHashBiggerThanSelectedParent {
			selectedParent = hash
		}
	}
	return selectedParent, nil
}

func (gm *ghostdagManager) less(blockA, blockB *model.DomainHash) (bool, error) {
	blockAGHOSTDAGData, err := gm.ghostdagDataStore.Get(gm.databaseContext, blockA)
	if err != nil {
		return false, err
	}
	blockBGHOSTDAGData, err := gm.ghostdagDataStore.Get(gm.databaseContext, blockB)
	if err != nil {
		return false, err
	}
	chosenSelectedParent := gm.ChooseSelectedParent(blockA, blockAGHOSTDAGData, blockB, blockBGHOSTDAGData)
	return chosenSelectedParent == blockB, nil
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
