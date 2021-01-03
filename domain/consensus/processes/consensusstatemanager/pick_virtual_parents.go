package consensusstatemanager

import (
	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
)

func (csm *consensusStateManager) pickVirtualParents(tips []*externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	log.Debugf("pickVirtualParents start for tips len: %d", len(tips))
	defer log.Debugf("pickVirtualParents end for tips len: %d", len(tips))

	log.Debugf("Pushing all tips into a DownHeap")
	candidatesHeap := csm.dagTraversalManager.NewDownHeap()
	for _, tip := range tips {
		err := candidatesHeap.Push(tip)
		if err != nil {
			return nil, err
		}
	}

	// If the first candidate has been disqualified from the chain or violates finality -
	// it cannot be virtual's parent, since it will make it virtual's selectedParent - disqualifying virtual itself.
	// Therefore, in such a case we remove it from the list of virtual parent candidates, and replace with
	// its parents that have no disqualified children
	virtualSelectedParent, err := csm.selectVirtualSelectedParent(candidatesHeap)
	if err != nil {
		return nil, err
	}
	log.Debugf("The selected parent of the virtual is: %s", virtualSelectedParent)

	selectedVirtualParents := hashset.NewFromSlice(virtualSelectedParent)
	candidates := candidatesHeap.ToSlice()
	// prioritize half the blocks with highest blueWork and half with lowest, so the network will merge splits faster.
	if len(candidates) >= int(csm.maxBlockParents) {
		// We already have the selectedParent, so we're left with csm.maxBlockParents-1.
		maxParents := csm.maxBlockParents - 1
		end := len(candidates) - 1
		for i := (maxParents) / 2; i < maxParents; i++ {
			candidates[i], candidates[end] = candidates[end], candidates[i]
			end--
		}
	}

	mergeSetSize := uint64(1) // starts counting from 1 because selectedParent is already in the mergeSet

	for len(candidates) > 0 && uint64(len(selectedVirtualParents)) < uint64(csm.maxBlockParents) {
		candidate := candidates[0]
		candidates = candidates[1:]

		log.Debugf("Attempting to add %s to the virtual parents", candidate)
		log.Debugf("The current merge set size is %d", mergeSetSize)

		mergeSetIncrease, err := csm.mergeSetIncrease(candidate, selectedVirtualParents)
		if err != nil {
			return nil, err
		}
		log.Debugf("The merge set would increase by %d with block %s", mergeSetIncrease, candidate)

		if mergeSetSize+mergeSetIncrease > csm.mergeSetSizeLimit {
			log.Debugf("Cannot add block %s since that would violate the merge set size limit", candidate)
			continue
		}

		selectedVirtualParents.Add(candidate)
		mergeSetSize += mergeSetIncrease
		log.Tracef("Added block %s to the virtual parents set", candidate)
	}

	boundedMergeBreakingParents, err := csm.boundedMergeBreakingParents(selectedVirtualParents.ToSlice())
	if err != nil {
		return nil, err
	}
	log.Tracef("The following parents are omitted for "+
		"breaking the bounded merge set: %s", boundedMergeBreakingParents)

	virtualParents := selectedVirtualParents.Subtract(boundedMergeBreakingParents).ToSlice()
	log.Tracef("The virtual parents resolved to be: %s", virtualParents)
	return virtualParents, nil
}

func (csm *consensusStateManager) selectVirtualSelectedParent(
	candidatesHeap model.BlockHeap) (*externalapi.DomainHash, error) {

	log.Tracef("selectVirtualSelectedParent start")
	defer log.Tracef("selectVirtualSelectedParent end")

	disqualifiedCandidates := hashset.New()

	for {
		if candidatesHeap.Len() == 0 {
			return nil, errors.New("virtual has no valid parent candidates")
		}
		selectedParentCandidate := candidatesHeap.Pop()

		log.Debugf("Checking block %s for selected parent eligibility", selectedParentCandidate)
		selectedParentCandidateStatus, err := csm.blockStatusStore.Get(csm.databaseContext, selectedParentCandidate)
		if err != nil {
			return nil, err
		}
		if selectedParentCandidateStatus == externalapi.StatusUTXOValid {
			log.Debugf("Block %s is valid. Returning it as the selected parent", selectedParentCandidate)
			return selectedParentCandidate, nil
		}

		log.Debugf("Block %s is not valid. Adding it to the disqualified set", selectedParentCandidate)
		disqualifiedCandidates.Add(selectedParentCandidate)

		candidateParents, err := csm.dagTopologyManager.Parents(selectedParentCandidate)
		if err != nil {
			return nil, err
		}
		log.Debugf("The parents of block %s are: %s", selectedParentCandidate, candidateParents)
		for _, parent := range candidateParents {
			parentChildren, err := csm.dagTopologyManager.Children(parent)
			if err != nil {
				return nil, err
			}

			// remove virtual from parentChildren if it's there
			for i, parentChild := range parentChildren {
				if parentChild.Equal(model.VirtualBlockHash) {
					parentChildren = append(parentChildren[:i], parentChildren[i+1:]...)
					break
				}
			}
			log.Debugf("The children of block %s are: %s", parent, parentChildren)

			if disqualifiedCandidates.ContainsAllInSlice(parentChildren) {
				log.Debugf("The disqualified set contains all the "+
					"children of %s. Adding it to the candidate heap", parentChildren)
				err := candidatesHeap.Push(parent)
				if err != nil {
					return nil, err
				}
			}
		}
	}
}

func (csm *consensusStateManager) mergeSetIncrease(
	candidate *externalapi.DomainHash, selectedVirtualParents hashset.HashSet) (uint64, error) {

	log.Tracef("mergeSetIncrease start")
	defer log.Tracef("mergeSetIncrease end")

	visited := hashset.New()
	queue := csm.dagTraversalManager.NewDownHeap()
	err := queue.Push(candidate)
	if err != nil {
		return 0, err
	}
	mergeSetIncrease := uint64(1) // starts with 1 for the candidate itself

	for queue.Len() > 0 {
		current := queue.Pop()
		log.Debugf("Attempting to increment the merge set size increase for block %s", current)

		isInPastOfSelectedVirtualParents, err := csm.dagTopologyManager.IsAncestorOfAny(
			current, selectedVirtualParents.ToSlice())
		if err != nil {
			return 0, err
		}
		if isInPastOfSelectedVirtualParents {
			log.Debugf("Skipping block %s because it's in the past of one "+
				"(or more) of the selected virtual parents", current)
			continue
		}

		log.Tracef("Incrementing the merge set size increase")
		mergeSetIncrease++

		parents, err := csm.dagTopologyManager.Parents(current)
		if err != nil {
			return 0, err
		}
		for _, parent := range parents {
			if !visited.Contains(parent) {
				visited.Add(parent)
				err = queue.Push(parent)
				if err != nil {
					return 0, err
				}
			}
		}
	}
	log.Debugf("The resolved merge set size increase is: %d", mergeSetIncrease)

	return mergeSetIncrease, nil
}

func (csm *consensusStateManager) boundedMergeBreakingParents(
	parents []*externalapi.DomainHash) (hashset.HashSet, error) {

	log.Tracef("boundedMergeBreakingParents start for parents: %s", parents)
	defer log.Tracef("boundedMergeBreakingParents end for parents: %s", parents)

	log.Debug("Temporarily setting virtual to all parents, so that we can run ghostdag on it")
	err := csm.dagTopologyManager.SetParents(model.VirtualBlockHash, parents)
	if err != nil {
		return nil, err
	}

	err = csm.ghostdagManager.GHOSTDAG(model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	potentiallyKosherizingBlocks, err := csm.mergeDepthManager.NonBoundedMergeDepthViolatingBlues(model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	log.Debugf("The potentially kosherizing blocks are: %s", potentiallyKosherizingBlocks)

	virtualFinalityPoint, err := csm.finalityManager.VirtualFinalityPoint()
	if err != nil {
		return nil, err
	}
	log.Debugf("The finality point of the virtual is: %s", virtualFinalityPoint)

	var badReds []*externalapi.DomainHash

	virtualGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	for _, redBlock := range virtualGHOSTDAGData.MergeSetReds() {
		log.Debugf("Check whether red block %s is kosherized", redBlock)
		isFinalityPointInPast, err := csm.dagTopologyManager.IsAncestorOf(virtualFinalityPoint, redBlock)
		if err != nil {
			return nil, err
		}
		if isFinalityPointInPast {
			log.Debugf("Skipping red block %s because it has the virtual's"+
				" finality point in its past", redBlock)
			continue
		}

		isKosherized := false
		for _, potentiallyKosherizingBlock := range potentiallyKosherizingBlocks {
			isKosherized, err = csm.dagTopologyManager.IsAncestorOf(redBlock, potentiallyKosherizingBlock)
			if err != nil {
				return nil, err
			}
			log.Debugf("Red block %s is an ancestor of potentially kosherizing "+
				"block %s, therefore the red block is kosher", redBlock, potentiallyKosherizingBlock)
			if isKosherized {
				break
			}
		}
		if !isKosherized {
			log.Debugf("Red block %s is not kosher. Adding it to the bad reds set", redBlock)
			badReds = append(badReds, redBlock)
		}
	}

	boundedMergeBreakingParents := hashset.New()
	for _, parent := range parents {
		log.Debugf("Checking whether parent %s breaks the bounded merge set", parent)
		isBadRedInPast := false
		for _, badRedBlock := range badReds {
			isBadRedInPast, err = csm.dagTopologyManager.IsAncestorOf(parent, badRedBlock)
			if err != nil {
				return nil, err
			}
			if isBadRedInPast {
				log.Debugf("Parent %s is an ancestor of bad red %s", parent, badRedBlock)
				break
			}
		}
		if isBadRedInPast {
			log.Debugf("Adding parent %s to the bounded merge breaking parents set", parent)
			boundedMergeBreakingParents.Add(parent)
		}
	}

	return boundedMergeBreakingParents, nil
}
