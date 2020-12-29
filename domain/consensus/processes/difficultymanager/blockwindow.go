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
	timeInMilliseconds int64
	Bits               uint32
}

type blockWindow []difficultyBlock

func (dm *difficultyManager) getDifficultyBlock(blockHash *externalapi.DomainHash) (difficultyBlock, error) {
	header, err := dm.headerStore.BlockHeader(dm.databaseContext, blockHash)
	if err != nil {
		return difficultyBlock{}, err
	}
	return difficultyBlock{
		timeInMilliseconds: header.TimeInMilliseconds(),
		Bits:               header.Bits(),
	}, nil
}

// blueBlockWindow returns a blockWindow of the given size that contains the
// blues in the past of startindNode, the sorting is unspecified.
// If the number of blues in the past of startingNode is less then windowSize,
// the window will be padded by genesis blocks to achieve a size of windowSize.
func (dm *difficultyManager) blueBlockWindow(startingNode *externalapi.DomainHash, windowSize int) (blockWindow, error) {
	window := make(blockWindow, 0, windowSize)
	windowHashes, err := dm.dagTraversalManager.BlueWindow(startingNode, windowSize)
	if err != nil {
		return nil, err
	}

	for _, hash := range windowHashes {
		block, err := dm.getDifficultyBlock(hash)
		if err != nil {
			return nil, err
		}
		window = append(window, block)
	}
	return window, nil
}

func (window blockWindow) minMaxTimestamps() (min, max int64, minIndex, maxIndex int) {
	min = math.MaxInt64
	minIndex = math.MaxInt64
	max = 0
	maxIndex = 0
	for i, block := range window {
		if block.timeInMilliseconds < min {
			min = block.timeInMilliseconds
			minIndex = i
		}
		if block.timeInMilliseconds > max {
			max = block.timeInMilliseconds
			maxIndex = i
		}
	}
	return
}

func (window *blockWindow) remove(n int) {
	(*window)[n] = (*window)[len(*window)-1]
	*window = (*window)[:len(*window)-1]
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
