package difficultymanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/util/difficulty"
	"math"
	"math/big"
)

type difficultyBlock struct {
	timeInMilliseconds int64
	Bits               uint32
	hash               *externalapi.DomainHash
	blueWork           *big.Int
}

type blockWindow []difficultyBlock

func (dm *difficultyManager) getDifficultyBlock(
	stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (difficultyBlock, error) {

	header, err := dm.headerStore.BlockHeader(dm.databaseContext, stagingArea, blockHash)
	if err != nil {
		return difficultyBlock{}, err
	}
	return difficultyBlock{
		timeInMilliseconds: header.TimeInMilliseconds(),
		Bits:               header.Bits(),
		hash:               blockHash,
		blueWork:           header.BlueWork(),
	}, nil
}

// blockWindow returns a blockWindow of the given size that contains the
// blocks in the past of startingNode, the sorting is unspecified.
// If the number of blocks in the past of startingNode is less then windowSize,
// the window will be padded by genesis blocks to achieve a size of windowSize.
func (dm *difficultyManager) blockWindow(stagingArea *model.StagingArea, startingNode *externalapi.DomainHash, windowSize int) (blockWindow,
	[]*externalapi.DomainHash, error) {

	window := make(blockWindow, 0, windowSize)
	windowHashes, err := dm.dagTraversalManager.BlockWindow(stagingArea, startingNode, windowSize)
	if err != nil {
		return nil, nil, err
	}

	for _, hash := range windowHashes {
		block, err := dm.getDifficultyBlock(stagingArea, hash)
		if err != nil {
			return nil, nil, err
		}
		window = append(window, block)
	}
	return window, windowHashes, nil
}

func (window blockWindow) minMaxTimestamps() (min, max int64, minIndex int) {
	min = math.MaxInt64
	minIndex = 0
	max = 0
	for i, block := range window {

		if block.timeInMilliseconds < min ||
			(block.timeInMilliseconds == min && block.blueWork.Cmp(window[minIndex].blueWork) < 0) ||
			(block.timeInMilliseconds == min && block.blueWork.Cmp(window[minIndex].blueWork) == 0 && block.hash.Less(window[minIndex].hash)) {
			min = block.timeInMilliseconds
			minIndex = i
		}
		if block.timeInMilliseconds > max {
			max = block.timeInMilliseconds
		}
	}
	return
}

func (window *blockWindow) remove(n int) {
	(*window)[n] = (*window)[len(*window)-1]
	*window = (*window)[:len(*window)-1]
}

func (window blockWindow) averageTarget() *big.Int {
	averageTarget := new(big.Int)
	targetTmp := new(big.Int)
	for _, block := range window {
		difficulty.CompactToBigWithDestination(block.Bits, targetTmp)
		averageTarget.Add(averageTarget, targetTmp)
	}
	return averageTarget.Div(averageTarget, big.NewInt(int64(len(window))))
}
