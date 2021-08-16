package ghostmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
)

// GHOST calculates and returns the GHOST chain above `lowHash`
func (gm *ghostManager) GHOST(stagingArea *model.StagingArea,
	lowHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {

	futureSizes, err := gm.futureSizes(stagingArea, lowHash)
	if err != nil {
		return nil, err
	}
	ghostChain := []*externalapi.DomainHash{}

	currentHash := lowHash
	for !currentHash.Equal(model.VirtualBlockHash) {
		ghostChain = append(ghostChain, currentHash)

		childHashes, err := gm.dagTopologyManager.Children(stagingArea, currentHash)
		if err != nil {
			return nil, err
		}

		largestFutureSize := uint64(0)
		var childHashWithLargestFutureSize *externalapi.DomainHash
		for _, childHash := range childHashes {
			childFutureSize := futureSizes[*childHash]
			if childHashWithLargestFutureSize == nil || childFutureSize > largestFutureSize {
				largestFutureSize = childFutureSize
				childHashWithLargestFutureSize = childHash
			}
		}
		currentHash = childHashWithLargestFutureSize
	}
	return ghostChain, nil
}

func (gm *ghostManager) futureSizes(stagingArea *model.StagingArea,
	lowHash *externalapi.DomainHash) (map[externalapi.DomainHash]uint64, error) {

	tips, err := gm.consensusStateStore.Tips(stagingArea, gm.databaseContext)
	if err != nil {
		return nil, err
	}
	futureSizes := make(map[externalapi.DomainHash]uint64)
	for _, tip := range tips {
		futureSizes[*tip] = 0
	}

	queue := gm.dagTraversalManager.NewDownHeap(stagingArea)
	err = queue.PushSlice(tips)
	if err != nil {
		return nil, err
	}
	visited := hashset.New()
	for queue.Len() > 0 {
		blockHash := queue.Pop()
		visited.Add(blockHash)

		// We are only interested in blocks that are in the future of lowHash
		isLowHashAncestorOfBlockHash, err := gm.dagTopologyManager.IsAncestorOf(stagingArea, lowHash, blockHash)
		if err != nil {
			return nil, err
		}
		if !isLowHashAncestorOfBlockHash {
			continue
		}

		// Calculate the future size of blockHash
		childHashes, err := gm.dagTopologyManager.Children(stagingArea, blockHash)
		if err != nil {
			return nil, err
		}
		blockFutureSize := uint64(0)
		for _, childHash := range childHashes {
			if childHash.Equal(model.VirtualBlockHash) {
				continue
			}
			childFutureSize := futureSizes[*childHash]
			blockFutureSize += childFutureSize
		}
		futureSizes[*blockHash] = blockFutureSize + 1 // The "1" represents the current block

		// Add the block's parents to the queue
		parentHashes, err := gm.dagTopologyManager.Parents(stagingArea, blockHash)
		if err != nil {
			return nil, err
		}
		for _, parentHash := range parentHashes {
			if visited.Contains(parentHash) {
				continue
			}
			err = queue.Push(parentHash)
			if err != nil {
				return nil, err
			}
		}
	}
	return futureSizes, nil
}
