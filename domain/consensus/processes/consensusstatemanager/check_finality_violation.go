package consensusstatemanager

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

func (csm *consensusStateManager) isViolatingFinality(blockHash *externalapi.DomainHash) (isViolatingFinality bool,
	shouldSendNotification bool, err error) {

	log.Tracef("isViolatingFinality start for block %s", blockHash)
	defer log.Tracef("isViolatingFinality end for block %s", blockHash)

	if *blockHash == *csm.genesisHash {
		log.Tracef("Block %s is the genesis block, "+
			"and does not violate finality by definition", blockHash)
		return false, false, nil
	}

	var finalityPoint *externalapi.DomainHash
	virtualFinalityPoint, err := csm.finalityManager.VirtualFinalityPoint()
	if err != nil {
		return false, false, err
	}
	log.Tracef("The virtual finality point is: %s", virtualFinalityPoint)

	pruningPoint, err := csm.pruningStore.PruningPoint(csm.databaseContext)
	if err != nil {
		return false, false, err
	}
	log.Tracef("The pruning point is: %s", pruningPoint)

	isFinalityPointInPastOfPruningPoint, err := csm.dagTopologyManager.IsAncestorOf(virtualFinalityPoint, pruningPoint)
	if err != nil {
		return false, false, err
	}

	if !isFinalityPointInPastOfPruningPoint {
		finalityPoint = virtualFinalityPoint
	} else {
		log.Tracef("The virtual finality point is in the past of the pruning point, so finality is validated "+
			"using the pruning point", virtualFinalityPoint)
		finalityPoint = pruningPoint
	}

	isInSelectedParentChainOfFinalityPoint, err := csm.dagTopologyManager.IsInSelectedParentChainOf(finalityPoint,
		blockHash)
	if err != nil {
		return false, false, err
	}

	if !isInSelectedParentChainOfFinalityPoint {
		if !isFinalityPointInPastOfPruningPoint {
			return true, true, nil
		}
		return true, false, nil
	}
	log.Tracef("Block %s does not violate finality", blockHash)

	return false, false, nil
}
