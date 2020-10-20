package validator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
)

// ValidateHeaderInContext validates block headers in the context of the current
// consensus state
func (v *validator) ValidateHeaderInContext(header *model.DomainBlockHeader) error {
	err := v.checkParentsIncest(header)
	if err != nil {
		return err
	}

	err = v.validateDifficulty(header)
	if err != nil {
		return err
	}

	err = v.validateMedianTime(header)
	if err != nil {
		return err
	}

	ghostdagData, err := v.ghostdagManager.GHOSTDAG(header.ParentHashes)
	if err != nil {
		return err
	}

	err = v.checkMergeSizeLimit(ghostdagData)
	if err != nil {
		return err
	}

	return nil
}

// checkParentsIncest validates that no parent is an ancestor of another parent
func (v *validator) checkParentsIncest(header *model.DomainBlockHeader) error {
	for _, parentA := range header.ParentHashes {
		for _, parentB := range header.ParentHashes {
			if *parentA == *parentB {
				continue
			}

			isAAncestorOfB, err := v.dagTopologyManager.IsAncestorOf(parentA, parentB)
			if err != nil {
				return err
			}

			if isAAncestorOfB {
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

func (v *validator) validateDifficulty(header *model.DomainBlockHeader) error {
	// Ensure the difficulty specified in the block header matches
	// the calculated difficulty based on the previous block and
	// difficulty retarget rules.
	expectedBits, err := v.difficultyManager.RequiredDifficulty(header.ParentHashes)
	if err != nil {
		return err
	}

	if header.Bits != expectedBits {
		return ruleerrors.Errorf(ruleerrors.ErrUnexpectedDifficulty, "block difficulty of %d is not the expected value of %d", header.Bits, expectedBits)
	}

	return nil
}

func (v *validator) validateMedianTime(header *model.DomainBlockHeader) error {
	panic("unimplemented")
}

func (v *validator) checkMergeSizeLimit(ghostdagData *model.BlockGHOSTDAGData) error {
	mergeSetSize := len(ghostdagData.MergeSetReds) + len(ghostdagData.MergeSetBlues)

	const mergeSetSizeLimit = 1000
	if mergeSetSize > mergeSetSizeLimit {
		return ruleerrors.Errorf(ruleerrors.ErrViolatingMergeLimit,
			"The block merges %d blocks > %d merge set size limit", mergeSetSize, mergeSetSizeLimit)
	}

	return nil
}
