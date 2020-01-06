package blockdag

import (
	"github.com/pkg/errors"
	"sort"
)

func (dag *BlockDAG) selectedParentAnticone(node *blockNode) ([]*blockNode, error) {
	anticoneSet := newSet()
	var anticoneSlice []*blockNode
	selectedParentPast := newSet()
	var queue []*blockNode
	for _, parent := range node.parents {
		if parent == node.selectedParent {
			continue
		}
		anticoneSet.add(parent)
		anticoneSlice = append(anticoneSlice, parent)
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
			anticoneSlice = append(anticoneSlice, parent)
			queue = append(queue, parent)
		}
	}
	return anticoneSlice, nil
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

// ghostdag runs the GHOSTDAG protocol and updates newNode.blues,
// newNode.selectedParent and newNode.bluesAnticoneSizes accordingly.
// The function updates newNode.blues by iterating over the blocks in
// the anticone of newNode.selectedParent (which is the parent with the
// highest blue score) and adds any block to newNode.blues if by adding
// it to newNode.blues these conditions will not be violated:
//
// 1) |anticone-of-candidate-block ∩ blueset-of-newNode| ≤ K
//
// 2) For every blue block in blueset-of-newNode:
//    |(anticone-of-blue-block ∩ blueset-newNode) ∪ {candidate-block}| ≤ K.
//    We validate this condition by maintaining a map bluesAnticoneSizes for
//    each block which holds all the blue anticone sizes that were affected by
//    the new added blue blocks.
//    So to find out what is |anticone-of-blue ∩ blueset-of-newNode| we just iterate in
//    the selected parent chain of newNode until we find an existing entry in
//    bluesAnticoneSizes.
//
// For further details see the article https://eprint.iacr.org/2018/104.pdf
func (dag *BlockDAG) ghostdag(newNode *blockNode) (selectedParentAnticone []*blockNode, err error) {
	newNode.selectedParent = newNode.parents.bluest()
	newNode.bluesAnticoneSizes[*newNode.hash] = 0
	newNode.blues = []*blockNode{newNode.selectedParent}
	selectedParentAnticone, err = dag.selectedParentAnticone(newNode)
	if err != nil {
		return nil, err
	}

	sort.Slice(selectedParentAnticone, func(i, j int) bool {
		return selectedParentAnticone[i].less(selectedParentAnticone[j])
	})

	for _, blueCandidate := range selectedParentAnticone {
		candidateBluesAnticoneSizes := make(map[*blockNode]uint32)
		var candidateAnticoneSize uint32
		possiblyBlue := true

		// Iterate over all blocks in the blueset of newNode that are not in the past
		// of blueCandidate, and check for each one of them if blueCandidate potentially
		// enlarges their blue anticone to be over K, or that they enlarge the blue anticone
		// of blueCandidate to be over K.
		for chainBlock := newNode; possiblyBlue; chainBlock = chainBlock.selectedParent {
			// If blueCandidate is in the future of chainBlock, it means
			// that all remaining blues are in the past of chainBlock and thus
			// in the past of blueCandidate. In this case we know for sure that
			// the anticone of blueCandidate will not exceed K, and we can mark
			// it as blue.
			//
			// newNode is always in the future of blueCandidate, so there's
			// no point in checking it.
			if chainBlock != newNode {
				if isAncestorOfBlueCandidate, err := dag.isAncestorOf(chainBlock, blueCandidate); err != nil {
					return nil, err
				} else if isAncestorOfBlueCandidate {
					break
				}
			}

			for _, block := range chainBlock.blues {
				// Skip blocks that exists in the past of blueCandidate.
				if isAncestorOfBlueCandidate, err := dag.isAncestorOf(block, blueCandidate); err != nil {
					return nil, err
				} else if isAncestorOfBlueCandidate {
					continue
				}

				candidateBluesAnticoneSizes[block], err = dag.blueAnticoneSize(block, newNode)
				if err != nil {
					return nil, err
				}
				candidateAnticoneSize++

				if candidateAnticoneSize > dag.dagParams.K {
					// k-cluster violation: The candidate's blue anticone exceeded k
					possiblyBlue = false
					break
				}

				if candidateBluesAnticoneSizes[block] == dag.dagParams.K {
					// k-cluster violation: A block in candidate's blue anticone already
					// has k blue blocks in its own anticone
					possiblyBlue = false
					break
				}

				// This is a sanity check that validates that a blue
				// block's blue anticone is not already larger than K.
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
