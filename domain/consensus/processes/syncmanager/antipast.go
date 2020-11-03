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
		var isCurrentAncestorOfLowNode bool
		if current == lowHash {
			isCurrentAncestorOfLowNode = false
		} else {
			var err error
			isCurrentAncestorOfLowNode, err = sm.dagTopologyManager.IsAncestorOf(current, lowHash)
			if err != nil {
				return nil, err
			}
		}
		if isCurrentAncestorOfLowNode {
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
	panic("implement me")
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
