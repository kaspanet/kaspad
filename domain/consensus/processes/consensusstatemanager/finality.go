package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func (csm *consensusStateManager) checkFinalityViolation(
	blockHash *externalapi.DomainHash) error {

	if *blockHash == *csm.genesisHash {
		return nil
	}

	isViolatingFinality, err := csm.isViolatingFinality(blockHash)
	if err != nil {
		return err
	}

	if isViolatingFinality {
		csm.blockStatusStore.Stage(blockHash, externalapi.StatusUTXOPendingVerification)
		log.Warnf("Finality Violation Detected! Block %s violates finality!", blockHash)
	}

	return nil
}

func (csm *consensusStateManager) virtualFinalityPoint() (
	*externalapi.DomainHash, error) {

	virtualGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	finalityPointBlueScore := virtualGHOSTDAGData.BlueScore - csm.finalityDepth
	if virtualGHOSTDAGData.BlueScore < csm.finalityDepth {
		// if there's no `csm.finalityDepth` blocks in the DAG
		// practically - returns the genesis
		finalityPointBlueScore = 0
	}

	return csm.dagTraversalManager.HighestChainBlockBelowBlueScore(
		model.VirtualBlockHash, finalityPointBlueScore)
}

func (csm *consensusStateManager) isViolatingFinality(
	blockHash *externalapi.DomainHash) (bool, error) {

	virtualFinalityPoint, err := csm.virtualFinalityPoint()
	if err != nil {
		return false, err
	}

	isInSelectedParentChain, err := csm.dagTopologyManager.IsInSelectedParentChainOf(virtualFinalityPoint, blockHash)
	if err != nil {
		return false, err
	}
	return !isInSelectedParentChain, nil
}
