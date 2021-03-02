package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
)

func (csm *consensusStateManager) pickVirtualParents(tips []*externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "pickVirtualParents")
	defer onEnd()

	log.Debugf("pickVirtualParents start for tips len: %d", len(tips))

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
	// Limit to 30 candidates, that way we don't go over thousands of tips when the network isn't healthy.
	if len(candidates) > int(csm.maxBlockParents)*3 {
		candidates = candidates[:int(csm.maxBlockParents)*3]
	}

	selectedVirtualParents := []*externalapi.DomainHash{virtualSelectedParent}
	mergeSetSize := uint64(1) // starts counting from 1 because selectedParent is already in the mergeSet

	for len(candidates) > 0 && uint64(len(selectedVirtualParents)) < uint64(csm.maxBlockParents) {
		candidate := candidates[0]
		candidates = candidates[1:]

		log.Debugf("Attempting to add %s to the virtual parents", candidate)
		log.Debugf("The current merge set size is %d", mergeSetSize)

		canBeParent, newCandidate, mergeSetIncrease, err := csm.mergeSetIncrease(candidate, selectedVirtualParents, mergeSetSize)
		if err != nil {
			return nil, err
		}
		if canBeParent {
			mergeSetSize += mergeSetIncrease
			selectedVirtualParents = append(selectedVirtualParents, candidate)
			log.Tracef("Added block %s to the virtual parents set", candidate)
			continue
		}
		// If we already have a candidate in the past of newCandidate then skip.
		isInFutureOfCandidates, err := csm.dagTopologyManager.IsAnyAncestorOf(candidates, newCandidate)
		if err != nil {
			return nil, err
		}
		if isInFutureOfCandidates {
			continue
		}
		// Remove all candidates in the future of newCandidate
		candidates, err = csm.removeHashesInFutureOf(candidates, newCandidate)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, newCandidate)
		log.Debugf("Cannot add block %s, instead added new candidate: %s", candidate, newCandidate)
	}

	boundedMergeBreakingParents, err := csm.boundedMergeBreakingParents(selectedVirtualParents)
	if err != nil {
		return nil, err
	}
	log.Tracef("The following parents are omitted for "+
		"breaking the bounded merge set: %s", boundedMergeBreakingParents)

	// Remove all boundedMergeBreakingParents from selectedVirtualParents
	for _, breakingParent := range boundedMergeBreakingParents {
		for i, parent := range selectedVirtualParents {
			if parent.Equal(breakingParent) {
				selectedVirtualParents[i] = selectedVirtualParents[len(selectedVirtualParents)-1]
				selectedVirtualParents = selectedVirtualParents[:len(selectedVirtualParents)-1]
				break
			}
		}
	}
	log.Tracef("The virtual parents resolved to be: %s", selectedVirtualParents)
	return selectedVirtualParents, nil
}

func (csm *consensusStateManager) removeHashesInFutureOf(hashes []*externalapi.DomainHash, ancestor *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	// Source: https://github.com/golang/go/wiki/SliceTricks#filter-in-place
	i := 0
	for _, hash := range hashes {
		isInFutureOfAncestor, err := csm.dagTopologyManager.IsAncestorOf(ancestor, hash)
		if err != nil {
			return nil, err
		}
		if !isInFutureOfAncestor {
			hashes[i] = hash
			i++
		}
	}
	return hashes[:i], nil
}

func (csm *consensusStateManager) selectVirtualSelectedParent(
	candidatesHeap model.BlockHeap) (*externalapi.DomainHash, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "selectVirtualSelectedParent")
	defer onEnd()

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
			allParentChildren, err := csm.dagTopologyManager.Children(parent)
			if err != nil {
				return nil, err
			}
			log.Debugf("The children of block %s are: %s", parent, allParentChildren)

			// remove virtual and any headers-only blocks from parentChildren if such are there
			nonHeadersOnlyParentChildren := make([]*externalapi.DomainHash, 0, len(allParentChildren))
			for _, parentChild := range allParentChildren {
				if parentChild.Equal(model.VirtualBlockHash) {
					continue
				}

				parentChildStatus, err := csm.blockStatusStore.Get(csm.databaseContext, parentChild)
				if err != nil {
					return nil, err
				}
				if parentChildStatus == externalapi.StatusHeaderOnly {
					continue
				}
				nonHeadersOnlyParentChildren = append(nonHeadersOnlyParentChildren, parentChild)
			}
			log.Debugf("The non-virtual, non-headers-only children of block %s are: %s", parent, nonHeadersOnlyParentChildren)

			if disqualifiedCandidates.ContainsAllInSlice(nonHeadersOnlyParentChildren) {
				log.Debugf("The disqualified set contains all the "+
					"children of %s. Adding it to the candidate heap", nonHeadersOnlyParentChildren)
				err := candidatesHeap.Push(parent)
				if err != nil {
					return nil, err
				}
			}
		}
	}
}

// mergeSetIncrease returns different things depending on the result:
// If the candidate can be a virtual parent then canBeParent=true and mergeSetIncrease=The increase in merge set size
// If the candidate can't be a virtual parent, then canBeParent=false and newCandidate is a new proposed candidate in the past of candidate.
func (csm *consensusStateManager) mergeSetIncrease(candidate *externalapi.DomainHash, selectedVirtualParents []*externalapi.DomainHash, mergeSetSize uint64,
) (canBeParent bool, newCandidate *externalapi.DomainHash, mergeSetIncrease uint64, err error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "mergeSetIncrease")
	defer onEnd()

	visited := hashset.New()
	// Start with the parents in the queue as we already know the candidate isn't an ancestor of the parents.
	parents, err := csm.dagTopologyManager.Parents(candidate)
	if err != nil {
		return false, nil, 0, err
	}
	for _, parent := range parents {
		visited.Add(parent)
	}
	queue := parents
	mergeSetIncrease = uint64(1) // starts with 1 for the candidate itself

	var current *externalapi.DomainHash
	for len(queue) > 0 {
		current, queue = queue[0], queue[1:]
		log.Tracef("Attempting to increment the merge set size increase for block %s", current)

		isInPastOfSelectedVirtualParents, err := csm.dagTopologyManager.IsAncestorOfAny(current, selectedVirtualParents)
		if err != nil {
			return false, nil, 0, err
		}
		if isInPastOfSelectedVirtualParents {
			log.Tracef("Skipping block %s because it's in the past of one (or more) of the selected virtual parents", current)
			continue
		}

		log.Tracef("Incrementing the merge set size increase")
		mergeSetIncrease++

		if (mergeSetSize + mergeSetIncrease) > csm.mergeSetSizeLimit {
			log.Debugf("The merge set would increase by more than the limit with block %s", candidate)
			return false, current, mergeSetIncrease, nil
		}

		parents, err := csm.dagTopologyManager.Parents(current)
		if err != nil {
			return false, nil, 0, err
		}
		for _, parent := range parents {
			if !visited.Contains(parent) {
				visited.Add(parent)
				queue = append(queue, parent)
			}
		}
	}
	log.Debugf("The resolved merge set size increase is: %d", mergeSetIncrease)

	return true, nil, mergeSetIncrease, nil
}

func (csm *consensusStateManager) boundedMergeBreakingParents(
	parents []*externalapi.DomainHash) ([]*externalapi.DomainHash, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "boundedMergeBreakingParents")
	defer onEnd()

	log.Tracef("boundedMergeBreakingParents start for parents: %s", parents)

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

	var boundedMergeBreakingParents []*externalapi.DomainHash
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
			boundedMergeBreakingParents = append(boundedMergeBreakingParents, parent)
		}
	}

	return boundedMergeBreakingParents, nil
}
