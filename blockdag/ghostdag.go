package blockdag

import (
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/pkg/errors"
	"sort"
)

// ghostdag runs the GHOSTDAG protocol and updates newNode.blues,
// newNode.selectedParent and newNode.bluesAnticoneSizes accordingly.
// The function updates newNode.blues by iterating over the blocks in
// the anticone of newNode.selectedParent (which is the parent with the
// highest blue score) and adds any block to newNode.blues if by adding
// it to newNode.blues these conditions will not be violated:
//
// 1) |anticone-of-candidate-block ∩ blue-set-of-newNode| ≤ K
//
// 2) For every blue block in blue-set-of-newNode:
//    |(anticone-of-blue-block ∩ blue-set-newNode) ∪ {candidate-block}| ≤ K.
//    We validate this condition by maintaining a map bluesAnticoneSizes for
//    each block which holds all the blue anticone sizes that were affected by
//    the new added blue blocks.
//    So to find out what is |anticone-of-blue ∩ blue-set-of-newNode| we just iterate in
//    the selected parent chain of newNode until we find an existing entry in
//    bluesAnticoneSizes.
//
// For further details see the article https://eprint.iacr.org/2018/104.pdf
func (dag *BlockDAG) ghostdag(newNode *blockNode) (selectedParentAnticone []*blockNode, err error) {
	newNode.selectedParent = newNode.parents.bluest()
	newNode.bluesAnticoneSizes[newNode.selectedParent] = 0
	newNode.blues = []*blockNode{newNode.selectedParent}
	selectedParentAnticone, err = dag.selectedParentAnticone(newNode)
	if err != nil {
		return nil, err
	}

	sort.Slice(selectedParentAnticone, func(i, j int) bool {
		return selectedParentAnticone[i].less(selectedParentAnticone[j])
	})

	for _, blueCandidate := range selectedParentAnticone {
		candidateBluesAnticoneSizes := make(map[*blockNode]dagconfig.KType)
		var candidateAnticoneSize dagconfig.KType
		possiblyBlue := true

		// Iterate over all blocks in the blue set of newNode that are not in the past
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
				if isAncestorOfBlueCandidate, err := dag.isInPast(chainBlock, blueCandidate); err != nil {
					return nil, err
				} else if isAncestorOfBlueCandidate {
					break
				}
			}

			for _, block := range chainBlock.blues {
				// Skip blocks that exist in the past of blueCandidate.
				if isAncestorOfBlueCandidate, err := dag.isInPast(block, blueCandidate); err != nil {
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
			newNode.bluesAnticoneSizes[blueCandidate] = candidateAnticoneSize
			for blue, blueAnticoneSize := range candidateBluesAnticoneSizes {
				newNode.bluesAnticoneSizes[blue] = blueAnticoneSize + 1
			}

			// The maximum length of node.blues can be K+1 because
			// it contains the selected parent.
			if dagconfig.KType(len(newNode.blues)) == dag.dagParams.K+1 {
				break
			}
		}
	}

	newNode.blueScore = newNode.selectedParent.blueScore + uint64(len(newNode.blues))
	return selectedParentAnticone, nil
}

// selectedParentAnticone returns the blocks in the anticone of the selected parent of the given node.
// The function work as follows.
// We start by adding all parents of the node (other than the selected parent) to a process queue.
// For each node in the queue:
//   we check whether it is in the past of the selected parent.
//   If not, we add the node to the resulting anticone-set and queue it for processing.
func (dag *BlockDAG) selectedParentAnticone(node *blockNode) ([]*blockNode, error) {
	anticoneSet := newBlockSet()
	var anticoneSlice []*blockNode
	selectedParentPast := newBlockSet()
	var queue []*blockNode
	// Queueing all parents (other than the selected parent itself) for processing.
	for parent := range node.parents {
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
		// For each parent of a the current node we check whether it is in the past of the selected parent. If not,
		// we add the it to the resulting anticone-set and queue it for further processing.
		for parent := range current.parents {
			if anticoneSet.contains(parent) || selectedParentPast.contains(parent) {
				continue
			}
			isAncestorOfSelectedParent, err := dag.isInPast(parent, node.selectedParent)
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
// Expects 'block' to be in the blue set of 'context'
func (dag *BlockDAG) blueAnticoneSize(block, context *blockNode) (dagconfig.KType, error) {
	for current := context; current != nil; current = current.selectedParent {
		if blueAnticoneSize, ok := current.bluesAnticoneSizes[block]; ok {
			return blueAnticoneSize, nil
		}
	}
	return 0, errors.Errorf("block %s is not in blue set of %s", block.hash, context.hash)
}
