package syncmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
	"github.com/pkg/errors"
)

// antiPastHashesBetween returns the hashes of the blocks between the
// lowHash's antiPast and highHash's antiPast, or up to
// `maxBlueScoreDifference`, if non-zero.
func (sm *syncManager) antiPastHashesBetween(lowHash, highHash *externalapi.DomainHash,
	maxBlueScoreDifference uint64) ([]*externalapi.DomainHash, error) {

	// If lowHash is not in the selectedParentChain of highHash - SelectedChildIterator will fail.
	// Therefore, we traverse down lowHash's selectedParentChain until we reach a block that is in
	// highHash's selectedParentChain.
	// We keep originalLowHash to filter out blocks in it's past later down the road
	originalLowHash := lowHash
	for {
		isInSelectedParentChain, err := sm.dagTopologyManager.IsInSelectedParentChainOf(lowHash, highHash)
		if err != nil {
			return nil, err
		}
		if isInSelectedParentChain {
			break
		}
		lowBlockGHOSTDAGData, err := sm.ghostdagDataStore.Get(sm.databaseContext, lowHash)
		if err != nil {
			return nil, err
		}
		lowHash = lowBlockGHOSTDAGData.SelectedParent()
	}

	lowBlockGHOSTDAGData, err := sm.ghostdagDataStore.Get(sm.databaseContext, lowHash)
	if err != nil {
		return nil, err
	}
	highBlockGHOSTDAGData, err := sm.ghostdagDataStore.Get(sm.databaseContext, highHash)
	if err != nil {
		return nil, err
	}
	if lowBlockGHOSTDAGData.BlueScore() >= highBlockGHOSTDAGData.BlueScore() {
		return nil, errors.Errorf("low hash blueScore >= high hash blueScore (%d >= %d)",
			lowBlockGHOSTDAGData.BlueScore(), highBlockGHOSTDAGData.BlueScore())
	}

	if maxBlueScoreDifference != 0 {
		// In order to get no more then maxBlueScoreDifference
		// blocks from the future of the lowHash (including itself),
		// we iterate the selected parent chain of the highNode and
		// stop once we reach
		// highBlockBlueScore-lowBlockBlueScore+1 <= maxBlueScoreDifference.
		// That stop point becomes the new highHash.
		// Using blueScore as an approximation is considered to be
		// fairly accurate because we presume that most DAG blocks are
		// blue.
		iterator, err := sm.dagTraversalManager.SelectedChildIterator(highHash, lowHash)
		if err != nil {
			return nil, err
		}
		for ok := iterator.First(); ok; ok = iterator.Next() {
			highHash, err = iterator.Get()
			if err != nil {
				return nil, err
			}
			highBlockGHOSTDAGData, err = sm.ghostdagDataStore.Get(sm.databaseContext, highHash)
			if err != nil {
				return nil, err
			}
			if highBlockGHOSTDAGData.BlueScore()-lowBlockGHOSTDAGData.BlueScore()+1 > maxBlueScoreDifference {
				break
			}
		}
	}

	// Collect every node in highHash's past (including itself) but
	// NOT in the lowHash's past (excluding itself) into an up-heap
	// (a heap sorted by blueScore from lowest to greatest).
	visited := hashset.New()
	hashesUpHeap := sm.dagTraversalManager.NewUpHeap()
	queue := sm.dagTraversalManager.NewDownHeap()
	err = queue.Push(highHash)
	if err != nil {
		return nil, err
	}
	for queue.Len() > 0 {
		current := queue.Pop()
		if visited.Contains(current) {
			continue
		}
		visited.Add(current)
		var isCurrentAncestorOfLowHash bool
		if current == lowHash {
			isCurrentAncestorOfLowHash = false
		} else {
			var err error
			isCurrentAncestorOfLowHash, err = sm.dagTopologyManager.IsAncestorOf(current, lowHash)
			if err != nil {
				return nil, err
			}
		}
		if isCurrentAncestorOfLowHash {
			continue
		}
		// Push current to hashesUpHeap if it's not in the past of originalLowHash
		isInPastOfOriginalLowHash, err := sm.dagTopologyManager.IsAncestorOf(current, originalLowHash)
		if err != nil {
			return nil, err
		}
		if !isInPastOfOriginalLowHash {
			err = hashesUpHeap.Push(current)
			if err != nil {
				return nil, err
			}
		}
		parents, err := sm.dagTopologyManager.Parents(current)
		if err != nil {
			return nil, err
		}
		for _, parent := range parents {
			err := queue.Push(parent)
			if err != nil {
				return nil, err
			}
		}
	}

	return hashesUpHeap.ToSlice(), nil
}

func (sm *syncManager) missingBlockBodyHashes(highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	pruningPoint, err := sm.pruningStore.PruningPoint(sm.databaseContext)
	if err != nil {
		return nil, err
	}

	selectedChildIterator, err := sm.dagTraversalManager.SelectedChildIterator(highHash, pruningPoint)
	if err != nil {
		return nil, err
	}

	lowHash := pruningPoint
	foundHeaderOnlyBlock := false
	for ok := selectedChildIterator.First(); ok; ok = selectedChildIterator.Next() {
		selectedChild, err := selectedChildIterator.Get()
		if err != nil {
			return nil, err
		}
		hasBlock, err := sm.blockStore.HasBlock(sm.databaseContext, selectedChild)
		if err != nil {
			return nil, err
		}

		if !hasBlock {
			foundHeaderOnlyBlock = true
			break
		}
		lowHash = selectedChild
	}
	if !foundHeaderOnlyBlock {
		if lowHash == highHash {
			// Blocks can be inserted inside the DAG during IBD if those were requested before IBD started.
			// In rare cases, all the IBD blocks might be already inserted by the time we reach this point.
			// In these cases - return an empty list of blocks to sync
			return []*externalapi.DomainHash{}, nil
		}
		// TODO: Once block children are fixed (https://github.com/kaspanet/kaspad/issues/1499),
		// this error should be returned rather the logged
		log.Errorf("no header-only blocks between %s and %s",
			lowHash, highHash)
	}

	hashesBetween, err := sm.antiPastHashesBetween(lowHash, highHash, 0)
	if err != nil {
		return nil, err
	}

	missingBlocks := make([]*externalapi.DomainHash, 0, len(hashesBetween))
	for _, blockHash := range hashesBetween {
		blockStatus, err := sm.blockStatusStore.Get(sm.databaseContext, blockHash)
		if err != nil {
			return nil, err
		}
		if blockStatus == externalapi.StatusHeaderOnly {
			missingBlocks = append(missingBlocks, blockHash)
		}
	}

	return missingBlocks, nil
}

func (sm *syncManager) isHeaderOnlyBlock(blockHash *externalapi.DomainHash) (bool, error) {
	exists, err := sm.blockStatusStore.Exists(sm.databaseContext, blockHash)
	if err != nil {
		return false, err
	}

	if !exists {
		return false, nil
	}

	status, err := sm.blockStatusStore.Get(sm.databaseContext, blockHash)
	if err != nil {
		return false, err
	}

	return status == externalapi.StatusHeaderOnly, nil
}
