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

func futureSizes(subDAG *model.SubDAG) map[externalapi.DomainHash]uint64 {
	futureSizes := make(map[externalapi.DomainHash]uint64)
	for _, tipHash := range subDAG.TipHashes {
		futureSizes[*tipHash] = 0
	}

	queue := append([]*externalapi.DomainHash{}, subDAG.TipHashes...)
	visited := hashset.New()
	for len(queue) > 0 {
		var blockHash *externalapi.DomainHash
		blockHash, queue = queue[0], queue[1:]
		visited.Add(blockHash)

		// Calculate the future size of blockHash
		block := subDAG.Blocks[*blockHash]
		blockFutureSize := uint64(0)
		for _, childHash := range block.ChildHashes {
			childFutureSize := futureSizes[*childHash]
			blockFutureSize += childFutureSize
		}
		futureSizes[*blockHash] = blockFutureSize + 1 // The "1" represents the current block

		// Add the block's parents to the queue
		for _, parentHash := range block.ParentHashes {
			if visited.Contains(parentHash) {
				continue
			}
			queue = append(queue, parentHash)
		}
	}
	return futureSizes
}

func blockHashWithLargestFutureSize(futureSizes map[externalapi.DomainHash]uint64, blockHashes []*externalapi.DomainHash) *externalapi.DomainHash {
	var blockHashWithLargestFutureSize *externalapi.DomainHash
	largestFutureSize := uint64(0)
	for _, blockHash := range blockHashes {
		blockFutureSize := futureSizes[*blockHash]
		if blockHashWithLargestFutureSize == nil || blockFutureSize > largestFutureSize ||
			(blockFutureSize == largestFutureSize && blockHashWithLargestFutureSize.Less(blockHash)) {
			largestFutureSize = blockFutureSize
			blockHashWithLargestFutureSize = blockHash
		}
	}
	return blockHashWithLargestFutureSize
}
