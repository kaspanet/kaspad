package blockdag

import (
	"github.com/daglabs/btcd/util"
	"math"
	"math/big"
	"sort"
)

type blockWindow []*blockNode

func blueBlockWindow(startingNode *blockNode, windowSize uint64, padWithGenesis bool) (window blockWindow, ok bool) {
	window = make(blockWindow, 0, windowSize)
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
		} else {
			if !padWithGenesis {
				return nil, false
			}
			window = append(window, currentNode)
		}
	}

	if uint64(len(window)) < windowSize {
		if !padWithGenesis {
			return nil, false
		}
		genesis := currentNode
		for uint64(len(window)) < windowSize {
			window = append(window, genesis)
		}
	}

	return window, true
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

func (window blockWindow) averageTarget() *big.Int {
	averageTarget := big.NewInt(0)
	for _, node := range window {
		target := util.CompactToBig(node.bits)
		averageTarget.Add(averageTarget, target)
	}
	return averageTarget.Div(averageTarget, big.NewInt(int64(len(window))))
}

func (window blockWindow) medianTimestamp() int64 {
	timestamps := make([]int64, len(window))
	for i, node := range window {
		timestamps[i] = node.timestamp
	}
	sort.Sort(timeSorter(timestamps))
	return timestamps[len(timestamps)/2]
}
