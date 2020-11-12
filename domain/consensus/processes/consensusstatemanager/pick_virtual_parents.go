package consensusstatemanager

import (
	"errors"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
)

func (csm *consensusStateManager) pickVirtualParents(tips []*externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
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

	selectedVirtualParents := hashset.NewFromSlice(virtualSelectedParent)

	mergeSetSize := 1 // starts counting from 1 because selectedParent is already in the mergeSet

	for candidatesHeap.Len() > 0 && len(selectedVirtualParents) < constants.MaxBlockParents {
		candidate := candidatesHeap.Pop()
		mergeSetIncrease, err := csm.mergeSetIncrease(candidate, selectedVirtualParents)
		if err != nil {
			return nil, err
		}

		if mergeSetSize+mergeSetIncrease > constants.MergeSetSizeLimit {
			continue
		}

		selectedVirtualParents.Add(candidate)
		mergeSetSize += mergeSetIncrease
	}

	boundedMergeBreakingParents, err := csm.boundedMergeBreakingParents(selectedVirtualParents.ToSlice())
	if err != nil {
		return nil, err
	}

	return selectedVirtualParents.Subtract(boundedMergeBreakingParents).ToSlice(), nil
}

func (csm *consensusStateManager) selectVirtualSelectedParent(candidatesHeap model.BlockHeap) (*externalapi.DomainHash, error) {
	disqualifiedCandidates := hashset.New()

	for {
		if candidatesHeap.Len() == 0 {
			return nil, errors.New("virtual has no valid parent candidates")
		}
		selectedParentCandidate := candidatesHeap.Pop()

		selectedParentCandidateStatus, err := csm.blockStatusStore.Get(csm.databaseContext, selectedParentCandidate)
		if err != nil {
			return nil, err
		}
		if selectedParentCandidateStatus == externalapi.StatusValid {
			return selectedParentCandidate, nil
		}

		disqualifiedCandidates.Add(selectedParentCandidate)

		candidateParents, err := csm.dagTopologyManager.Parents(selectedParentCandidate)
		if err != nil {
			return nil, err
		}
		for _, parent := range candidateParents {
			parentChildren, err := csm.dagTopologyManager.Children(parent)
			if err != nil {
				return nil, err
			}

			if disqualifiedCandidates.ContainsAllInSlice(parentChildren) {
				err := candidatesHeap.Push(parent)
				if err != nil {
					return nil, err
				}
			}
		}
	}
}

func (csm *consensusStateManager) mergeSetIncrease(
	candidate *externalapi.DomainHash, selectedVirtualParents hashset.HashSet) (int, error) {

	visited := hashset.New()
	queue := csm.dagTraversalManager.NewDownHeap()
	err := queue.Push(candidate)
	if err != nil {
		return 0, err
	}
	mergeSetIncrease := 1 // starts with 1 for the candidate itself

	for queue.Len() > 0 {
		current := queue.Pop()
		isInPastOfSelectedVirtualParents, err := csm.dagTopologyManager.IsAncestorOfAny(
			current, selectedVirtualParents.ToSlice())
		if err != nil {
			return 0, err
		}
		if isInPastOfSelectedVirtualParents {
			continue
		}

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

	return mergeSetIncrease, nil
}

func (csm *consensusStateManager) boundedMergeBreakingParents(parents []*externalapi.DomainHash) (hashset.HashSet, error) {
	// Temporarily set virtual to all parents, so that we can run ghostdag on it
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

	virtualGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	virtualFinalityPoint, err := csm.virtualFinalityPoint(virtualGHOSTDAGData)
	if err != nil {
		return nil, err
	}

	badReds := []*externalapi.DomainHash{}
	for _, redBlock := range virtualGHOSTDAGData.MergeSetReds {
		isFinalityPointInPast, err := csm.dagTopologyManager.IsAncestorOf(virtualFinalityPoint, redBlock)
		if err != nil {
			return nil, err
		}
		if isFinalityPointInPast {
			continue
		}

		isKosherized := false
		for _, potentiallyKosherizingBlock := range potentiallyKosherizingBlocks {
			isKosherized, err = csm.dagTopologyManager.IsAncestorOf(redBlock, potentiallyKosherizingBlock)
			if err != nil {
				return nil, err
			}
			if isKosherized {
				break
			}
		}
		if !isKosherized {
			badReds = append(badReds, redBlock)
		}
	}

	boundedMergeBreakingParents := hashset.New()
	for _, parent := range parents {
		isBadRedInPast := false
		for _, badRedBlock := range badReds {
			isBadRedInPast, err = csm.dagTopologyManager.IsAncestorOf(parent, badRedBlock)
			if err != nil {
				return nil, err
			}
			if isBadRedInPast {
				break
			}
		}
		if isBadRedInPast {
			boundedMergeBreakingParents.Add(parent)
		}
	}

	return boundedMergeBreakingParents, nil
}
