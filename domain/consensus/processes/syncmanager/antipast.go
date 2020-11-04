package syncmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
	"github.com/pkg/errors"
)

const maxHashesInAntiPastHashesBetween = 1 << 17

// antiPastHashesBetween returns the hashes of the blocks between the
// lowHash's antiPast and highHash's antiPast, or up to
// maxHashesInAntiPastHashesBetween.
func (sm *syncManager) antiPastHashesBetween(lowHash, highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	lowBlockGHOSTDAGData, err := sm.ghostdagDataStore.Get(sm.databaseContext, lowHash)
	if err != nil {
		return nil, err
	}
	lowBlockBlueScore := lowBlockGHOSTDAGData.BlueScore
	highBlockGHOSTDAGData, err := sm.ghostdagDataStore.Get(sm.databaseContext, highHash)
	if err != nil {
		return nil, err
	}
	highBlockBlueScore := highBlockGHOSTDAGData.BlueScore
	if lowBlockBlueScore >= highBlockBlueScore {
		return nil, errors.Errorf("low hash blueScore >= high hash blueScore (%d >= %d)",
			lowBlockBlueScore, highBlockBlueScore)
	}

	// In order to get no more then maxHashesInAntiPastHashesBetween
	// blocks from th future of the lowHash (including itself),
	// we iterate the selected parent chain of the highNode and
	// stop once we reach
	// highBlockBlueScore-lowBlockBlueScore+1 <= maxHashesInAntiPastHashesBetween.
	// That stop point becomes the new highHash.
	// Using blueScore as an approximation is considered to be
	// fairly accurate because we presume that most DAG blocks are
	// blue.
	for highBlockBlueScore-lowBlockBlueScore+1 > maxHashesInAntiPastHashesBetween {
		highHash = highBlockGHOSTDAGData.SelectedParent
	}

	// Collect every node in highHash's past (including itself) but
	// NOT in the lowHash's past (excluding itself) into an up-heap
	// (a heap sorted by blueScore from lowest to greatest).
	visited := hashset.New()
	candidateHashes := sm.dagTraversalManager.NewUpHeap()
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
		err = candidateHashes.Push(current)
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

	// Pop candidateHashes into a slice. Since candidateHashes is
	// an up-heap, it's guaranteed to be ordered from low to high
	hashesLength := maxHashesInAntiPastHashesBetween
	if candidateHashes.Len() < hashesLength {
		hashesLength = candidateHashes.Len()
	}
	hashes := make([]*externalapi.DomainHash, hashesLength)
	for i := 0; i < hashesLength; i++ {
		hashes[i] = candidateHashes.Pop()
	}
	return hashes, nil
}

func (sm *syncManager) missingBlockBodyHashes(highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	headerTipsPruningPoint, err := sm.consensusStateManager.HeaderTipsPruningPoint()
	if err != nil {
		return nil, err
	}

	selectedChildIterator, err := sm.dagTraversalManager.SelectedChildIterator(highHash, headerTipsPruningPoint)
	if err != nil {
		return nil, err
	}

	lowHash := headerTipsPruningPoint
	for selectedChildIterator.Next() {
		lowHash = selectedChildIterator.Get()
	}

	hashesBetween, err := sm.antiPastHashesBetween(lowHash, highHash)
	if err != nil {
		return nil, err
	}

	lowHashAnticone, err := sm.dagTraversalManager.AnticoneFromContext(highHash, lowHash)
	if err != nil {
		return nil, err
	}

	blockToRemoveFromHashesBetween := hashset.New()
	for _, blockHash := range lowHashAnticone {
		isHeaderOnlyBlock, err := sm.isHeaderOnlyBlock(blockHash)
		if err != nil {
			return nil, err
		}

		if !isHeaderOnlyBlock {
			blockToRemoveFromHashesBetween.Add(blockHash)
		}
	}

	missingBlocks := make([]*externalapi.DomainHash, 0, len(hashesBetween)-len(lowHashAnticone))
	for i, blockHash := range hashesBetween {
		if blockToRemoveFromHashesBetween.Contains(blockHash) {
			blockToRemoveFromHashesBetween.Remove(blockHash)
			if blockToRemoveFromHashesBetween.Length() == 0 && i != len(hashesBetween)-1 {
				missingBlocks = append(missingBlocks, hashesBetween[i+1:]...)
				break
			}
			continue
		}
		missingBlocks = append(missingBlocks, blockHash)
	}

	if blockToRemoveFromHashesBetween.Length() == 0 {
		return nil, errors.Errorf("blockToRemoveFromHashesBetween.Length() is expected to be 0")
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

func (sm *syncManager) isBlockInHeaderPruningPointFutureAndVirtualPast(blockHash *externalapi.DomainHash) (bool, error) {
	exists, err := sm.blockStatusStore.Exists(sm.databaseContext, blockHash)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}
	panic("implement me")
}
