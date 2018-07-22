package blockdag

import "github.com/daglabs/phantompoc/model"

func (dag *BlockDAG) blues(block *blockNode) (blues []*blockNode, selectedParent *blockNode, score int) {
	bestScore := -1
	var bestParent *blockNode
	var bestBlues []*blockNode
	past := dag.relevantPast(block)
	for parent := range block.parents {
		chainStart := p.digToChainStart(block, parent)
		candidates := p.blueCandidates(chainStart, past)
		blues := p.traverseCandidates(block, candidates, parent)
		score := len(blues) + parent.blueScore

		if score > bestScore {
			bestScore = score
			bestBlues = blues
			bestParent = parent
		} else {
		}
	}

	return bestBlues, bestParent, bestScore
}

// relevantPast returns all the past of block K blocks deep.
func (dag *BlockDAG) relevantPast(block *blockNode) blockSet {
	queue := NewHeap(model.HeapDirectionDown)
	queue.Push(block)
	past := model.NewSet()
	depths := map[*model.Block]int{
		block: 0,
	}

	for queue.Len() > 0 {
		current := queue.Pop()
		depth := depths[current]
		parentDepth := depth + 1
		if depth < p.K {
			for parent := range current.Parents {
				if !past.Contains(parent) {
					past.Add(parent)
					queue.Push(parent)
					depths[parent] = parentDepth
				} else if previousDepth := depths[parent]; parentDepth < previousDepth {
					depths[parent] = parentDepth
					queue.Push(parent)
				}
			}
		}
	}

	return past
}
