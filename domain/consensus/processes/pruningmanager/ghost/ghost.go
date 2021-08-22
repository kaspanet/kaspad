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

	height := heightMaps.maxHeight
	for {
		for _, blockHash := range heightMaps.heightToBlockHashesMap[height] {
			block := subDAG.Blocks[*blockHash]
			currentBlockReverseMergeSetSize, err := calculateReverseMergeSetSize(subDAG, ghostReachabilityManager, block)
			if err != nil {
				return nil, err
			}

			futureSize := currentBlockReverseMergeSetSize
			if currentBlockReverseMergeSetSize > 0 {
				selectedChild := block.ChildHashes[0]
				selectedChildFutureSize := futureSizes[*selectedChild]
				futureSize += selectedChildFutureSize
			}
			futureSizes[*blockHash] = futureSize
		}
		if height == 0 {
			break
		}
		height--
	}
	return futureSizes, nil
}

func calculateReverseMergeSetSize(subDAG *model.SubDAG,
	ghostReachabilityManager *ghostReachabilityManager, block *model.SubDAGBlock) (uint64, error) {

	if len(block.ChildHashes) == 0 {
		return 0, nil
	}

	selectedChild := block.ChildHashes[0]
	reverseMergeSetSize := uint64(1)

	knownSelectedChildDescendants := hashset.NewFromSlice(selectedChild)

	queue := append([]*externalapi.DomainHash{}, block.ChildHashes...)
	addedToQueue := hashset.NewFromSlice(block.ChildHashes...)
	for len(queue) > 0 {
		var currentBlockHash *externalapi.DomainHash
		currentBlockHash, queue = queue[0], queue[1:]

		currentBlock := subDAG.Blocks[*currentBlockHash]
		if knownSelectedChildDescendants.Contains(currentBlockHash) {
			for _, childHash := range currentBlock.ChildHashes {
				knownSelectedChildDescendants.Add(childHash)
			}
			continue
		}

		isCurrentBlockDescendantOfSelectedChild, err := ghostReachabilityManager.isDescendantOf(currentBlockHash, selectedChild)
		if err != nil {
			return 0, err
		}
		if isCurrentBlockDescendantOfSelectedChild {
			knownSelectedChildDescendants.Add(currentBlockHash)
			for _, childHash := range currentBlock.ChildHashes {
				knownSelectedChildDescendants.Add(childHash)
			}
			continue
		}
		reverseMergeSetSize++

		for _, childHash := range currentBlock.ChildHashes {
			if addedToQueue.Contains(childHash) {
				continue
			}
			queue = append(queue, childHash)
			addedToQueue.Add(childHash)
		}
	}
	return reverseMergeSetSize, nil
}
