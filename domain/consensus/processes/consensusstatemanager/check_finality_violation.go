package consensusstatemanager

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

func (csm *consensusStateManager) isViolatingFinality(blockHash *externalapi.DomainHash) (isViolatingFinality bool,
	shouldSendNotification bool, err error) {

	log.Debugf("isViolatingFinality start for block %s", blockHash)
	defer log.Debugf("isViolatingFinality end for block %s", blockHash)

	if blockHash.Equal(csm.genesisHash) {
		log.Debugf("Block %s is the genesis block, "+
			"and does not violate finality by definition", blockHash)
		return false, false, nil
	}

	var finalityPoint *externalapi.DomainHash
	virtualFinalityPoint, err := csm.finalityManager.VirtualFinalityPoint()
	if err != nil {
		return false, false, err
	}
	log.Debugf("The virtual finality point is: %s", virtualFinalityPoint)

	// There can be a situation where the virtual points close to the pruning point (or even in the past
	// of the pruning point before calling validateAndInsertBlock for the pruning point block) and the
	// finality point from the virtual point-of-view is in the past of the pruning point.
	// In such situation we override the finality point to be the pruning point to avoid situations where
	// the virtual selected parent chain don't include the pruning point.
	pruningPoint, err := csm.pruningStore.PruningPoint(csm.databaseContext)
	if err != nil {
		return false, false, err
	}
	log.Debugf("The pruning point is: %s", pruningPoint)

	isFinalityPointInPastOfPruningPoint, err := csm.dagTopologyManager.IsAncestorOf(virtualFinalityPoint, pruningPoint)
	if err != nil {
		return false, false, err
	}

	if !isFinalityPointInPastOfPruningPoint {
		finalityPoint = virtualFinalityPoint
	} else {
		log.Debugf("The virtual finality point is %s in the past of the pruning point, so finality is validated "+
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
		// On IBD it's pretty normal to get blocks in the anticone of the pruning
		// point, so we don't notify on cases when the pruning point is in the future
		// of the finality point.
		return true, false, nil
	}
	log.Debugf("Block %s does not violate finality", blockHash)

	return false, false, nil
}
