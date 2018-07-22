package blockdag

func (dag *BlockDAG) blues(block *blockNode) (blues []*blockNode, selectedParent *blockNode, score int64) {
	bestScore := int64(-1)
	var bestParent *blockNode
	var bestBlues []*blockNode
	past := dag.relevantPast(block)
	for _, parent := range block.parents {
		chainStart := dag.digToChainStart(block, parent)
		candidates := dag.blueCandidates(chainStart, past)
		blues := dag.traverseCandidates(block, candidates, parent)
		score := int64(len(blues)) + parent.blueScore

		if score > bestScore {
			bestScore = score
			bestBlues = blues
			bestParent = parent
		}
	}

	return bestBlues, bestParent, bestScore
}

// relevantPast returns all the past of block K blocks deep.
func (dag *BlockDAG) relevantPast(block *blockNode) blockSet {
	queue := newBlockHeap(blockHeapDirectionDown)
	queue.Push(block)
	past := newSet()
	depths := map[*blockNode]uint{
		block: 0,
	}

	for queue.Len() > 0 {
		current := queue.Pop()
		depth := depths[current]
		parentDepth := depth + 1
		if depth < dag.dagParams.K {
			for _, parent := range current.parents {
				if !past.contains(parent) {
					past.add(parent)
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
func (dag *BlockDAG) digToChainStart(block *blockNode, parent *blockNode) *blockNode {
	current := parent

	for i := uint(0); i < dag.dagParams.K; i++ {
		if current.isGenesis() {
			break
		}
		current = current.selectedParent
	}

	return current
}

func (dag *BlockDAG) blueCandidates(chainStart *blockNode, past blockSet) blockSet {
	candidates := newSet()
	candidates.add(chainStart)

	queue := []*blockNode{chainStart}
	for len(queue) > 0 {
		var current *blockNode
		current, queue = queue[0], queue[1:]

		children := current.children
		for _, child := range children {
			if !candidates.contains(child) && past.contains(child) {
				candidates.add(child)
				queue = append(queue, child)
			}
		}
	}

	return candidates
}

func (dag *BlockDAG) traverseCandidates(newBlock *blockNode, candidates blockSet, selectedParent *blockNode) []*blockNode {
	blues := []*blockNode{}
	selectedParentPast := newSet()
	queue := newBlockHeap(blockHeapDirectionDown)
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
