package ghostdagmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
)

func (gm *ghostdagManager) findSelectedParent(parentHashes []*externalapi.DomainHash) (*externalapi.DomainHash, error) {
	var selectedParent *externalapi.DomainHash
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

func (gm *ghostdagManager) less(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	chosenSelectedParent, err := gm.ChooseSelectedParent([]*externalapi.DomainHash{blockHashA, blockHashB})
	if err != nil {
		return false, err
	}
	return chosenSelectedParent == blockHashB, nil
}

func (gm *ghostdagManager) ChooseSelectedParent(blockHashes ...*externalapi.DomainHash) (*externalapi.DomainHash, error) {
	selectedParent := blockHashes[0]
	selectedParentGHOSTDAGData, err := gm.ghostdagDataStore.Get(gm.databaseContext, selectedParent)
	if err != nil {
		return nil, err
	}
	for _, blockHash := range blockHashes {
		blockGHOSTDAGData, err := gm.ghostdagDataStore.Get(gm.databaseContext, blockHash)
		if err != nil {
			return nil, err
		}

		if gm.Less(selectedParent, selectedParentGHOSTDAGData, blockHash, blockGHOSTDAGData) {
			selectedParent = blockHash
			selectedParentGHOSTDAGData = blockGHOSTDAGData
		}
	}

	return selectedParent, nil
}

func (gm *ghostdagManager) Less(blockHashA *externalapi.DomainHash, ghostdagDataA *model.BlockGHOSTDAGData,
	blockHashB *externalapi.DomainHash, ghostdagDataB *model.BlockGHOSTDAGData) bool {

	blockBlueScoreA := ghostdagDataA.BlueScore
	blockBlueScoreB := ghostdagDataB.BlueScore
	if blockBlueScoreA == blockBlueScoreB {
		return hashes.Less(blockHashA, blockHashB)
	}

	return blockBlueScoreA < blockBlueScoreB
}
