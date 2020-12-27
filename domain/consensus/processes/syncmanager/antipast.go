package syncmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
	"github.com/pkg/errors"
)

// antiPastHashesBetween returns the hashes of the blocks between the
// lowHash's antiPast and highHash's antiPast.
func (sm *syncManager) antiPastHashesBetween(lowHash, highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
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
		err = hashesUpHeap.Push(current)
		if err != nil {
			return nil, err
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
	for selectedChildIterator.Next() {
		selectedChild := selectedChildIterator.Get()
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
		// TODO: Once block children are fixed, this error
		// should be returned instead of simply logged
		log.Errorf("no header-only blocks between %s and %s",
			lowHash, highHash)
	}

	hashesBetween, err := sm.antiPastHashesBetween(lowHash, highHash)
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
