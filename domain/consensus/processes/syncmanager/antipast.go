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

	lowBlockGHOSTDAGData, err := sm.ghostdagDataStore.Get(sm.databaseContext, stagingArea, lowHash, false)
	if err != nil {
		return nil, nil, err
	}
	highBlockGHOSTDAGData, err := sm.ghostdagDataStore.Get(sm.databaseContext, stagingArea, highHash, false)
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
		lowBlockGHOSTDAGData, err := sm.ghostdagDataStore.Get(sm.databaseContext, stagingArea, lowHash, false)
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
