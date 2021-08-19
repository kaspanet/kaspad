package ghost

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
)

// GHOST calculates the GHOST chain for the given `subDAG`
func GHOST(subDAG *model.SubDAG) ([]*externalapi.DomainHash, error) {
	futureSizes, err := futureSizes(subDAG)
	if err != nil {
		return nil, err
	}

	ghostChain := []*externalapi.DomainHash{}
	dagRootHashWithLargestFutureSize := blockHashWithLargestFutureSize(futureSizes, subDAG.RootHashes)
	currentHash := dagRootHashWithLargestFutureSize
	for {
		ghostChain = append(ghostChain, currentHash)

		currentBlock := subDAG.Blocks[*currentHash]
		childHashes := currentBlock.ChildHashes
		if len(childHashes) == 0 {
			break
		}

		childHashWithLargestFutureSize := blockHashWithLargestFutureSize(futureSizes, childHashes)
		currentHash = childHashWithLargestFutureSize
	}
	return ghostChain, nil
}

func blockHashWithLargestFutureSize(futureSizes map[externalapi.DomainHash]uint64,
	blockHashes []*externalapi.DomainHash) *externalapi.DomainHash {

	var blockHashWithLargestFutureSize *externalapi.DomainHash
	largestFutureSize := uint64(0)
	for _, blockHash := range blockHashes {
		blockFutureSize := futureSizes[*blockHash]
		if blockHashWithLargestFutureSize == nil || blockFutureSize > largestFutureSize ||
			(blockFutureSize == largestFutureSize && blockHash.Less(blockHashWithLargestFutureSize)) {
			largestFutureSize = blockFutureSize
			blockHashWithLargestFutureSize = blockHash
		}
	}
	return blockHashWithLargestFutureSize
}

func futureSizes(subDAG *model.SubDAG) (map[externalapi.DomainHash]uint64, error) {
	heightMaps := buildHeightMaps(subDAG)
	ghostReachabilityManager, err := newGHOSTReachabilityManager(subDAG, heightMaps)
	if err != nil {
		return nil, err
	}

	futureSizes := make(map[externalapi.DomainHash]uint64, len(subDAG.Blocks))
	reverseMergeSets := make(map[externalapi.DomainHash]hashset.HashSet, len(subDAG.Blocks))

	queue := append([]*externalapi.DomainHash{}, subDAG.TipHashes...)
	addedToQueue := hashset.NewFromSlice(subDAG.TipHashes...)
	for len(queue) > 0 {
		var currentBlockHash *externalapi.DomainHash
		currentBlockHash, queue = queue[0], queue[1:]

		// Send the block to the back of the queue if one or more of its children had not been processed yet
		currentBlock := subDAG.Blocks[*currentBlockHash]
		hasMissingChildData := false
		for _, childHash := range currentBlock.ChildHashes {
			if _, ok := futureSizes[*childHash]; !ok {
				hasMissingChildData = true
				continue
			}
		}
		if hasMissingChildData {
			queue = append(queue, currentBlockHash)
			continue
		}

		for _, parentHash := range currentBlock.ParentHashes {
			if addedToQueue.Contains(parentHash) {
				continue
			}
			queue = append(queue, parentHash)
			addedToQueue.Add(parentHash)
		}

		currentBlockReverseMergeSet, err := calculateReverseMergeSet(subDAG, ghostReachabilityManager, currentBlock)
		if err != nil {
			return nil, err
		}
		reverseMergeSets[*currentBlockHash] = currentBlockReverseMergeSet

		currentBlockReverseMergeSetSize := currentBlockReverseMergeSet.Length()
		futureSize := uint64(currentBlockReverseMergeSetSize)
		if currentBlockReverseMergeSet.Length() > 0 {
			selectedChild := currentBlock.ChildHashes[0]
			selectedChildFutureSize := futureSizes[*selectedChild]
			futureSize += selectedChildFutureSize
		}
		futureSizes[*currentBlockHash] = futureSize
	}
	return futureSizes, nil
}

func calculateReverseMergeSet(subDAG *model.SubDAG,
	ghostReachabilityManager *ghostReachabilityManager, block *model.SubDAGBlock) (hashset.HashSet, error) {

	if len(block.ChildHashes) == 0 {
		return hashset.New(), nil
	}

	selectedChild := block.ChildHashes[0]
	reverseMergeSet := hashset.NewFromSlice(selectedChild)

	queue := append([]*externalapi.DomainHash{}, block.ChildHashes...)
	addedToQueue := hashset.NewFromSlice(block.ChildHashes...)
	for len(queue) > 0 {
		var currentBlockHash *externalapi.DomainHash
		currentBlockHash, queue = queue[0], queue[1:]

		isCurrentBlockDescendantOfSelectedChild, err := ghostReachabilityManager.isDescendantOf(currentBlockHash, selectedChild)
		if err != nil {
			return nil, err
		}
		if isCurrentBlockDescendantOfSelectedChild {
			continue
		}
		reverseMergeSet.Add(currentBlockHash)

		currentBlock := subDAG.Blocks[*currentBlockHash]
		for _, childHash := range currentBlock.ChildHashes {
			if addedToQueue.Contains(childHash) {
				continue
			}
			queue = append(queue, childHash)
			addedToQueue.Add(childHash)
		}
	}
	return reverseMergeSet, nil
}
