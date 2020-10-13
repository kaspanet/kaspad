package blockvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
)

// ValidateHeaderInContext validates block headers in the context of the current
// consensus state
func (bv *BlockValidator) ValidateHeaderInContext(header *model.DomainBlockHeader) error {
	err := bv.checkParentsIncest(header)
	if err != nil {
		return err
	}

	err = bv.validateDifficulty(header)
	if err != nil {
		return err
	}

	err = bv.validateMedianTime(header)
	if err != nil {
		return err
	}

	ghostdagData := bv.ghostdagManager.GHOSTDAG(header.ParentHashes)
	err = bv.checkMergeSizeLimit(ghostdagData)
	if err != nil {
		return err
	}

	return nil
}

// checkParentsIncest validates that no parent is an ancestor of another parent
func (bv *BlockValidator) checkParentsIncest(header *model.DomainBlockHeader) error {
	for _, parentA := range header.ParentHashes {
		for _, parentB := range header.ParentHashes {
			if *parentA == *parentB {
				continue
			}

			if bv.dagTopologyManager.IsAncestorOf(parentA, parentB) {
				return ruleerrors.Errorf(ruleerrors.ErrInvalidParentsRelation, "parent %s is an "+
					"ancestor of another parent %s",
					parentA,
					parentB,
				)
			}
		}
	}
	return nil
}

func (bv *BlockValidator) validateDifficulty(header *model.DomainBlockHeader) error {
	// Ensure the difficulty specified in the block header matches
	// the calculated difficulty based on the previous block and
	// difficulty retarget rules.
	panic("unimplemented")
}

func (bv *BlockValidator) validateMedianTime(header *model.DomainBlockHeader) error {
	panic("unimplemented")
}

func (bv *BlockValidator) checkMergeSizeLimit(ghostdagData *model.BlockGHOSTDAGData) error {
	mergeSetSize := len(ghostdagData.MergeSetReds) + len(ghostdagData.MergeSetBlues)

	const mergeSetSizeLimit = 1000
	if mergeSetSize > mergeSetSizeLimit {
		return ruleerrors.Errorf(ruleerrors.ErrViolatingMergeLimit,
			"The block merges %d blocks > %d merge set size limit", mergeSetSize, mergeSetSizeLimit)
	}

	return nil
}
