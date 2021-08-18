package ghost

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
)

// GHOST calculates the GHOST chain for the given `subDAG`
func GHOST(subDAG *model.SubDAG) []*externalapi.DomainHash {
	futureSizes := futureSizes(subDAG)

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
	return ghostChain
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

func futureSizes(subDAG *model.SubDAG) map[externalapi.DomainHash]uint64 {
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

		populateReverseMergeSet(subDAG, currentBlock, reverseMergeSets)
		currentBlockReverseMergeSet := reverseMergeSets[*currentBlockHash]
		currentBlockReverseMergeSetSize := currentBlockReverseMergeSet.Length()
		futureSize := uint64(currentBlockReverseMergeSetSize)
		if currentBlockReverseMergeSet.Length() > 0 {
			selectedChild := currentBlock.ChildHashes[0]
			selectedChildFutureSize := futureSizes[*selectedChild]
			futureSize += selectedChildFutureSize
		}
		futureSizes[*currentBlockHash] = futureSize
	}
	return futureSizes
}

func populateReverseMergeSet(subDAG *model.SubDAG, block *model.SubDAGBlock, reverseMergeSets map[externalapi.DomainHash]hashset.HashSet) {
	if len(block.ChildHashes) == 0 {
		reverseMergeSets[*block.BlockHash] = hashset.New()
		return
	}

	selectedChild := block.ChildHashes[0]
	reverseMergeSet := hashset.NewFromSlice(selectedChild)

	queue := append([]*externalapi.DomainHash{}, block.ChildHashes...)
	addedToQueue := hashset.NewFromSlice(block.ChildHashes...)
	for len(queue) > 0 {
		var currentBlockHash *externalapi.DomainHash
		currentBlockHash, queue = queue[0], queue[1:]

		if isDescendantOf(subDAG, currentBlockHash, selectedChild) {
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
	reverseMergeSets[*block.BlockHash] = reverseMergeSet
}

func isDescendantOf(subDAG *model.SubDAG, blockAHash *externalapi.DomainHash, blockBHash *externalapi.DomainHash) bool {
	if blockAHash.Equal(blockBHash) {
		return true
	}

	blockB := subDAG.Blocks[*blockBHash]

	queue := append([]*externalapi.DomainHash{}, blockB.ChildHashes...)
	addedToQueue := hashset.NewFromSlice(blockB.ChildHashes...)
	for len(queue) > 0 {
		var currentBlockHash *externalapi.DomainHash
		currentBlockHash, queue = queue[0], queue[1:]

		if currentBlockHash.Equal(blockAHash) {
			return true
		}

		currentBlock := subDAG.Blocks[*currentBlockHash]
		for _, childHash := range currentBlock.ChildHashes {
			if addedToQueue.Contains(childHash) {
				continue
			}
			queue = append(queue, childHash)
			addedToQueue.Add(childHash)
		}
	}
	return false
}
