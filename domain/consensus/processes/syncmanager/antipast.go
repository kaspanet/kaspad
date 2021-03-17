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
func (sm *syncManager) antiPastHashesBetween(lowHash, highHash *externalapi.DomainHash,
	maxBlueScoreDifference uint64) (hashes []*externalapi.DomainHash, actualHighHash *externalapi.DomainHash, err error) {

	// If lowHash is not in the selectedParentChain of highHash - SelectedChildIterator will fail.
	// Therefore, we traverse down lowHash's selectedParentChain until we reach a block that is in
	// highHash's selectedParentChain.
	// We keep originalLowHash to filter out blocks in it's past later down the road
	originalLowHash := lowHash
	lowHash, err = sm.findLowHashInHighHashSelectedParentChain(lowHash, highHash)
	if err != nil {
		return nil, nil, err
	}

	lowBlockGHOSTDAGData, err := sm.ghostdagDataStore.Get(sm.databaseContext, lowHash)
	if err != nil {
		return nil, nil, err
	}
	highBlockGHOSTDAGData, err := sm.ghostdagDataStore.Get(sm.databaseContext, highHash)
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
		highHash, err = sm.findHighHashAccordingToMaxBlueScoreDifference(lowHash, highHash, maxBlueScoreDifference, highBlockGHOSTDAGData, lowBlockGHOSTDAGData)
		if err != nil {
			return nil, nil, err
		}
	}

	// Collect all hashes by concatenating the merge-sets of all blocks between highHash and lowHash
	blockHashes := []*externalapi.DomainHash{}
	iterator, err := sm.dagTraversalManager.SelectedChildIterator(highHash, lowHash)
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
		sortedMergeSet, err := sm.getSortedMergeSet(current)
		if err != nil {
			return nil, nil, err
		}

		// append to blockHashes all blocks in sortedMergeSet which are not in the past of originalLowHash
		for _, blockHash := range sortedMergeSet {
			isInPastOfOriginalLowHash, err := sm.dagTopologyManager.IsAncestorOf(blockHash, originalLowHash)
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

func (sm *syncManager) getSortedMergeSet(current *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	currentGhostdagData, err := sm.ghostdagDataStore.Get(sm.databaseContext, current)
	if err != nil {
		return nil, err
	}

	blueMergeSet := currentGhostdagData.MergeSetBlues()
	redMergeSet := currentGhostdagData.MergeSetReds()
	sortedMergeSet := make([]*externalapi.DomainHash, 0, len(blueMergeSet)+len(redMergeSet))
	selectedParent, blueMergeSet := blueMergeSet[0], blueMergeSet[1:]
	sortedMergeSet = append(sortedMergeSet, selectedParent)
	i, j := 0, 0
	for i < len(blueMergeSet) && j < len(redMergeSet) {
		currentBlue := blueMergeSet[i]
		currentBlueGhostdagData, err := sm.ghostdagDataStore.Get(sm.databaseContext, currentBlue)
		if err != nil {
			return nil, err
		}
		currentRed := redMergeSet[j]
		currentRedGhostdagData, err := sm.ghostdagDataStore.Get(sm.databaseContext, currentRed)
		if err != nil {
			return nil, err
		}
		if sm.ghostdagManager.Less(currentBlue, currentBlueGhostdagData, currentRed, currentRedGhostdagData) {
			sortedMergeSet = append(sortedMergeSet, currentBlue)
			i++
		} else {
			sortedMergeSet = append(sortedMergeSet, currentRed)
			j++
		}
	}
	sortedMergeSet = append(sortedMergeSet, blueMergeSet[i:]...)
	sortedMergeSet = append(sortedMergeSet, redMergeSet[j:]...)

	return sortedMergeSet, nil
}

func (sm *syncManager) findHighHashAccordingToMaxBlueScoreDifference(lowHash *externalapi.DomainHash,
	highHash *externalapi.DomainHash, maxBlueScoreDifference uint64, highBlockGHOSTDAGData *model.BlockGHOSTDAGData,
	lowBlockGHOSTDAGData *model.BlockGHOSTDAGData) (*externalapi.DomainHash, error) {

	if highBlockGHOSTDAGData.BlueScore()-lowBlockGHOSTDAGData.BlueScore() <= maxBlueScoreDifference {
		return highHash, nil
	}

	iterator, err := sm.dagTraversalManager.SelectedChildIterator(highHash, lowHash)
	if err != nil {
		return nil, err
	}
	defer iterator.Close()
	for ok := iterator.First(); ok; ok = iterator.Next() {
		highHashCandidate, err := iterator.Get()
		if err != nil {
			return nil, err
		}
		highBlockGHOSTDAGData, err = sm.ghostdagDataStore.Get(sm.databaseContext, highHashCandidate)
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

func (sm *syncManager) findLowHashInHighHashSelectedParentChain(
	lowHash *externalapi.DomainHash, highHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {
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
	return lowHash, nil
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
	defer selectedChildIterator.Close()

	lowHash := pruningPoint
	foundHeaderOnlyBlock := false
	for ok := selectedChildIterator.First(); ok; ok = selectedChildIterator.Next() {
		selectedChild, err := selectedChildIterator.Get()
		if err != nil {
			return nil, err
		}
		hasBlock, err := sm.blockStore.HasBlock(sm.databaseContext,, selectedChild)
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

	hashesBetween, _, err := sm.antiPastHashesBetween(lowHash, highHash, 0)
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
