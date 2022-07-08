package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util/staging"
	"github.com/pkg/errors"
	"sort"
)

func (csm *consensusStateManager) findNextPendingTip() (*externalapi.DomainHash, externalapi.BlockStatus, error) {
	readStagingArea := model.NewStagingArea()
	tips, err := csm.consensusStateStore.Tips(readStagingArea, csm.databaseContext)
	if err != nil {
		return nil, externalapi.StatusInvalid, err
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
		return nil, externalapi.StatusInvalid, sortErr
	}

	for _, tip := range tips {
		log.Debugf("Resolving tip %s", tip)
		isViolatingFinality, shouldNotify, err := csm.isViolatingFinality(readStagingArea, tip)
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

		status, err := csm.blockStatusStore.Get(csm.databaseContext, readStagingArea, tip)
		if err != nil {
			return nil, externalapi.StatusInvalid, err
		}
		if status == externalapi.StatusUTXOValid || status == externalapi.StatusUTXOPendingVerification {
			return tip, status, nil
		}
	}

	return nil, externalapi.StatusInvalid, nil
}

func (csm *consensusStateManager) ResolveVirtual(maxBlocksToResolve uint64) (*externalapi.VirtualChangeSet, bool, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "csm.ResolveVirtual")
	defer onEnd()

	readStagingArea := model.NewStagingArea()

	/*
		Algo (begin resolve):
			Go over tips by GHOSTDAG select parent order ignoring UTXO disqualified blocks and finality violating blocks
			Set pending tip to the first tip
			if this tip is already UTXO valid, finalize virtual state and return
			if the tip is UTXO pending, find the earliest UTXO pending block in its chain and set it as position
			set the tip as resolving virtual pending
			try resolving a chunk up the chain from position to pending tip

		Algo (continue resolve):
			Start from position and try to continue resolving another chunk up the chain to pending tip
			If we encounter a UTXO disqualified block, we should
				mark the whole chain up to pending tip as disqualified
				set position and pending tip to the next candidate chain
				return and let the next call continue the processing
			If we reach the tip, and it is valid, only then set virtual parents to DAG tips, and clear resolving state
	*/

	pendingTip, pendingTipStatus, err := csm.findNextPendingTip()
	if err != nil {
		return nil, false, err
	}

	if pendingTip == nil {
		log.Warnf("Non of the DAG tips are valid")
		return &externalapi.VirtualChangeSet{}, true, nil
	}

	prevVirtualSelectedParent, err := csm.virtualSelectedParent(readStagingArea)
	if err != nil {
		return nil, false, err
	}

	if pendingTipStatus == externalapi.StatusUTXOValid && prevVirtualSelectedParent.Equal(pendingTip) {
		return &externalapi.VirtualChangeSet{}, true, nil
	}

	// Resolve a chunk from the pending chain
	resolveStagingArea := model.NewStagingArea()
	unverifiedBlocks, err := csm.getUnverifiedChainBlocks(resolveStagingArea, pendingTip)
	if err != nil {
		return nil, false, err
	}

	intermediateTip := pendingTip

	// Too many blocks to verify, so we only process a chunk and return
	if maxBlocksToResolve != 0 && uint64(len(unverifiedBlocks)) > maxBlocksToResolve {

		intermediateTipIndex := uint64(len(unverifiedBlocks)) - maxBlocksToResolve
		intermediateTip = unverifiedBlocks[intermediateTipIndex]
		isNewVirtualSelectedParent, err := csm.isNewSelectedTip(readStagingArea, intermediateTip, prevVirtualSelectedParent)
		if err != nil {
			return nil, false, err
		}

		// We must find an intermediate tip which wins previous virtual selected parent
		// even if we process more than `maxBlocksToResolve` for that.
		// Otherwise, internal UTXO diff logic gets all messed up
		for !isNewVirtualSelectedParent {
			if intermediateTipIndex == 0 {
				return nil, false, errors.Errorf(
					"Expecting the pending tip %s to overcome the previous selected parent %s", pendingTip, prevVirtualSelectedParent)
			}
			intermediateTipIndex--
			intermediateTip = unverifiedBlocks[intermediateTipIndex]
			isNewVirtualSelectedParent, err = csm.isNewSelectedTip(readStagingArea, intermediateTip, prevVirtualSelectedParent)
			if err != nil {
				return nil, false, err
			}
		}
		log.Debugf("Has more than %d blocks to resolve. Changing the resolve tip to %s", maxBlocksToResolve, intermediateTip)
	}

	intermediateTipStatus, reversalData, err := csm.resolveBlockStatus(
		resolveStagingArea, intermediateTip, true)
	if err != nil {
		return nil, false, err
	}

	if intermediateTipStatus == externalapi.StatusUTXOValid {
		err = staging.CommitAllChanges(csm.databaseContext, resolveStagingArea)
		if err != nil {
			return nil, false, err
		}

		if reversalData != nil {
			err = csm.ReverseUTXODiffs(intermediateTip, reversalData)
			if err != nil {
				return nil, false, err
			}
		}
	}

	isActualTip := intermediateTip.Equal(pendingTip)
	isCompletelyResolved := isActualTip && intermediateTipStatus == externalapi.StatusUTXOValid

	updateVirtualStagingArea := model.NewStagingArea()

	// TODO: if `isCompletelyResolved`, set virtual correctly with all tips which have less blue work than pending
	virtualUTXODiff, err := csm.updateVirtualWithParents(updateVirtualStagingArea, []*externalapi.DomainHash{intermediateTip})
	if err != nil {
		return nil, false, err
	}

	err = staging.CommitAllChanges(csm.databaseContext, updateVirtualStagingArea)
	if err != nil {
		return nil, false, err
	}

	// TODO: why was `readStagingArea` used here ?
	selectedParentChainChanges, err := csm.dagTraversalManager.
		CalculateChainPath(updateVirtualStagingArea, prevVirtualSelectedParent, pendingTip)
	if err != nil {
		return nil, false, err
	}

	virtualParents, err := csm.dagTopologyManager.Parents(updateVirtualStagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, false, err
	}

	return &externalapi.VirtualChangeSet{
		VirtualSelectedParentChainChanges: selectedParentChainChanges,
		VirtualUTXODiff:                   virtualUTXODiff,
		VirtualParents:                    virtualParents,
	}, isCompletelyResolved, nil
}
