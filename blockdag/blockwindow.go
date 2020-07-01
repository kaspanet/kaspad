package blockdag

import (
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/bigintpool"
	"github.com/pkg/errors"
	"math"
	"math/big"
	"sort"
)

type blockWindow []*blockNode

// blueBlockWindow returns a blockWindow of the given size that contains the
// blues in the past of startindNode, sorted by GHOSTDAG order.
// If the number of blues in the past of startingNode is less then windowSize,
// the window will be padded by genesis blocks to achieve a size of windowSize.
func blueBlockWindow(startingNode *blockNode, windowSize uint64) blockWindow {
	window := make(blockWindow, 0, windowSize)
	currentNode := startingNode
	for uint64(len(window)) < windowSize && currentNode.selectedParent != nil {
		if currentNode.selectedParent != nil {
			for _, blue := range currentNode.blues {
				window = append(window, blue)
				if uint64(len(window)) == windowSize {
					break
				}
			}
			currentNode = currentNode.selectedParent
		}
	}

	if uint64(len(window)) < windowSize {
		genesis := currentNode
		for uint64(len(window)) < windowSize {
			window = append(window, genesis)
		}
	}

	return window
}

func (window blockWindow) minMaxTimestamps() (min, max int64) {
	min = math.MaxInt64
	max = 0
	for _, node := range window {
		if node.timestamp < min {
			min = node.timestamp
		}
		if node.timestamp > max {
			max = node.timestamp
		}
	}
	return
}

func (window blockWindow) averageTarget(averageTarget *big.Int) {
	averageTarget.SetInt64(0)

	target := bigintpool.Acquire(0)
	defer bigintpool.Release(target)
	for _, node := range window {
		util.CompactToBigWithDestination(node.bits, target)
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
		timestamps[i] = node.timestamp
	}
	sort.Sort(timeSorter(timestamps))
	return timestamps[len(timestamps)/2], nil
}
