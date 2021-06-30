package syncmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

// antiPastHashesBetween returns the hashes of the blocks between the
// lowHash's antiPast and highHash's antiPast, or up to `maxBlocks`, if non-zero.
// The result excludes lowHash and includes highHash. If lowHash == highHash, returns nothing.
// If maxBlocks != 0 then maxBlocks MUST be >= MergeSetSizeLimit + 1
// because it returns blocks with MergeSet granularity,
// so if MergeSet > maxBlocks, function will return nothing
func (sm *syncManager) antiPastHashesBetween(stagingArea *model.StagingArea, lowHash, highHash *externalapi.DomainHash,
	maxBlocks uint64) (hashes []*externalapi.DomainHash, actualHighHash *externalapi.DomainHash, err error) {

	// Sanity check, for debugging only
	if maxBlocks != 0 && maxBlocks < sm.mergeSetSizeLimit+1 {
		return nil, nil,
			errors.Errorf("maxBlocks (%d) MUST be >= MergeSetSizeLimit + 1 (%d)", maxBlocks, sm.mergeSetSizeLimit+1)
	}

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

		if maxBlocks != 0 && uint64(len(blockHashes)+len(sortedMergeSet)) > maxBlocks {
			break
		}

		highHash = current

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

func (sm *syncManager) missingBlockBodyHashes(stagingArea *model.StagingArea, highHash *externalapi.DomainHash) (
	[]*externalapi.DomainHash, error) {

	pruningPoint, err := sm.pruningStore.PruningPoint(sm.databaseContext, stagingArea)
	if err != nil {
		return nil, err
	}

	selectedChildIterator, err := sm.dagTraversalManager.SelectedChildIterator(stagingArea, highHash, pruningPoint)
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
		hasBlock, err := sm.blockStore.HasBlock(sm.databaseContext, stagingArea, selectedChild)
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

	hashesBetween, _, err := sm.antiPastHashesBetween(stagingArea, lowHash, highHash, 0)
	if err != nil {
		return nil, err
	}

	missingBlocks := make([]*externalapi.DomainHash, 0, len(hashesBetween))
	for _, blockHash := range hashesBetween {
		blockStatus, err := sm.blockStatusStore.Get(sm.databaseContext, stagingArea, blockHash)
		if err != nil {
			return nil, err
		}
		if blockStatus == externalapi.StatusHeaderOnly {
			missingBlocks = append(missingBlocks, blockHash)
		}
	}

	return missingBlocks, nil
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
