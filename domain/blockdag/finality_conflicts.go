package blockdag

import (
	"time"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/util/daghash"
)

type FinalityConflict struct {
	ID                     int
	ConflictTime           time.Time
	CurrentSelectedTipHash *daghash.Hash
	ViolatingBlockHash     *daghash.Hash

	ResolutionTime *time.Time
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

func (dag *BlockDAG) addFinalityConflict(node *blockNode) {
	topFinalityConflictID := 0
	for _, finalityConfict := range dag.finalityConflicts {
		if finalityConfict.ID > topFinalityConflictID {
			topFinalityConflictID = finalityConfict.ID
		}
	}

	finalityConflict := &FinalityConflict{
		ID:                     topFinalityConflictID + 1,
		ConflictTime:           time.Now(),
		CurrentSelectedTipHash: dag.SelectedTipHash(),
		ViolatingBlockHash:     node.hash,
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

	currentSelectedTip, violatingBlock, validBlocks, invalidBlocks, err :=
		dag.lookupFinalityConflictBlocks(finalityConflict, validBlockHashes, invalidBlockHashes)
	if err != nil {
		return err
	}

	violatingBranchStart, currentSelectedTipBranchStart :=
		dag.findFinalityConflictBranchStarts(violatingBlock, currentSelectedTip)

	isSwitchingBranches, err := dag.checkIfSwitchingBranches(currentSelectedTipBranchStart, violatingBranchStart,
		currentSelectedTip, violatingBlock, validBlocks, invalidBlocks)
	if err != nil {
		return err
	}

	if !isSwitchingBranches {
		isKeepingBranches, err := dag.checkIfKeepingBranches(currentSelectedTipBranchStart, violatingBranchStart,
			currentSelectedTip, violatingBlock, validBlocks, invalidBlocks)
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
		validBranchStart = currentSelectedTip
	}

	addedTips := newBlockSet()
	if isSwitchingBranches {
		addedTipsFromValidBlocksFuture := dag.UpdateValidBlocksFuture(validBlocks)
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

	areAllResolved, err = dag.updateFinalityConflictResolution(finalityConflict)
	if err != nil {
		return err
	}

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

func (dag *BlockDAG) UpdateValidBlocksFuture(validBlocks blockSet) (addedTips blockSet) {
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
	currentSelectedTip, violatingBlock *blockNode, validBlocks, invalidBlocks blockSet, err error) {

	currentSelectedTip, ok := dag.index.LookupNode(finalityConflict.CurrentSelectedTipHash)
	if !ok {
		return nil, nil, nil, nil,
			errors.Errorf("Couldn't find currentSelectedTip with hash %s", finalityConflict.CurrentSelectedTipHash)
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

	return currentSelectedTip, violatingBlock, validBlocks, invalidBlocks, nil
}

func (dag *BlockDAG) findFinalityConflictBranchStarts(currentSelectedTip *blockNode, violatingBlock *blockNode) (
	currentSelectedTipBranchStart, violatingBranchStart *blockNode) {

	currentSelectedTipBranchStart = currentSelectedTip
	violatingBranchStart = violatingBlock

	for currentSelectedTipBranchStart.selectedParent != violatingBranchStart.selectedParent {
		if currentSelectedTipBranchStart.selectedParent.blueScore > violatingBranchStart.selectedParent.blueScore {
			currentSelectedTipBranchStart = currentSelectedTipBranchStart.selectedParent
		} else {
			violatingBranchStart = violatingBranchStart.selectedParent
		}
	}

	return currentSelectedTipBranchStart, violatingBranchStart
}

func (dag *BlockDAG) checkIfSwitchingBranches(
	currentSelectedTipBranchStart, violatingBranchStart, currentSelectedTip, violatingBlock *blockNode,
	validBlocks, invalidBlocks blockSet) (bool, error) {

	// Make sure that all validBlocks have violatingBranchStart in their selectedParentChain
	isOK, err := dag.areAllInSelectedParentChainOf(validBlocks, violatingBranchStart)
	if err != nil || !isOK {
		return false, err
	}

	// Make sure that all invalidBlocks have currentSelectedTipBranchStart in their selectedParentChain
	isOK, err = dag.areAllInSelectedParentChainOf(invalidBlocks, currentSelectedTipBranchStart)
	if err != nil || !isOK {
		return false, err
	}

	// Make sure that at least one validBlock is violatingBlock or has violatingBlock in his past
	isOK, err = dag.isAnyInPastOf(validBlocks, violatingBlock)
	if err != nil || !isOK {
		return false, err
	}

	// Make sure that at least one invalidBlock is currentSelectedTip or has currentSelectedTip in his past
	isOK, err = dag.isAnyInPastOf(invalidBlocks, currentSelectedTip)
	if err != nil || !isOK {
		return false, err
	}

	return true, nil
}

func (dag *BlockDAG) checkIfKeepingBranches(
	currentSelectedTipBranchStart, violatingBranchStart, currentSelectedTip, violatingBlock *blockNode,
	validBlocks, invalidBlocks blockSet) (bool, error) {

	// Make sure that all invalidBlocks have violatingBranchStart in their selectedParentChain
	isOK, err := dag.areAllInSelectedParentChainOf(invalidBlocks, violatingBranchStart)
	if err != nil || !isOK {
		return false, err
	}

	// Make sure that all validBlocks have currentSelectedTipBranchStart in their selectedParentChain
	isOK, err = dag.areAllInSelectedParentChainOf(validBlocks, currentSelectedTipBranchStart)
	if err != nil || !isOK {
		return false, err
	}

	// Make sure that at least one invalidBlock is violatingBlock or has violatingBlock in his past
	isOK, err = dag.isAnyInPastOf(invalidBlocks, violatingBlock)
	if err != nil || !isOK {
		return false, err
	}

	// Make sure that at least one validBlock is currentSelectedTip or has currentSelectedTip in his past
	isOK, err = dag.isAnyInPastOf(validBlocks, currentSelectedTip)
	if err != nil || !isOK {
		return false, err
	}

	return true, nil
}

func (dag *BlockDAG) updateTipsAfterFinalityConflictResolution(removedTips, addedTips blockSet) (
	virtualSelectedParentChainUpdates *chainUpdates, err error) {

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

	_, virtualSelectedParentChainUpdates, err = dag.setTips(newTips)

	return virtualSelectedParentChainUpdates, err
}
