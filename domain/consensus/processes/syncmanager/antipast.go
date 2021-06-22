package syncmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

// antiPastHashesBetween returns the hashes of the blocks between the
// lowHash's antiPast and highHash's antiPast, or up to
// `maxBlueScoreDifference`, if non-zero.
// The result excludes lowHash and includes highHash. If lowHash == highHash, returns nothing.
func (sm *syncManager) antiPastHashesBetween(stagingArea *model.StagingArea, lowHash, highHash *externalapi.DomainHash,
	maxBlueScoreDifference uint64) (hashes []*externalapi.DomainHash, actualHighHash *externalapi.DomainHash, err error) {

	// If lowHash is not in the selectedParentChain of highHash - SelectedChildIterator will fail.
	// Therefore, we traverse down lowHash's selectedParentChain until we reach a block that is in
	// highHash's selectedParentChain.
	// We keep originalLowHash to filter out blocks in it's past later down the road
	originalLowHash := lowHash
	lowHash, err = sm.findLowHashInHighHashSelectedParentChain(stagingArea, lowHash, highHash)
	if err != nil {
		return nil, nil, err
	}

	lowBlockGHOSTDAGData, err := sm.ghostdagDataStore.Get(sm.databaseContext, stagingArea, lowHash)
	if err != nil {
		return nil, nil, err
	}
	highBlockGHOSTDAGData, err := sm.ghostdagDataStore.Get(sm.databaseContext, stagingArea, highHash)
	if err != nil {
		return nil, nil, err
	}
	if lowBlockGHOSTDAGData.BlueScore() > highBlockGHOSTDAGData.BlueScore() {
		return nil, nil, errors.Errorf("low hash blueScore > high hash blueScore (%d > %d)",
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
		highHash, err = sm.findHighHashAccordingToMaxBlueScoreDifference(stagingArea,
			lowHash, highHash, maxBlueScoreDifference, highBlockGHOSTDAGData, lowBlockGHOSTDAGData)
		if err != nil {
			return nil, nil, err
		}
	}

	// Collect all hashes by concatenating the merge-sets of all blocks between highHash and lowHash
	blockHashes := []*externalapi.DomainHash{}
	iterator, err := sm.dagTraversalManager.SelectedChildIterator(stagingArea, highHash, lowHash)
	if err != nil {
		return nil, nil, err
	}
	defer iterator.Close()
	for ok := iterator.First(); ok; ok = iterator.Next() {
		current, err := iterator.Get()
		if err != nil {
			return nil, nil, err
		}
		// Both blue and red merge sets are topologically sorted, but not the concatenation of the two.
		// We require the blocks to be topologically sorted. In addition,  for optimal performance,
		// we want the selectedParent to be first.
		// Since the rest of the merge set is in the anticone of selectedParent, it's position in the list does not
		// matter, even though it's blue score is the highest, we can arbitrarily decide it comes first.
		// Therefore we first append the selectedParent, then the rest of blocks in ghostdag order.
		sortedMergeSet, err := sm.ghostdagManager.GetSortedMergeSet(stagingArea, current)
		if err != nil {
			return nil, nil, err
		}

		// append to blockHashes all blocks in sortedMergeSet which are not in the past of originalLowHash
		for _, blockHash := range sortedMergeSet {
			isInPastOfOriginalLowHash, err := sm.dagTopologyManager.IsAncestorOf(stagingArea, blockHash, originalLowHash)
			if err != nil {
				return nil, nil, err
			}
			if isInPastOfOriginalLowHash {
				continue
			}
			blockHashes = append(blockHashes, blockHash)
		}
	}

	// The process above doesn't return highHash, so include it explicitly, unless highHash == lowHash
	if !lowHash.Equal(highHash) {
		blockHashes = append(blockHashes, highHash)
	}

	return blockHashes, highHash, nil
}

func (sm *syncManager) findHighHashAccordingToMaxBlueScoreDifference(stagingArea *model.StagingArea, lowHash *externalapi.DomainHash,
	highHash *externalapi.DomainHash, maxBlueScoreDifference uint64, highBlockGHOSTDAGData *externalapi.BlockGHOSTDAGData,
	lowBlockGHOSTDAGData *externalapi.BlockGHOSTDAGData) (*externalapi.DomainHash, error) {

	if highBlockGHOSTDAGData.BlueScore()-lowBlockGHOSTDAGData.BlueScore() <= maxBlueScoreDifference {
		return highHash, nil
	}

	iterator, err := sm.dagTraversalManager.SelectedChildIterator(stagingArea, highHash, lowHash)
	if err != nil {
		return nil, err
	}
	defer iterator.Close()
	for ok := iterator.First(); ok; ok = iterator.Next() {
		highHashCandidate, err := iterator.Get()
		if err != nil {
			return nil, err
		}
		highBlockGHOSTDAGData, err = sm.ghostdagDataStore.Get(sm.databaseContext, stagingArea, highHashCandidate)
		if err != nil {
			return nil, err
		}
		if highBlockGHOSTDAGData.BlueScore()-lowBlockGHOSTDAGData.BlueScore() > maxBlueScoreDifference {
			break
		}
		highHash = highHashCandidate
	}
	return highHash, nil
}

func (sm *syncManager) findLowHashInHighHashSelectedParentChain(stagingArea *model.StagingArea,
	lowHash *externalapi.DomainHash, highHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	for {
		isInSelectedParentChain, err := sm.dagTopologyManager.IsInSelectedParentChainOf(stagingArea, lowHash, highHash)
		if err != nil {
			return nil, err
		}
		if isInSelectedParentChain {
			break
		}
		lowBlockGHOSTDAGData, err := sm.ghostdagDataStore.Get(sm.databaseContext, stagingArea, lowHash)
		if err != nil {
			return nil, err
		}
		lowHash = lowBlockGHOSTDAGData.SelectedParent()
	}
	return lowHash, nil
}

func (sm *syncManager) isHeaderOnlyBlock(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (bool, error) {
	exists, err := sm.blockStatusStore.Exists(sm.databaseContext, stagingArea, blockHash)
	if err != nil {
		return false, err
	}

	if !exists {
		return false, nil
	}

	status, err := sm.blockStatusStore.Get(sm.databaseContext, stagingArea, blockHash)
	if err != nil {
		return false, err
	}

	return status == externalapi.StatusHeaderOnly, nil
}
