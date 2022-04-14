package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util/staging"
	"github.com/pkg/errors"
	"sort"
)

func (csm *consensusStateManager) ResolveVirtual(maxBlocksToResolve uint64) (*externalapi.VirtualChangeSet, bool, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "csm.ResolveVirtual")
	defer onEnd()

	readStagingArea := model.NewStagingArea()
	tips, err := csm.consensusStateStore.Tips(readStagingArea, csm.databaseContext)
	if err != nil {
		return nil, false, err
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
		return nil, false, sortErr
	}

	var selectedTip *externalapi.DomainHash
	isCompletelyResolved := true
	for _, tip := range tips {
		log.Debugf("Resolving tip %s", tip)
		resolveStagingArea := model.NewStagingArea()
		unverifiedBlocks, err := csm.getUnverifiedChainBlocks(resolveStagingArea, tip)
		if err != nil {
			return nil, false, err
		}

		resolveTip := tip
		hasMoreUnverifiedThanMax := maxBlocksToResolve != 0 && uint64(len(unverifiedBlocks)) > maxBlocksToResolve
		if hasMoreUnverifiedThanMax {
			resolveTip = unverifiedBlocks[uint64(len(unverifiedBlocks))-maxBlocksToResolve]
			log.Debugf("Has more than %d blocks to resolve. Changing the resolve tip to %s", maxBlocksToResolve, resolveTip)
		}

		blockStatus, reversalData, err := csm.resolveBlockStatus(resolveStagingArea, resolveTip, true)
		if err != nil {
			return nil, false, err
		}

		if blockStatus == externalapi.StatusUTXOValid {
			selectedTip = resolveTip
			isCompletelyResolved = !hasMoreUnverifiedThanMax

			err = staging.CommitAllChanges(csm.databaseContext, resolveStagingArea)
			if err != nil {
				return nil, false, err
			}

			if reversalData != nil {
				err = csm.ReverseUTXODiffs(resolveTip, reversalData)
				// It's still not known what causes this error, but we can ignore it and not reverse the UTXO diffs
				// and harm performance in some cases.
				// TODO: Investigate why this error happens in the first place, and remove the workaround.
				if errors.Is(err, ErrReverseUTXODiffsUTXODiffChildNotFound) {
					log.Errorf("Could not reverse UTXO diffs while resolving virtual: %s", err)
				} else if err != nil {
					return nil, false, err
				}
			}
			break
		}
	}

	if selectedTip == nil {
		log.Warnf("Non of the DAG tips are valid")
		return nil, true, nil
	}

	oldVirtualGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, readStagingArea, model.VirtualBlockHash, false)
	if err != nil {
		return nil, false, err
	}

	updateVirtualStagingArea := model.NewStagingArea()
	virtualUTXODiff, err := csm.updateVirtualWithParents(updateVirtualStagingArea, []*externalapi.DomainHash{selectedTip})
	if err != nil {
		return nil, false, err
	}

	err = staging.CommitAllChanges(csm.databaseContext, updateVirtualStagingArea)
	if err != nil {
		return nil, false, err
	}

	selectedParentChainChanges, err := csm.dagTraversalManager.
		CalculateChainPath(readStagingArea, oldVirtualGHOSTDAGData.SelectedParent(), selectedTip)
	if err != nil {
		return nil, false, err
	}

	virtualParents, err := csm.dagTopologyManager.Parents(readStagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, false, err
	}

	return &externalapi.VirtualChangeSet{
		VirtualSelectedParentChainChanges: selectedParentChainChanges,
		VirtualUTXODiff:                   virtualUTXODiff,
		VirtualParents:                    virtualParents,
	}, isCompletelyResolved, nil
}
