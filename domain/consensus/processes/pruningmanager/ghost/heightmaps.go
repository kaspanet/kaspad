package ghost

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
)

type heightMaps struct {
	blockHashToHeightMap   map[externalapi.DomainHash]uint64
	heightToBlockHashesMap map[uint64][]*externalapi.DomainHash
	maxHeight              uint64
}

func buildHeightMaps(subDAG *model.SubDAG) *heightMaps {
	blockHashToHeightMap := make(map[externalapi.DomainHash]uint64, len(subDAG.Blocks))
	heightToBlockHashesMap := make(map[uint64][]*externalapi.DomainHash)
	maxHeight := uint64(0)

	queue := append([]*externalapi.DomainHash{}, subDAG.RootHashes...)
	addedToQueue := hashset.NewFromSlice(subDAG.RootHashes...)
	for len(queue) > 0 {
		var currentBlockHash *externalapi.DomainHash
		currentBlockHash, queue = queue[0], queue[1:]

		// Send the block to the back of the queue if one or more of its parents had not been processed yet
		currentBlock := subDAG.Blocks[*currentBlockHash]
		hasMissingParentData := false
		for _, parentHash := range currentBlock.ParentHashes {
			if _, ok := blockHashToHeightMap[*parentHash]; !ok {
				hasMissingParentData = true
				continue
			}
		}
		if hasMissingParentData {
			queue = append(queue, currentBlockHash)
			continue
		}

		for _, childHash := range currentBlock.ChildHashes {
			if addedToQueue.Contains(childHash) {
				continue
			}
			queue = append(queue, childHash)
			addedToQueue.Add(childHash)
		}

		currentBlockHeight := uint64(0)
		if len(currentBlock.ParentHashes) > 0 {
			highestParentHeight := uint64(0)
			for _, parentHash := range currentBlock.ParentHashes {
				parentHeight := blockHashToHeightMap[*parentHash]
				if parentHeight > highestParentHeight {
					highestParentHeight = parentHeight
				}
			}
			currentBlockHeight = highestParentHeight + 1
		}
		blockHashToHeightMap[*currentBlockHash] = currentBlockHeight

		if _, ok := heightToBlockHashesMap[currentBlockHeight]; !ok {
			heightToBlockHashesMap[currentBlockHeight] = []*externalapi.DomainHash{}
		}
		heightToBlockHashesMap[currentBlockHeight] = append(heightToBlockHashesMap[currentBlockHeight], currentBlockHash)

		if currentBlockHeight > maxHeight {
			maxHeight = currentBlockHeight
		}
	}
	return &heightMaps{
		blockHashToHeightMap:   blockHashToHeightMap,
		heightToBlockHashesMap: heightToBlockHashesMap,
		maxHeight:              maxHeight,
	}
}
