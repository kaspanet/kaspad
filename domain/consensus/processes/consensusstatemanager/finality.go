package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func (csm *consensusStateManager) checkFinalityViolation(
	blockHash *externalapi.DomainHash) error {

	log.Tracef("checkFinalityViolation start for block %s", blockHash)
	defer log.Tracef("checkFinalityViolation end for block %s", blockHash)

	if *blockHash == *csm.genesisHash {
		log.Tracef("Block %s is the genesis block, "+
			"and does not violate finality by definition", blockHash)
		return nil
	}

	isViolatingFinality, err := csm.isViolatingFinality(blockHash)
	if err != nil {
		return err
	}

	if isViolatingFinality {
		csm.blockStatusStore.Stage(blockHash, externalapi.StatusUTXOPendingVerification)
		log.Warnf("Finality Violation Detected! Block %s violates finality!", blockHash)
		return nil
	}
	log.Tracef("Block %s does not violate finality", blockHash)

	return nil
}

func (csm *consensusStateManager) virtualFinalityPoint() (
	*externalapi.DomainHash, error) {

	log.Tracef("virtualFinalityPoint start")
	defer log.Tracef("virtualFinalityPoint end")

	virtualFinalityPoint, err := csm.dagTraversalManager.BlockAtDepth(
		model.VirtualBlockHash, csm.finalityDepth)
	if err != nil {
		return nil, err
	}
	log.Tracef("The current virtual finality block is: %s", virtualFinalityPoint)

	return virtualFinalityPoint, nil
}

func (csm *consensusStateManager) isViolatingFinality(
	blockHash *externalapi.DomainHash) (bool, error) {

	log.Tracef("isViolatingFinality start for block %s", blockHash)
	defer log.Tracef("isViolatingFinality end for block %s", blockHash)

	virtualFinalityPoint, err := csm.virtualFinalityPoint()
	if err != nil {
		return false, err
	}
	log.Tracef("The virtual finality point is: %s", virtualFinalityPoint)

	isInSelectedParentChain, err := csm.dagTopologyManager.IsInSelectedParentChainOf(virtualFinalityPoint, blockHash)
	if err != nil {
		return false, err
	}
	log.Tracef("Is the virtual finality point %s "+
		"in the selected parent chain of %s: %t", virtualFinalityPoint, blockHash, isInSelectedParentChain)

	return !isInSelectedParentChain, nil
}
