package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util/staging"
	"github.com/pkg/errors"
	"sort"
)

func (csm *consensusStateManager) tipsInDecreasingGHOSTDAGParentOrder(stagingArea *model.StagingArea) ([]*externalapi.DomainHash, error) {
	tips, err := csm.consensusStateStore.Tips(stagingArea, csm.databaseContext)
	if err != nil {
		return nil, err
	}

	var sortErr error
	sort.Slice(tips, func(i, j int) bool {
		selectedParent, err := csm.ghostdagManager.ChooseSelectedParent(stagingArea, tips[i], tips[j])
		if err != nil {
			sortErr = err
			return false
		}

		return selectedParent.Equal(tips[i])
	})
	if sortErr != nil {
		return nil, sortErr
	}
	return tips, nil
}

func (csm *consensusStateManager) findNextPendingTip(stagingArea *model.StagingArea) (*externalapi.DomainHash, externalapi.BlockStatus, error) {
	orderedTips, err := csm.tipsInDecreasingGHOSTDAGParentOrder(stagingArea)
	if err != nil {
		return nil, externalapi.StatusInvalid, err
	}

	for _, tip := range orderedTips {
		log.Debugf("Resolving tip %s", tip)
		isViolatingFinality, shouldNotify, err := csm.isViolatingFinality(stagingArea, tip)
		if err != nil {
			return nil, externalapi.StatusInvalid, err
		}

		if isViolatingFinality {
			if shouldNotify {
				//TODO: Send finality conflict notification
				log.Warnf("Skipping %s tip resolution because it violates finality", tip)
			}
			continue
		}

		status, err := csm.blockStatusStore.Get(csm.databaseContext, stagingArea, tip)
		if err != nil {
			return nil, externalapi.StatusInvalid, err
		}
		if status == externalapi.StatusUTXOValid || status == externalapi.StatusUTXOPendingVerification {
			return tip, status, nil
		}
	}

	return nil, externalapi.StatusInvalid, nil
}

// getGHOSTDAGLowerTips returns the set of tips which are lower in GHOSTDAG parent order than `pendingTip`. i.e.,
// they can be added to virtual parents but `pendingTip` will remain the virtual selected parent
func (csm *consensusStateManager) getGHOSTDAGLowerTips(stagingArea *model.StagingArea, pendingTip *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	tips, err := csm.consensusStateStore.Tips(stagingArea, csm.databaseContext)
	if err != nil {
		return nil, err
	}

	lowerTips := []*externalapi.DomainHash{pendingTip}
	for _, tip := range tips {
		if tip.Equal(pendingTip) {
			continue
		}
		selectedParent, err := csm.ghostdagManager.ChooseSelectedParent(stagingArea, tip, pendingTip)
		if err != nil {
			return nil, err
		}
		if selectedParent.Equal(pendingTip) {
			lowerTips = append(lowerTips, tip)
		}
	}
	return lowerTips, nil
}

func (csm *consensusStateManager) ResolveVirtual(maxBlocksToResolve uint64) (*externalapi.VirtualChangeSet, bool, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "csm.ResolveVirtual")
	defer onEnd()

	// We use a read-only staging area for some read-only actions, to avoid
	// confusion with the resolve/updateVirtual staging areas below
	readStagingArea := model.NewStagingArea()

	pendingTip, pendingTipStatus, err := csm.findNextPendingTip(readStagingArea)
	if err != nil {
		return nil, false, err
	}

	if pendingTip == nil {
		log.Warnf("None of the DAG tips are valid")
		return nil, true, nil
	}

	previousVirtualSelectedParent, err := csm.virtualSelectedParent(readStagingArea)
	if err != nil {
		return nil, false, err
	}

	if pendingTipStatus == externalapi.StatusUTXOValid && previousVirtualSelectedParent.Equal(pendingTip) {
		return nil, true, nil
	}

	// Resolve a chunk from the pending chain
	resolveStagingArea := model.NewStagingArea()
	unverifiedBlocks, err := csm.getUnverifiedChainBlocks(resolveStagingArea, pendingTip)
	if err != nil {
		return nil, false, err
	}

	// Initially set the resolve processing point to the pending tip
	processingPoint := pendingTip

	// Too many blocks to verify, so we only process a chunk and return
	if maxBlocksToResolve != 0 && uint64(len(unverifiedBlocks)) > maxBlocksToResolve {
		processingPointIndex := uint64(len(unverifiedBlocks)) - maxBlocksToResolve
		processingPoint = unverifiedBlocks[processingPointIndex]
		isNewVirtualSelectedParent, err := csm.isNewSelectedTip(readStagingArea, processingPoint, previousVirtualSelectedParent)
		if err != nil {
			return nil, false, err
		}

		// We must find a processing point which wins previous virtual selected parent
		// even if we process more than `maxBlocksToResolve` for that.
		// Otherwise, internal UTXO diff logic gets all messed up
		for !isNewVirtualSelectedParent {
			if processingPointIndex == 0 {
				return nil, false, errors.Errorf(
					"Expecting the pending tip %s to overcome the previous selected parent %s", pendingTip, previousVirtualSelectedParent)
			}
			processingPointIndex--
			processingPoint = unverifiedBlocks[processingPointIndex]
			isNewVirtualSelectedParent, err = csm.isNewSelectedTip(readStagingArea, processingPoint, previousVirtualSelectedParent)
			if err != nil {
				return nil, false, err
			}
		}
		log.Debugf("Has more than %d blocks to resolve. Setting the resolve processing point to %s", maxBlocksToResolve, processingPoint)
	}

	processingPointStatus, reversalData, err := csm.resolveBlockStatus(
		resolveStagingArea, processingPoint, true)
	if err != nil {
		return nil, false, err
	}

	if processingPointStatus == externalapi.StatusUTXOValid {
		err = staging.CommitAllChanges(csm.databaseContext, resolveStagingArea)
		if err != nil {
			return nil, false, err
		}

		if reversalData != nil {
			err = csm.ReverseUTXODiffs(processingPoint, reversalData)
			if err != nil {
				return nil, false, err
			}
		}
	}

	isActualTip := processingPoint.Equal(pendingTip)
	isCompletelyResolved := isActualTip && processingPointStatus == externalapi.StatusUTXOValid

	updateVirtualStagingArea := model.NewStagingArea()

	virtualParents := []*externalapi.DomainHash{processingPoint}
	// If `isCompletelyResolved`, set virtual correctly with all tips which have less blue work than pending
	if isCompletelyResolved {
		lowerTips, err := csm.getGHOSTDAGLowerTips(readStagingArea, pendingTip)
		if err != nil {
			return nil, false, err
		}
		log.Debugf("Picking virtual parents from relevant tips len: %d", len(lowerTips))

		virtualParents, err = csm.pickVirtualParents(readStagingArea, lowerTips)
		if err != nil {
			return nil, false, err
		}
		log.Debugf("Picked virtual parents: %s", virtualParents)
	}
	virtualUTXODiff, err := csm.updateVirtualWithParents(updateVirtualStagingArea, virtualParents)
	if err != nil {
		return nil, false, err
	}

	err = staging.CommitAllChanges(csm.databaseContext, updateVirtualStagingArea)
	if err != nil {
		return nil, false, err
	}

	selectedParentChainChanges, err := csm.dagTraversalManager.
		CalculateChainPath(updateVirtualStagingArea, previousVirtualSelectedParent, pendingTip)
	if err != nil {
		return nil, false, err
	}

	virtualParentsOutcome, err := csm.dagTopologyManager.Parents(updateVirtualStagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, false, err
	}

	return &externalapi.VirtualChangeSet{
		VirtualSelectedParentChainChanges: selectedParentChainChanges,
		VirtualUTXODiff:                   virtualUTXODiff,
		VirtualParents:                    virtualParentsOutcome,
	}, isCompletelyResolved, nil
}
