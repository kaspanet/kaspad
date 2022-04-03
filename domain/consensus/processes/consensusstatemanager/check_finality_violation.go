package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func (csm *consensusStateManager) isViolatingFinality(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash,
) (isViolatingFinality bool, shouldSendNotification bool, err error) {

	log.Tracef("isViolatingFinality start for block %s", blockHash)
	defer log.Tracef("isViolatingFinality end for block %s", blockHash)

	if blockHash.Equal(csm.genesisHash) {
		log.Tracef("Block %s is the genesis block, "+
			"and does not violate finality by definition", blockHash)
		return false, false, nil
	}

	var finalityPoint *externalapi.DomainHash
	virtualFinalityPoint, err := csm.finalityManager.VirtualFinalityPoint(stagingArea)
	if err != nil {
		return false, false, err
	}
	log.Tracef("The virtual finality point is: %s", virtualFinalityPoint)

	// There can be a situation where the virtual points close to the pruning point (or even in the past
	// of the pruning point before calling validateAndInsertBlock for the pruning point block) and the
	// finality point from the virtual point-of-view is in the past of the pruning point.
	// In such situation we override the finality point to be the pruning point to avoid situations where
	// the virtual selected parent chain don't include the pruning point.
	pruningPoint, err := csm.pruningStore.PruningPoint(csm.databaseContext, stagingArea)
	if err != nil {
		return false, false, err
	}
	log.Tracef("The pruning point is: %s", pruningPoint)

	isFinalityPointInPastOfPruningPoint, err := csm.dagTopologyManager.IsAncestorOf(stagingArea, virtualFinalityPoint, pruningPoint)
	if err != nil {
		return false, false, err
	}

	if !isFinalityPointInPastOfPruningPoint {
		finalityPoint = virtualFinalityPoint
	} else {
		log.Tracef("The virtual finality point is %s in the past of the pruning point, so finality is validated "+
			"using the pruning point", virtualFinalityPoint)
		finalityPoint = pruningPoint
	}

	isInSelectedParentChainOfFinalityPoint, err :=
		csm.dagTopologyManager.IsInSelectedParentChainOf(stagingArea, finalityPoint, blockHash)
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
		log.Tracef("Block %s violates finality, but kaspad is currently doing IBD, so this is normal", blockHash)
		return true, false, nil
	}
	log.Tracef("Block %s does not violate finality", blockHash)

	return false, false, nil
}
