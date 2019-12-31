package blockdag

import "github.com/pkg/errors"

func (dag *BlockDAG) selectedParentAnticone(node *blockNode) (*blockHeap, error) {
	anticoneSet := newSet()
	anticoneHeap := newUpHeap()
	selectedParentPast := newSet()
	var queue []*blockNode
	for _, parent := range node.parents {
		if parent == node.selectedParent {
			continue
		}
		anticoneSet.add(parent)
		queue = append(queue, parent)
	}
	for len(queue) > 0 {
		var current *blockNode
		current, queue = queue[0], queue[1:]
		for _, parent := range current.parents {
			if anticoneSet.contains(parent) || selectedParentPast.contains(parent) {
				continue
			}
			isAncestorOfSelectedParent, err := dag.isAncestorOf(parent, node.selectedParent)
			if err != nil {
				return nil, err
			}
			if isAncestorOfSelectedParent {
				selectedParentPast.add(parent)
				continue
			}
			anticoneSet.add(parent)
			anticoneHeap.Push(parent)
			queue = append(queue, parent)
		}
	}
	return &anticoneHeap, nil
}

// blueAnticoneSize returns the blue anticone size of 'block' from the worldview of 'context'.
// Expects 'block' to be ∈ blue-set(context)
func (dag *BlockDAG) blueAnticoneSize(block, context *blockNode) (uint32, error) {
	for current := context; current != nil; current = current.selectedParent {
		if blueAnticoneSize, ok := current.bluesAnticoneSizes[*block.hash]; ok {
			return blueAnticoneSize, nil
		}
	}
	return 0, errors.Errorf("block %s is not in blue-set of %s", block.hash, context.hash)
}

// ghostdag updates newNode.blues, newNode.selectedParent
// and newNode.bluesAnticoneSizes according to the GHOSTDAG
// protocol.
// It updates newNode.blues by going over the anticone of
// newNode.selectedParent (which is the parent with the
// highest blue score) and adds it to newNode.blues if it
// passes two conditions:
// 1) |anticone(block) ∩ blueset(newNode)| <= K
// 2) For every blue in blueset(newNode) |anticone(blue) ∩ blueset(newNode) ∩ {block}| <= K.
//    We do this by maintaining for each block a map bluesAnticoneSizes which holds
//    all the blue anticone sizes that were affected by the new added blues.
//    So to find out what is |anticone(blue) ∩ blueset(newNode)| we just iterate in
//    the selected parent chain of newNode until we find an existing entry in
//    bluesAnticoneSizes.
//
// For further details see the article https://eprint.iacr.org/2018/104.pdf
func (dag *BlockDAG) ghostdag(newNode *blockNode) (selectedParentAnticone []*blockNode, err error) {
	newNode.selectedParent = newNode.parents.bluest()
	newNode.bluesAnticoneSizes[*newNode.hash] = 0
	newNode.blues = append(newNode.blues, newNode.selectedParent)
	selectedParentAnticoneHeap, err := dag.selectedParentAnticone(newNode)
	if err != nil {
		return nil, err
	}

	selectedParentAnticone = make([]*blockNode, selectedParentAnticoneHeap.Len())
	for selectedParentAnticoneHeap.Len() > 0 {
		blueCandidate := selectedParentAnticoneHeap.pop()
		selectedParentAnticone = append(selectedParentAnticone, blueCandidate)
		candidateBluesAnticoneSizes := make(map[*blockNode]uint32)
		var candidateAnticoneSize uint32
		possiblyBlue := true

		for chainBlock := newNode; possiblyBlue; chainBlock = chainBlock.selectedParent {
			// If blueCandidate is in the future of chainBlock, it means
			// that all remaining blues are in past(chainBlock) and thus
			// in past(blueCandidate). In this case we know for sure that
			// anticone(blueCandidate) will not exceed K, and we can mark
			// it as blue.
			//
			// newNode is always in the future of blueCandidate, so there's
			// no point in checking it.
			if chainBlock != newNode {
				if isAncestorOf, err := dag.isAncestorOf(chainBlock, blueCandidate); err != nil {
					return nil, err
				} else if isAncestorOf {
					break
				}
			}

			for _, block := range chainBlock.blues {
				// Skip blocks that exists in the past of blueCandidate.
				// We already checked it for chainBlock above, so if the
				// block is chainBlock, there's no need to recheck.
				if block != chainBlock {
					if isAncestorOf, err := dag.isAncestorOf(block, blueCandidate); err != nil {
						return nil, err
					} else if isAncestorOf {
						continue
					}
				}

				candidateBluesAnticoneSizes[block], err = dag.blueAnticoneSize(block, newNode)
				if err != nil {
					return nil, err
				}
				candidateAnticoneSize++

				if candidateAnticoneSize > dag.dagParams.K {
					// k-cluster violation: The candidate blue anticone now became larger than k
					possiblyBlue = false
					break
				}

				if candidateBluesAnticoneSizes[block] == dag.dagParams.K {
					// k-cluster violation: A block in candidate's blue anticone already
					// has k blue blocks in its own anticone
					possiblyBlue = false
					break
				}
				if candidateBluesAnticoneSizes[block] > dag.dagParams.K {
					return nil, errors.New("found blue anticone size larger than k")
				}
			}
		}

		if possiblyBlue {
			// No k-cluster violation found, we can now set the candidate block as blue
			newNode.blues = append(newNode.blues, blueCandidate)
			newNode.bluesAnticoneSizes[*blueCandidate.hash] = candidateAnticoneSize
			for blue, blueAnticoneSize := range candidateBluesAnticoneSizes {
				newNode.bluesAnticoneSizes[*blue.hash] = blueAnticoneSize + 1
			}

			// The maximum length of node.blues can be K+1 because
			// it contains the selected parent.
			if uint32(len(newNode.blues)) == dag.dagParams.K+1 {
				break
			}
		}
	}

	newNode.blueScore = newNode.selectedParent.blueScore + uint64(len(newNode.blues))
	return selectedParentAnticone, nil
}
