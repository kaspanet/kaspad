package blockdag

import (
	"github.com/daglabs/btcd/dagconfig/daghash"
)

// phantom calculates and returns the block's blue set, selected parent and blue score
func phantom(block *blockNode, k uint32) (blues []*blockNode, selectedParent *blockNode, score uint64) {
	bestScore := uint64(0)
	var bestParent *blockNode
	var bestBlues []*blockNode
	var bestHash *daghash.Hash
	for _, parent := range block.parents {
		chainStart := digToChainStart(parent, k)
		candidates := blueCandidates(chainStart)
		blues := traverseCandidates(block, candidates, parent)
		score := uint64(len(blues)) + parent.blueScore

		if score > bestScore || (score == bestScore && (bestHash == nil || daghash.Less(bestHash, &parent.hash))) {
			bestScore = score
			bestBlues = blues
			bestParent = parent
			bestHash = &parent.hash
		}
	}

	return bestBlues, bestParent, bestScore
}

// digToChainStart digs through the chain and returns the block in depth k+1
func digToChainStart(parent *blockNode, k uint32) *blockNode {
	current := parent

	for i := uint32(0); i < k; i++ {
		if current.isGenesis() {
			break
		}
		current = current.selectedParent
	}

	return current
}

func blueCandidates(chainStart *blockNode) blockSet {
	candidates := newSet()
	candidates.add(chainStart)

	queue := []*blockNode{chainStart}
	for len(queue) > 0 {
		var current *blockNode
		current, queue = queue[0], queue[1:]

		children := current.children
		for _, child := range children {
			if !candidates.contains(child) {
				candidates.add(child)
				queue = append(queue, child)
			}
		}
	}

	return candidates
}

func traverseCandidates(newBlock *blockNode, candidates blockSet, selectedParent *blockNode) []*blockNode {
	blues := []*blockNode{}
	selectedParentPast := newSet()
	queue := NewHeap()
	visited := newSet()

	for _, parent := range newBlock.parents {
		queue.Push(parent)
	}

	for queue.Len() > 0 {
		current := queue.Pop()
		if candidates.contains(current) {
			if current == selectedParent || selectedParentPast.anyChildInSet(current) {
				selectedParentPast.add(current)
			} else {
				blues = append(blues, current)
			}
			for _, parent := range current.parents {
				if !visited.contains(parent) {
					visited.add(parent)
					queue.Push(parent)
				}
			}
		}
	}

	return append(blues, selectedParent)
}
