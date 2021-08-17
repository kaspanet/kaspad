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
	for _, block := range subDAG.Blocks {
		futureSizes[*block.BlockHash] = futureSize(subDAG, block.BlockHash)
	}
	return futureSizes
}

func futureSize(subDAG *model.SubDAG, blockHash *externalapi.DomainHash) uint64 {
	queue := []*externalapi.DomainHash{blockHash}
	addedToQueue := hashset.NewFromSlice(blockHash)
	futureSize := uint64(0)
	for len(queue) > 0 {
		futureSize++

		var currentBlockHash *externalapi.DomainHash
		currentBlockHash, queue = queue[0], queue[1:]

		currentBlock := subDAG.Blocks[*currentBlockHash]
		for _, childHash := range currentBlock.ChildHashes {
			if addedToQueue.Contains(childHash) {
				continue
			}
			queue = append(queue, childHash)
			addedToQueue.Add(childHash)
		}
	}
	return futureSize
}
