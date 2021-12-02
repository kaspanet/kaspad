package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util/staging"
	"sort"
)

func (csm *consensusStateManager) ResolveVirtual(maxBlocksToResolve uint64) (bool, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "csm.ResolveVirtual")
	defer onEnd()

	readStagingArea := model.NewStagingArea()
	tips, err := csm.consensusStateStore.Tips(readStagingArea, csm.databaseContext)
	if err != nil {
		return false, err
	}

	var sortErr error
	sort.Slice(tips, func(i, j int) bool {
		selectedParent, err := csm.ghostdagManager.ChooseSelectedParent(readStagingArea, tips[i], tips[j])
		if err != nil {
			sortErr = err
			return false
		}

		return selectedParent.Equal(tips[i])
	})
	if sortErr != nil {
		return false, sortErr
	}

	var selectedTip *externalapi.DomainHash
	isCompletelyResolved := true
	for _, tip := range tips {
		log.Debugf("Resolving tip %s", tip)
		resolveStagingArea := model.NewStagingArea()
		unverifiedBlocks, err := csm.getUnverifiedChainBlocks(resolveStagingArea, tip)
		if err != nil {
			return false, err
		}

		resolveTip := tip
		hasMoreUnverifiedThanMax := maxBlocksToResolve != 0 && uint64(len(unverifiedBlocks)) > maxBlocksToResolve
		if hasMoreUnverifiedThanMax {
			resolveTip = unverifiedBlocks[uint64(len(unverifiedBlocks))-maxBlocksToResolve]
			log.Debugf("Has more than %d blocks to resolve. Changing the resolve tip to %s", maxBlocksToResolve, resolveTip)
		}

		blockStatus, reversalData, err := csm.resolveBlockStatus(resolveStagingArea, resolveTip, true)
		if err != nil {
			return false, err
		}

		if blockStatus == externalapi.StatusUTXOValid {
			selectedTip = resolveTip
			isCompletelyResolved = !hasMoreUnverifiedThanMax

			err = staging.CommitAllChanges(csm.databaseContext, resolveStagingArea)
			if err != nil {
				return false, err
			}

			if reversalData != nil {
				err = csm.ReverseUTXODiffs(resolveTip, reversalData)
				if err != nil {
					return false, err
				}
			}
			break
		}
	}

	if selectedTip == nil {
		log.Warnf("Non of the DAG tips are valid")
		return true, nil
	}

	updateVirtualStagingArea := model.NewStagingArea()
	utxoDiff, err := csm.updateVirtualWithParents(updateVirtualStagingArea, []*externalapi.DomainHash{selectedTip})
	if err != nil {
		return false, err
	}

	err = staging.CommitAllChanges(csm.databaseContext, updateVirtualStagingArea)
	if err != nil {
		return false, err
	}

	return isCompletelyResolved, csm.onResolveVirtualHandler(&externalapi.BlockInsertionResult{
		VirtualUTXODiff: utxoDiff,
		VirtualParents:  tips,
	})
}
