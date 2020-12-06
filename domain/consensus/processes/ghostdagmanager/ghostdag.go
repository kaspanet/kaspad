package ghostdagmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
	"math/big"
)

// GHOSTDAG runs the GHOSTDAG protocol and calculates the block BlockGHOSTDAGData by the given parents.
// The function calculates MergeSetBlues by iterating over the blocks in
// the anticone of the new block selected parent (which is the parent with the
// highest blue score) and adds any block to newNode.blues if by adding
// it to MergeSetBlues these conditions will not be violated:
//
// 1) |anticone-of-candidate-block ∩ blue-set-of-newBlock| ≤ K
//
// 2) For every blue block in blue-set-of-newBlock:
//    |(anticone-of-blue-block ∩ blue-set-newBlock) ∪ {candidate-block}| ≤ K.
//    We validate this condition by maintaining a map BluesAnticoneSizes for
//    each block which holds all the blue anticone sizes that were affected by
//    the new added blue blocks.
//    So to find out what is |anticone-of-blue ∩ blue-set-of-newBlock| we just iterate in
//    the selected parent chain of the new block until we find an existing entry in
//    BluesAnticoneSizes.
//
// For further details see the article https://eprint.iacr.org/2018/104.pdf
func (gm *ghostdagManager) GHOSTDAG(blockHash *externalapi.DomainHash) error {
	newBlockData := &blockGHOSTDAGData{
		blueWork:           new(big.Int),
		mergeSetBlues:      make([]*externalapi.DomainHash, 0),
		mergeSetReds:       make([]*externalapi.DomainHash, 0),
		bluesAnticoneSizes: make(map[externalapi.DomainHash]model.KType),
	}

	blockParents, err := gm.dagTopologyManager.Parents(blockHash)
	if err != nil {
		return err
	}

	isGenesis := len(blockParents) == 0
	if !isGenesis {
		selectedParent, err := gm.findSelectedParent(blockParents)
		if err != nil {
			return err
		}

		newBlockData.selectedParent = selectedParent
		newBlockData.mergeSetBlues = append(newBlockData.mergeSetBlues, selectedParent)
		newBlockData.bluesAnticoneSizes[*selectedParent] = 0
	}

	mergeSetWithoutSelectedParent, err := gm.mergeSetWithoutSelectedParent(newBlockData.selectedParent, blockParents)
	if err != nil {
		return err
	}

	for _, blueCandidate := range mergeSetWithoutSelectedParent {
		isBlue, candidateAnticoneSize, candidateBluesAnticoneSizes, err := gm.checkBlueCandidate(newBlockData, blueCandidate)
		if err != nil {
			return err
		}

		if isBlue {
			// No k-cluster violation found, we can now set the candidate block as blue
			newBlockData.mergeSetBlues = append(newBlockData.mergeSetBlues, blueCandidate)
			newBlockData.bluesAnticoneSizes[*blueCandidate] = candidateAnticoneSize
			for blue, blueAnticoneSize := range candidateBluesAnticoneSizes {
				newBlockData.bluesAnticoneSizes[blue] = blueAnticoneSize + 1
			}
		} else {
			newBlockData.mergeSetReds = append(newBlockData.mergeSetReds, blueCandidate)
		}
	}

	if !isGenesis {
		selectedParentGHOSTDAGData, err := gm.ghostdagDataStore.Get(gm.databaseContext, newBlockData.selectedParent)
		if err != nil {
			return err
		}
		newBlockData.blueScore = selectedParentGHOSTDAGData.BlueScore() + uint64(len(newBlockData.mergeSetBlues))
		// We inherit the bluework from the selected parent
		newBlockData.blueWork.Set(selectedParentGHOSTDAGData.BlueWork())
		// Then we add up all the *work*(not blueWork) that all of newBlock merge set blues and selected parent did
		for _, blue := range newBlockData.mergeSetBlues {
			header, err := gm.headerStore.BlockHeader(gm.databaseContext, blue)
			if err != nil {
				return err
			}
			newBlockData.blueWork.Add(newBlockData.blueWork, util.CalcWork(header.Bits))
		}
	} else {
		// Genesis's blue score is defined to be 0.
		newBlockData.blueScore = 0
		newBlockData.blueWork.SetUint64(0)
	}

	gm.ghostdagDataStore.Stage(blockHash, newBlockData)

	return nil
}

type chainBlockData struct {
	hash      *externalapi.DomainHash
	blockData model.BlockGHOSTDAGData
}

func (gm *ghostdagManager) checkBlueCandidate(newBlockData *blockGHOSTDAGData, blueCandidate *externalapi.DomainHash) (
	isBlue bool, candidateAnticoneSize model.KType, candidateBluesAnticoneSizes map[externalapi.DomainHash]model.KType, err error) {

	// The maximum length of node.blues can be K+1 because
	// it contains the selected parent.
	if model.KType(len(newBlockData.mergeSetBlues)) == gm.k+1 {
		return false, 0, nil, nil
	}

	candidateBluesAnticoneSizes = make(map[externalapi.DomainHash]model.KType, gm.k)

	// Iterate over all blocks in the blue set of newNode that are not in the past
	// of blueCandidate, and check for each one of them if blueCandidate potentially
	// enlarges their blue anticone to be over K, or that they enlarge the blue anticone
	// of blueCandidate to be over K.
	chainBlock := chainBlockData{
		blockData: newBlockData,
	}

	for {
		isBlue, isRed, err := gm.checkBlueCandidateWithChainBlock(newBlockData, chainBlock, blueCandidate, candidateBluesAnticoneSizes,
			&candidateAnticoneSize)
		if err != nil {
			return false, 0, nil, err
		}

		if isBlue {
			break
		}

		if isRed {
			return false, 0, nil, nil
		}

		selectedParentGHOSTDAGData, err := gm.ghostdagDataStore.Get(gm.databaseContext, chainBlock.blockData.SelectedParent())
		if err != nil {
			return false, 0, nil, err
		}

		chainBlock = chainBlockData{hash: chainBlock.blockData.SelectedParent(),
			blockData: selectedParentGHOSTDAGData,
		}
	}

	return true, candidateAnticoneSize, candidateBluesAnticoneSizes, nil
}

func (gm *ghostdagManager) checkBlueCandidateWithChainBlock(newBlockData model.BlockGHOSTDAGData,
	chainBlock chainBlockData, blueCandidate *externalapi.DomainHash,
	candidateBluesAnticoneSizes map[externalapi.DomainHash]model.KType,
	candidateAnticoneSize *model.KType) (isBlue, isRed bool, err error) {

	// If blueCandidate is in the future of chainBlock, it means
	// that all remaining blues are in the past of chainBlock and thus
	// in the past of blueCandidate. In this case we know for sure that
	// the anticone of blueCandidate will not exceed K, and we can mark
	// it as blue.
	//
	// The new block is always in the future of blueCandidate, so there's
	// no point in checking it.

	// We check if chainBlock is not the new block by checking if it has a hash.
	if chainBlock.hash != nil {
		isAncestorOfBlueCandidate, err := gm.dagTopologyManager.IsAncestorOf(chainBlock.hash, blueCandidate)
		if err != nil {
			return false, false, err
		}
		if isAncestorOfBlueCandidate {
			return true, false, nil
		}
	}

	for _, block := range chainBlock.blockData.MergeSetBlues() {
		// Skip blocks that exist in the past of blueCandidate.
		isAncestorOfBlueCandidate, err := gm.dagTopologyManager.IsAncestorOf(block, blueCandidate)
		if err != nil {
			return false, false, err
		}

		if isAncestorOfBlueCandidate {
			continue
		}

		candidateBluesAnticoneSizes[*block], err = gm.blueAnticoneSize(block, newBlockData)
		if err != nil {
			return false, false, err
		}
		*candidateAnticoneSize++

		if *candidateAnticoneSize > gm.k {
			// k-cluster violation: The candidate's blue anticone exceeded k
			return false, true, nil
		}

		if candidateBluesAnticoneSizes[*block] == gm.k {
			// k-cluster violation: A block in candidate's blue anticone already
			// has k blue blocks in its own anticone
			return false, true, nil
		}

		// This is a sanity check that validates that a blue
		// block's blue anticone is not already larger than K.
		if candidateBluesAnticoneSizes[*block] > gm.k {
			return false, false, errors.New("found blue anticone size larger than k")
		}
	}

	return false, false, nil
}

// blueAnticoneSize returns the blue anticone size of 'block' from the worldview of 'context'.
// Expects 'block' to be in the blue set of 'context'
func (gm *ghostdagManager) blueAnticoneSize(block *externalapi.DomainHash, context model.BlockGHOSTDAGData) (model.KType, error) {
	for current := context; current != nil; {
		if blueAnticoneSize, ok := current.BluesAnticoneSizes()[*block]; ok {
			return blueAnticoneSize, nil
		}
		if current.SelectedParent() == nil {
			break
		}
		var err error
		current, err = gm.ghostdagDataStore.Get(gm.databaseContext, current.SelectedParent())
		if err != nil {
			return 0, err
		}
	}
	return 0, errors.Errorf("block %s is not in blue set of the given context", block)
}
