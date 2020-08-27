package blockdag

import (
	"github.com/kaspanet/kaspad/util/mstime"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/util/daghash"
)

// FinalityConflict represents an entry in the finality conflicts event log of the DAG
type FinalityConflict struct {
	ID                 int
	ConflictTime       mstime.Time
	SelectedTipHash    *daghash.Hash
	ViolatingBlockHash *daghash.Hash

	// Only resolved FinalityConflict will have non-null ResolutionTime
	ResolutionTime *mstime.Time
}

// FinalityConflicts returns the list of all finality conflicts that occured in this node
func (dag *BlockDAG) FinalityConflicts() []*FinalityConflict {
	dag.dagLock.RLock()
	defer dag.dagLock.RUnlock()

	return dag.finalityConflicts
}

func (dag *BlockDAG) finalityConflictByID(id int) (*FinalityConflict, bool) {
	for _, finalityConflict := range dag.finalityConflicts {
		if finalityConflict.ID == id {
			return finalityConflict, true
		}
	}
	return nil, false
}

func (dag *BlockDAG) addFinalityConflict(violatingNode *blockNode) {
	topFinalityConflictID := 0
	for _, finalityConfict := range dag.finalityConflicts {
		if finalityConfict.ID > topFinalityConflictID {
			topFinalityConflictID = finalityConfict.ID
		}
	}

	finalityConflict := &FinalityConflict{
		ID:                 topFinalityConflictID + 1,
		ConflictTime:       mstime.Now(),
		SelectedTipHash:    dag.SelectedTipHash(),
		ViolatingBlockHash: violatingNode.hash,
	}

	dag.finalityConflicts = append(dag.finalityConflicts, finalityConflict)

	dag.sendNotification(NTFinalityConflict, &FinalityConflictNotificationData{FinalityConflict: finalityConflict})
}

// ResolveFinalityConflict resolves
func (dag *BlockDAG) ResolveFinalityConflict(id int, validBlockHashes, invalidBlockHashes []*daghash.Hash) error {

	finalityConflict, ok := dag.finalityConflictByID(id)
	if !ok {
		return errors.Errorf("No finality conflict with ID %d found", id)
	}

	selectedTip, violatingBlock, validBlocks, invalidBlocks, err :=
		dag.lookupFinalityConflictBlocks(finalityConflict, validBlockHashes, invalidBlockHashes)
	if err != nil {
		return err
	}

	violatingBranchStart, selectedTipBranchStart :=
		dag.findFinalityConflictBranchStarts(violatingBlock, selectedTip)

	isSwitchingBranches, err := dag.checkIfSwitchingBranches(selectedTipBranchStart, violatingBranchStart,
		selectedTip, violatingBlock, validBlocks, invalidBlocks)
	if err != nil {
		return err
	}

	if !isSwitchingBranches {
		isKeepingBranches, err := dag.checkIfKeepingBranches(selectedTipBranchStart, violatingBranchStart,
			selectedTip, violatingBlock, validBlocks, invalidBlocks)
		if err != nil {
			return err
		}
		if !isKeepingBranches {
			return errors.Errorf("neither Switching Branches not Keeping Branches conditions were met")
		}
	}

	var validBranchStart *blockNode
	if isSwitchingBranches {
		validBranchStart = violatingBranchStart
	} else {
		validBranchStart = selectedTip
	}

	addedTips := newBlockSet()
	if isSwitchingBranches {
		addedTipsFromValidBlocksFuture := dag.updateValidBlocksFuture(validBlocks)
		addedTips.addSet(addedTipsFromValidBlocksFuture)
	}

	removedTips, rehabilitatedBlocks, err := dag.updateInvalidBlocksFuture(invalidBlocks, validBranchStart)
	if err != nil {
		return err
	}

	addedTipsFromRehabilitateBlocks := dag.rehabilitateBlocks(rehabilitatedBlocks)
	addedTips.addSet(addedTipsFromRehabilitateBlocks)

	virtualSelectedParentChainUpdates, err :=
		dag.updateTipsAfterFinalityConflictResolution(removedTips, addedTips)
	if err != nil {
		return err
	}

	err = dag.updateFinalityConflictResolution(finalityConflict)
	if err != nil {
		return err
	}

	dag.sendNotification(NTChainChanged, ChainChangedNotificationData{
		RemovedChainBlockHashes: virtualSelectedParentChainUpdates.removedChainBlockHashes,
		AddedChainBlockHashes:   virtualSelectedParentChainUpdates.addedChainBlockHashes,
	})

	return nil
}

func (dag *BlockDAG) rehabilitateBlocks(rehabilitatedBlocks blockSet) (addedTips blockSet) {
	addedTips = newBlockSet()

	for rehabilitatedBlock := range rehabilitatedBlocks {
		hasNonManuallyRejectedChildren := false
		for child := range rehabilitatedBlock.children {
			if dag.index.BlockNodeStatus(child) != statusManuallyRejected {
				hasNonManuallyRejectedChildren = true
				break
			}
		}
		if !hasNonManuallyRejectedChildren {
			addedTips.add(rehabilitatedBlock)
		}
	}

	return addedTips
}

func (dag *BlockDAG) updateInvalidBlocksFuture(
	invalidBlocks blockSet, validBranchStart *blockNode) (removedTips, rehabilitatedBlocks blockSet, err error) {

	queue := newUpHeap()
	queue.pushSet(invalidBlocks)
	visited := invalidBlocks.clone()
	rehabilitatedBlocks = newBlockSet()
	removedTips = newBlockSet()

	for queue.Len() > 0 {
		current := queue.pop()

		hasValidBranchStartInSelectedParentChain, err := dag.isInSelectedParentChainOf(validBranchStart, current)
		if err != nil {
			return nil, nil, err
		}
		if hasValidBranchStartInSelectedParentChain {
			rehabilitatedBlocks.add(current)
		} else {
			dag.index.SetBlockNodeStatus(current, statusManuallyRejected)
			if len(current.children) == 0 {
				removedTips.add(current)
			}
		}

		for child := range current.children {
			if !visited.contains(child) {
				queue.Push(child)
				visited.add(child)
			}
		}
	}

	return removedTips, rehabilitatedBlocks, nil
}

func (dag *BlockDAG) updateValidBlocksFuture(validBlocks blockSet) (addedTips blockSet) {
	addedTips = newBlockSet()

	queue := newUpHeap()
	queue.pushSet(validBlocks)
	visited := validBlocks.clone()
	for queue.Len() > 0 {
		current := queue.pop()
		if dag.index.BlockNodeStatus(current) == statusViolatedSubjectiveFinality {
			dag.index.SetBlockNodeStatus(current, statusUTXONotVerified)
			if len(current.children) == 0 {
				addedTips.add(current)
			}
			for child := range current.children {
				if !visited.contains(child) {
					queue.Push(child)
					visited.add(child)
				}
			}
		}
	}

	return addedTips
}

func (dag *BlockDAG) lookupFinalityConflictBlocks(
	finalityConflict *FinalityConflict, validBlockHashes []*daghash.Hash, invalidBlockHashes []*daghash.Hash) (
	selectedTip, violatingBlock *blockNode, validBlocks, invalidBlocks blockSet, err error) {

	selectedTip, ok := dag.index.LookupNode(finalityConflict.SelectedTipHash)
	if !ok {
		return nil, nil, nil, nil,
			errors.Errorf("Couldn't find selectedTip with hash %s", finalityConflict.SelectedTipHash)
	}

	violatingBlock, ok = dag.index.LookupNode(finalityConflict.ViolatingBlockHash)
	if !ok {
		return nil, nil, nil, nil,
			errors.Errorf("Couldn't find violatingBlock with hash %s", finalityConflict.ViolatingBlockHash)
	}

	validBlocksSlice, err := dag.index.LookupNodes(validBlockHashes)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	validBlocks = blockSetFromSlice(validBlocksSlice...)

	invalidBlocksSlice, err := dag.index.LookupNodes(invalidBlockHashes)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	invalidBlocks = blockSetFromSlice(invalidBlocksSlice...)

	return selectedTip, violatingBlock, validBlocks, invalidBlocks, nil
}

func (dag *BlockDAG) findFinalityConflictBranchStarts(selectedTip *blockNode, violatingBlock *blockNode) (
	selectedTipBranchStart, violatingBranchStart *blockNode) {

	selectedTipBranchStart = selectedTip
	violatingBranchStart = violatingBlock

	for selectedTipBranchStart.selectedParent != violatingBranchStart.selectedParent {
		if selectedTipBranchStart.selectedParent.blueScore > violatingBranchStart.selectedParent.blueScore {
			selectedTipBranchStart = selectedTipBranchStart.selectedParent
		} else {
			violatingBranchStart = violatingBranchStart.selectedParent
		}
	}

	return selectedTipBranchStart, violatingBranchStart
}

func (dag *BlockDAG) checkIfSwitchingBranches(
	selectedTipBranchStart, violatingBranchStart, selectedTip, violatingBlock *blockNode,
	validBlocks, invalidBlocks blockSet) (bool, error) {

	// Make sure that all validBlocks have violatingBranchStart in their selectedParentChain
	isOK, err := dag.areAllInSelectedParentChainOf(validBlocks, violatingBranchStart)
	if err != nil || !isOK {
		return false, err
	}

	// Make sure that all invalidBlocks have selectedTipBranchStart in their selectedParentChain
	isOK, err = dag.areAllInSelectedParentChainOf(invalidBlocks, selectedTipBranchStart)
	if err != nil || !isOK {
		return false, err
	}

	// Make sure that at least one validBlock is violatingBlock or has violatingBlock in his past
	isOK, err = dag.isAnyInPastOf(validBlocks, violatingBlock)
	if err != nil || !isOK {
		return false, err
	}

	// Make sure that at least one invalidBlock is selectedTip or has selectedTip in his past
	isOK, err = dag.isAnyInPastOf(invalidBlocks, selectedTip)
	if err != nil || !isOK {
		return false, err
	}

	return true, nil
}

func (dag *BlockDAG) checkIfKeepingBranches(
	selectedTipBranchStart, violatingBranchStart, selectedTip, violatingBlock *blockNode,
	validBlocks, invalidBlocks blockSet) (bool, error) {

	// Make sure that all invalidBlocks have violatingBranchStart in their selectedParentChain
	isOK, err := dag.areAllInSelectedParentChainOf(invalidBlocks, violatingBranchStart)
	if err != nil || !isOK {
		return false, err
	}

	// Make sure that all validBlocks have selectedTipBranchStart in their selectedParentChain
	isOK, err = dag.areAllInSelectedParentChainOf(validBlocks, selectedTipBranchStart)
	if err != nil || !isOK {
		return false, err
	}

	// Make sure that at least one invalidBlock is violatingBlock or has violatingBlock in his past
	isOK, err = dag.isAnyInPastOf(invalidBlocks, violatingBlock)
	if err != nil || !isOK {
		return false, err
	}

	// Make sure that at least one validBlock is selectedTip or has selectedTip in his past
	isOK, err = dag.isAnyInPastOf(validBlocks, selectedTip)
	if err != nil || !isOK {
		return false, err
	}

	return true, nil
}

func (dag *BlockDAG) updateTipsAfterFinalityConflictResolution(
	removedTips, addedTips blockSet) (chainUpdates *chainUpdates, err error) {

	newTips := dag.tips.clone()
	for removedTip := range removedTips {
		newTips.remove(removedTip)
	}

	for addedTip := range addedTips {
		for parent := range addedTip.parents {
			newTips.remove(parent)
		}

		newTips.add(addedTip)
	}

	_, chainUpdates, err = dag.setTips(newTips)

	return chainUpdates, err
}

func (dag *BlockDAG) updateFinalityConflictResolution(resolvedFinalityConflict *FinalityConflict) (err error) {
	resolutionTime := mstime.Now()
	resolvedFinalityConflict.ResolutionTime = &resolutionTime

	dbTx, err := dag.databaseContext.NewTx()
	if err != nil {
		return err
	}
	err = dag.saveState(dbTx)
	if err != nil {
		return err
	}

	areAllResolved := true
	for _, finalityConflict := range dag.finalityConflicts {
		if finalityConflict.ResolutionTime == nil {
			areAllResolved = false
			break
		}
	}

	dag.sendNotification(NTFinalityConflictResolved, FinalityConflictResolvedNotificationData{
		FinalityConflictID:              resolvedFinalityConflict.ID,
		ResolutionTime:                  resolutionTime,
		AreAllFinalityConflictsResolved: areAllResolved,
	})

	return nil
}
