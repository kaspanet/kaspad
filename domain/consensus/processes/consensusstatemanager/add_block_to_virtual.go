package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// AddBlockToVirtual submits the given block to be added to the
// current virtual. This process may result in a new virtual block
// getting created
func (csm *consensusStateManager) AddBlockToVirtual(blockHash *externalapi.DomainHash) error {
	isNextVirtualSelectedParent, err := csm.isNextVirtualSelectedParent(blockHash)
	if err != nil {
		return err
	}

	if isNextVirtualSelectedParent {
		blockStatus, err := csm.resolveBlockStatus(blockHash)
		if err != nil {
			return err
		}
		if blockStatus == model.StatusValid {
			err = csm.checkFinalityViolation(blockHash)
			if err != nil {
				return err
			}

			err = csm.reachabilityManager.UpdateReindexRoot(blockHash)
			if err != nil {
				return err
			}
		}
	}

	newTips, err := csm.addTip(blockHash)
	if err != nil {
		return err
	}

	err = csm.updateVirtual(blockHash, newTips)
	if err != nil {
		return err
	}

	return nil
}

func (csm *consensusStateManager) isNextVirtualSelectedParent(blockHash *externalapi.DomainHash) (bool, error) {
	virtualGhostdagData, err := csm.ghostdagDataStore.Get(csm.databaseContext, model.VirtualBlockHash)
	if err != nil {
		return false, err
	}

	nextVirtualSelectedParent, err := csm.ghostdagManager.ChooseSelectedParent(virtualGhostdagData.SelectedParent, blockHash)
	if err != nil {
		return false, err
	}

	return *blockHash == *nextVirtualSelectedParent, nil
}

func (csm *consensusStateManager) addTip(newTipHash *externalapi.DomainHash) (newTips []*externalapi.DomainHash, err error) {
	currentTips, err := csm.consensusStateStore.Tips(csm.databaseContext)
	if err != nil {
		return nil, err
	}

	newTipParents, err := csm.dagTopologyManager.Parents(newTipHash)
	if err != nil {
		return nil, err
	}

	newTips = []*externalapi.DomainHash{newTipHash}

	for _, currentTip := range currentTips {
		isCurrentTipInNewTipParents := false
		for _, newTipParent := range newTipParents {
			if *currentTip == *newTipParent {
				isCurrentTipInNewTipParents = true
				break
			}
		}
		if !isCurrentTipInNewTipParents {
			newTips = append(newTips, currentTip)
		}
	}

	err = csm.consensusStateStore.StageTips(newTips)
	if err != nil {
		return nil, err
	}

	return newTips, nil
}
