package blockdag

import (
	"sort"

	"github.com/kaspanet/kaspad/domain/blocknode"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/pkg/errors"
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
func (dag *BlockDAG) ghostdag(newNode *blocknode.Node) (selectedParentAnticone []*blocknode.Node, err error) {
	newNode.SelectedParent = newNode.Parents.Bluest()
	newNode.BluesAnticoneSizes[newNode.SelectedParent] = 0
	newNode.Blues = []*blocknode.Node{newNode.SelectedParent}
	selectedParentAnticone, err = dag.selectedParentAnticone(newNode)
	if err != nil {
		return nil, err
	}

	sort.Slice(selectedParentAnticone, func(i, j int) bool {
		return selectedParentAnticone[i].Less(selectedParentAnticone[j])
	})

	for _, blueCandidate := range selectedParentAnticone {
		isBlue, candidateAnticoneSize, candidateBluesAnticoneSizes, err := dag.checkBlueCandidate(newNode, blueCandidate)
		if err != nil {
			return nil, err
		}

		if isBlue {
			// No k-cluster violation found, we can now set the candidate block as blue
			newNode.Blues = append(newNode.Blues, blueCandidate)
			newNode.BluesAnticoneSizes[blueCandidate] = candidateAnticoneSize
			for blue, blueAnticoneSize := range candidateBluesAnticoneSizes {
				newNode.BluesAnticoneSizes[blue] = blueAnticoneSize + 1
			}
		} else {
			newNode.Reds = append(newNode.Reds, blueCandidate)
		}
	}

	newNode.BlueScore = newNode.SelectedParent.BlueScore + uint64(len(newNode.Blues))

	return selectedParentAnticone, nil
}

func (dag *BlockDAG) checkBlueCandidate(newNode *blocknode.Node, blueCandidate *blocknode.Node) (
	isBlue bool, candidateAnticoneSize dagconfig.KType, candidateBluesAnticoneSizes map[*blocknode.Node]dagconfig.KType,
	err error) {

	// The maximum length of node.blues can be K+1 because
	// it contains the selected parent.
	if dagconfig.KType(len(newNode.Blues)) == dag.Params.K+1 {
		return false, 0, nil, nil
	}

	candidateBluesAnticoneSizes = make(map[*blocknode.Node]dagconfig.KType)

	// Iterate over all blocks in the blue set of newNode that are not in the past
	// of blueCandidate, and check for each one of them if blueCandidate potentially
	// enlarges their blue anticone to be over K, or that they enlarge the blue anticone
	// of blueCandidate to be over K.
	for chainBlock := newNode; ; chainBlock = chainBlock.SelectedParent {
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
				return false, 0, nil, err
			} else if isAncestorOfBlueCandidate {
				break
			}
		}

		for _, block := range chainBlock.Blues {
			// Skip blocks that exist in the past of blueCandidate.
			if isAncestorOfBlueCandidate, err := dag.isInPast(block, blueCandidate); err != nil {
				return false, 0, nil, err
			} else if isAncestorOfBlueCandidate {
				continue
			}

			candidateBluesAnticoneSizes[block], err = dag.blueAnticoneSize(block, newNode)
			if err != nil {
				return false, 0, nil, err
			}
			candidateAnticoneSize++

			if candidateAnticoneSize > dag.Params.K {
				// k-cluster violation: The candidate's blue anticone exceeded k
				return false, 0, nil, nil
			}

			if candidateBluesAnticoneSizes[block] == dag.Params.K {
				// k-cluster violation: A block in candidate's blue anticone already
				// has k blue blocks in its own anticone
				return false, 0, nil, nil
			}

			// This is a sanity check that validates that a blue
			// block's blue anticone is not already larger than K.
			if candidateBluesAnticoneSizes[block] > dag.Params.K {
				return false, 0, nil, errors.New("found blue anticone size larger than k")
			}
		}
	}

	return true, candidateAnticoneSize, candidateBluesAnticoneSizes, nil
}

// selectedParentAnticone returns the blocks in the anticone of the selected parent of the given node.
// The function work as follows.
// We start by adding all parents of the node (other than the selected parent) to a process queue.
// For each node in the queue:
//   we check whether it is in the past of the selected parent.
//   If not, we add the node to the resulting anticone-set and queue it for processing.
func (dag *BlockDAG) selectedParentAnticone(node *blocknode.Node) ([]*blocknode.Node, error) {
	anticoneSet := blocknode.NewSet()
	var anticoneSlice []*blocknode.Node
	selectedParentPast := blocknode.NewSet()
	var queue []*blocknode.Node
	// Queueing all parents (other than the selected parent itself) for processing.
	for parent := range node.Parents {
		if parent == node.SelectedParent {
			continue
		}
		anticoneSet.Add(parent)
		anticoneSlice = append(anticoneSlice, parent)
		queue = append(queue, parent)
	}
	for len(queue) > 0 {
		var current *blocknode.Node
		current, queue = queue[0], queue[1:]
		// For each parent of a the current node we check whether it is in the past of the selected parent. If not,
		// we add the it to the resulting anticone-set and queue it for further processing.
		for parent := range current.Parents {
			if anticoneSet.Contains(parent) || selectedParentPast.Contains(parent) {
				continue
			}
			isAncestorOfSelectedParent, err := dag.isInPast(parent, node.SelectedParent)
			if err != nil {
				return nil, err
			}
			if isAncestorOfSelectedParent {
				selectedParentPast.Add(parent)
				continue
			}
			anticoneSet.Add(parent)
			anticoneSlice = append(anticoneSlice, parent)
			queue = append(queue, parent)
		}
	}
	return anticoneSlice, nil
}

// blueAnticoneSize returns the blue anticone size of 'block' from the worldview of 'context'.
// Expects 'block' to be in the blue set of 'context'
func (dag *BlockDAG) blueAnticoneSize(block, context *blocknode.Node) (dagconfig.KType, error) {
	for current := context; current != nil; current = current.SelectedParent {
		if blueAnticoneSize, ok := current.BluesAnticoneSizes[block]; ok {
			return blueAnticoneSize, nil
		}
	}
	return 0, errors.Errorf("block %s is not in blue set of %s", block.Hash, context.Hash)
}
