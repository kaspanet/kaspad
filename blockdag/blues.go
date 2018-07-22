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
	queue := newBlockHeap(blockHeapDirectionDown)
	queue.Push(block)
	past := newSet()
	depths := map[*blockNode]int{
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

// digToChainStart digs through the chain and returns the block in depth k+1
func (dag *BlockDAG) digToChainStart(block *blockNode, parent *model.Block) *model.Block {
	current := parent

	for i := 0; i < p.K; i++ {
		if current.IsGenesis() {
			break
		}
		current = current.SelectedParent
	}

	return current
}

func (p *Phantom) blueCandidates(chainStart *model.Block, past model.BlockSet) model.BlockSet {
	candidates := model.NewSet()
	candidates.Add(chainStart)

	queue := []*model.Block{chainStart}
	for len(queue) > 0 {
		var current *model.Block
		current, queue = queue[0], queue[1:]

		children := current.Children
		for child := range children {
			if !candidates.Contains(child) && past.Contains(child) {
				candidates.Add(child)
				queue = append(queue, child)
			}
		}
	}

	return candidates
}

func (p *Phantom) traverseCandidates(newBlock *model.Block, candidates model.BlockSet, selectedParent *model.Block) []*model.Block {
	blues := []*model.Block{}
	selectedParentPast := model.NewSet()
	queue := model.NewHeap(model.HeapDirectionDown)
	visited := model.NewSet()

	for parent := range newBlock.Parents {
		queue.Push(parent)
	}

	for queue.Len() > 0 {
		current := queue.Pop()
		if candidates.Contains(current) {
			if current == selectedParent || selectedParentPast.AnyChildInSet(current) {
				selectedParentPast.Add(current)
			} else {
				blues = append(blues, current)
			}
			for parent := range current.Parents {
				if !visited.Contains(parent) {
					visited.Add(parent)
					queue.Push(parent)
				}
			}
		}
	}

	return append(blues, selectedParent)
}
