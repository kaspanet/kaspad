package difficultymanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/bigintpool"
	"github.com/pkg/errors"
	"math"
	"math/big"
	"sort"
)

type difficultyBlock struct {
	selectedParent     *externalapi.DomainHash
	mergeSetBlues      []*externalapi.DomainHash
	timeInMilliseconds int64
	Bits               uint32
}
type blockWindow []difficultyBlock

func (dm *difficultyManager) getDifficultyBlock(blockHash *externalapi.DomainHash) (difficultyBlock, error) {
	ghostdagData, err := dm.ghostdagStore.Get(dm.databaseContext, blockHash)
	if err != nil {
		return difficultyBlock{}, err
	}

	header, err := dm.headerStore.BlockHeader(dm.databaseContext, blockHash)
	if err != nil {
		return difficultyBlock{}, err
	}
	return difficultyBlock{
		selectedParent:     ghostdagData.SelectedParent,
		mergeSetBlues:      ghostdagData.MergeSetBlues,
		timeInMilliseconds: header.TimeInMilliseconds,
		Bits:               header.Bits,
	}, nil
}

// blueBlockWindow returns a blockWindow of the given size that contains the
// blues in the past of startindNode, sorted by GHOSTDAG order.
// If the number of blues in the past of startingNode is less then windowSize,
// the window will be padded by genesis blocks to achieve a size of windowSize.
func (dm *difficultyManager) blueBlockWindow(startingNode *externalapi.DomainHash, windowSize uint64) (blockWindow, error) {
	window := make(blockWindow, 0, windowSize)
	currentNode, err := dm.getDifficultyBlock(startingNode)
	if err != nil {
		return nil, err
	}
	for uint64(len(window)) < windowSize && currentNode.selectedParent != nil {
		if currentNode.selectedParent != nil {
			for _, blue := range currentNode.mergeSetBlues {
				diffBlock, err := dm.getDifficultyBlock(blue)
				if err != nil {
					return nil, err
				}
				window = append(window, diffBlock)
				if uint64(len(window)) == windowSize {
					break
				}
			}
			spDiffBlock, err := dm.getDifficultyBlock(currentNode.selectedParent)
			if err != nil {
				return nil, err
			}
			currentNode = spDiffBlock
		}
	}

	if uint64(len(window)) < windowSize {
		genesis := currentNode
		for uint64(len(window)) < windowSize {
			window = append(window, genesis)
		}
	}
	return window, err
}

func (window blockWindow) minMaxTimestamps() (min, max int64) {
	min = math.MaxInt64
	max = 0
	for _, block := range window {
		if block.timeInMilliseconds < min {
			min = block.timeInMilliseconds
		}
		if block.timeInMilliseconds > max {
			max = block.timeInMilliseconds
		}
	}
	return
}

func (window blockWindow) averageTarget(averageTarget *big.Int) {
	averageTarget.SetInt64(0)

	target := bigintpool.Acquire(0)
	defer bigintpool.Release(target)
	for _, block := range window {
		util.CompactToBigWithDestination(block.Bits, target)
		averageTarget.Add(averageTarget, target)
	}

	windowLen := bigintpool.Acquire(int64(len(window)))
	defer bigintpool.Release(windowLen)
	averageTarget.Div(averageTarget, windowLen)
}

func (window blockWindow) medianTimestamp() (int64, error) {
	if len(window) == 0 {
		return 0, errors.New("Cannot calculate median timestamp for an empty block window")
	}
	timestamps := make([]int64, len(window))
	for i, node := range window {
		timestamps[i] = node.timeInMilliseconds
	}
	sort.Sort(timeSorter(timestamps))
	return timestamps[len(timestamps)/2], nil
}
