package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// AddBlock submits the given block to be added to the
// current virtual. This process may result in a new virtual block
// getting created
func (csm *consensusStateManager) AddBlock(blockHash *externalapi.DomainHash) (*externalapi.SelectedParentChainChanges, error) {
	log.Tracef("AddBlock start for block %s", blockHash)
	defer log.Tracef("AddBlock end for block %s", blockHash)

	log.Tracef("Resolving whether the block %s is the next virtual selected parent", blockHash)
	isCandidateToBeNextVirtualSelectedParent, err := csm.isCandidateToBeNextVirtualSelectedParent(blockHash)
	if err != nil {
		return nil, err
	}

	if isCandidateToBeNextVirtualSelectedParent {
		// It's important to check for finality violation before resolving the block status, because the status of
		// blocks with a selected chain that doesn't contain the pruning point cannot be resolved because they will
		// eventually try to fetch UTXO diffs from the past of the pruning point.
		log.Tracef("Block %s is candidate to be the next virtual selected parent. Resolving whether it violates "+
			"finality", blockHash)
		isViolatingFinality, shouldNotify, err := csm.isViolatingFinality(blockHash)
		if err != nil {
			return nil, err
		}

		if shouldNotify {
			//TODO: Send finality conflict notification
			log.Warnf("Finality Violation Detected! Block %s violates finality!", blockHash)
		}

		if !isViolatingFinality {
			log.Tracef("Block %s doesn't violate finality. Resolving its block status", blockHash)
			blockStatus, err := csm.resolveBlockStatus(blockHash)
			if err != nil {
				return nil, err
			}

			log.Debugf("Block %s resolved to status `%s`", blockHash, blockStatus)
		}
	} else {
		log.Debugf("Block %s is not the next virtual selected parent, "+
			"therefore its status remains `%s`", blockHash, externalapi.StatusUTXOPendingVerification)
	}

	log.Tracef("Adding block %s to the DAG tips", blockHash)
	newTips, err := csm.addTip(blockHash)
	if err != nil {
		return nil, err
	}
	log.Tracef("After adding %s, the new tips are %s", blockHash, newTips)

	log.Tracef("Updating the virtual with the new tips")
	selectedParentChainChanges, err := csm.updateVirtual(blockHash, newTips)
	if err != nil {
		return nil, err
	}

	return selectedParentChainChanges, nil
}

func (csm *consensusStateManager) isCandidateToBeNextVirtualSelectedParent(blockHash *externalapi.DomainHash) (bool, error) {
	log.Tracef("isCandidateToBeNextVirtualSelectedParent start for block %s", blockHash)
	defer log.Tracef("isCandidateToBeNextVirtualSelectedParent end for block %s", blockHash)

	if blockHash.Equal(csm.genesisHash) {
		log.Tracef("Block %s is the genesis block, therefore it is "+
			"the selected parent by definition", blockHash)
		return true, nil
	}

	virtualGhostdagData, err := csm.ghostdagDataStore.Get(csm.databaseContext, model.VirtualBlockHash)
	if err != nil {
		return false, err
	}

	log.Tracef("Selecting the next selected parent between "+
		"the block %s the current selected parent %s", blockHash, virtualGhostdagData.SelectedParent())
	nextVirtualSelectedParent, err := csm.ghostdagManager.ChooseSelectedParent(virtualGhostdagData.SelectedParent(), blockHash)
	if err != nil {
		return false, err
	}
	log.Tracef("The next selected parent is: %s", nextVirtualSelectedParent)

	return blockHash.Equal(nextVirtualSelectedParent), nil
}

func (csm *consensusStateManager) addTip(newTipHash *externalapi.DomainHash) (newTips []*externalapi.DomainHash, err error) {
	log.Tracef("addTip start for new tip %s", newTipHash)
	defer log.Tracef("addTip end for new tip %s", newTipHash)

	log.Tracef("Calculating the new tips for new tip %s", newTipHash)
	newTips, err = csm.calculateNewTips(newTipHash)
	if err != nil {
		return nil, err
	}
	log.Tracef("The new tips are: %s", newTips)

	csm.consensusStateStore.StageTips(newTips)
	log.Tracef("Staged the new tips %s", newTips)

	return newTips, nil
}

func (csm *consensusStateManager) calculateNewTips(newTipHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	log.Tracef("calculateNewTips start for new tip %s", newTipHash)
	defer log.Tracef("calculateNewTips end for new tip %s", newTipHash)

	if newTipHash.Equal(csm.genesisHash) {
		log.Tracef("The new tip is the genesis block, therefore it is the only tip by definition")
		return []*externalapi.DomainHash{newTipHash}, nil
	}

	currentTips, err := csm.consensusStateStore.Tips(csm.databaseContext)
	if err != nil {
		return nil, err
	}
	log.Tracef("The current tips are: %s", currentTips)

	newTipParents, err := csm.dagTopologyManager.Parents(newTipHash)
	if err != nil {
		return nil, err
	}
	log.Tracef("The parents of the new tip are: %s", newTipParents)

	newTips := []*externalapi.DomainHash{newTipHash}

	for _, currentTip := range currentTips {
		isCurrentTipInNewTipParents := false
		for _, newTipParent := range newTipParents {
			if currentTip.Equal(newTipParent) {
				isCurrentTipInNewTipParents = true
				break
			}
		}
		if !isCurrentTipInNewTipParents {
			newTips = append(newTips, currentTip)
		}
	}
	log.Tracef("The calculated new tips are: %s", newTips)

	return newTips, nil
}
