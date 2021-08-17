package ghost

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
)

// GHOST calculates the GHOST chain for the given `subDAG`
func GHOST(subDAG *model.SubDAG) []*externalapi.DomainHash {
	ghostChain := []*externalapi.DomainHash{}
	dagRootHashWithLargestFutureSize := blockHashWithLargestFutureSize(subDAG, subDAG.RootHashes)
	currentHash := dagRootHashWithLargestFutureSize
	for {
		ghostChain = append(ghostChain, currentHash)

		currentBlock := subDAG.Blocks[*currentHash]
		childHashes := currentBlock.ChildHashes
		if len(childHashes) == 0 {
			break
		}

		childHashWithLargestFutureSize := blockHashWithLargestFutureSize(subDAG, childHashes)
		currentHash = childHashWithLargestFutureSize
	}
	return ghostChain
}

func blockHashWithLargestFutureSize(subDAG *model.SubDAG, blockHashes []*externalapi.DomainHash) *externalapi.DomainHash {
	var blockHashWithLargestFutureSize *externalapi.DomainHash
	largestFutureSize := uint64(0)
	for _, blockHash := range blockHashes {
		blockFutureSize := futureSize(subDAG, blockHash)
		if blockHashWithLargestFutureSize == nil || blockFutureSize > largestFutureSize ||
			(blockFutureSize == largestFutureSize && blockHashWithLargestFutureSize.Less(blockHash)) {
			largestFutureSize = blockFutureSize
			blockHashWithLargestFutureSize = blockHash
		}
	}
	return blockHashWithLargestFutureSize
}

func futureSize(subDAG *model.SubDAG, blockHash *externalapi.DomainHash) uint64 {
	queue := []*externalapi.DomainHash{blockHash}
	visited := hashset.New()
	futureSize := uint64(0)
	for len(queue) > 0 {
		futureSize++

		var currentBlockHash *externalapi.DomainHash
		currentBlockHash, queue = queue[0], queue[1:]
		visited.Add(currentBlockHash)

		currentBlock := subDAG.Blocks[*currentBlockHash]
		for _, childHash := range currentBlock.ChildHashes {
			if visited.Contains(childHash) {
				continue
			}
			queue = append(queue, childHash)
		}
	}
	return futureSize
}
