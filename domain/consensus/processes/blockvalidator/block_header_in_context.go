package blockvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashserialization"
	"github.com/pkg/errors"
)

// ValidateHeaderInContext validates block headers in the context of the current
// consensus state
func (v *blockValidator) ValidateHeaderInContext(blockHash *externalapi.DomainHash) error {
	block, err := v.blockStore.Block(v.databaseContext, blockHash)
	if err != nil {
		return err
	}

	header := block.Header

	err = v.checkParentsIncest(header)
	if err != nil {
		return err
	}

	err = v.validateDifficulty(blockHash)
	if err != nil {
		return err
	}

	err = v.validateMedianTime(header)
	if err != nil {
		return err
	}

	err = v.ghostdagManager.GHOSTDAG(blockHash)
	if err != nil {
		return err
	}

	err = v.checkMergeSizeLimit(blockHash)
	if err != nil {
		return err
	}

	return nil
}

// checkParentsIncest validates that no parent is an ancestor of another parent
func (v *blockValidator) checkParentsIncest(header *externalapi.DomainBlockHeader) error {
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
				return errors.Wrapf(ruleerrors.ErrInvalidParentsRelation, "parent %s is an "+
					"ancestor of another parent %s",
					parentA,
					parentB,
				)
			}
		}
	}
	return nil
}

func (v *blockValidator) validateDifficulty(blockHash *externalapi.DomainHash) error {
	// Ensure the difficulty specified in the block header matches
	// the calculated difficulty based on the previous block and
	// difficulty retarget rules.
	expectedBits, err := v.difficultyManager.RequiredDifficulty(blockHash)
	if err != nil {
		return err
	}

	block, err := v.blockStore.Block(v.databaseContext, blockHash)
	if err != nil {
		return err
	}

	header := block.Header
	if header.Bits != expectedBits {
		return errors.Wrapf(ruleerrors.ErrUnexpectedDifficulty, "block difficulty of %d is not the expected value of %d", header.Bits, expectedBits)
	}

	return nil
}

func (v *blockValidator) validateMedianTime(header *externalapi.DomainBlockHeader) error {
	if len(header.ParentHashes) == 0 {
		return nil
	}

	hash := hashserialization.HeaderHash(header)
	ghostdagData, err := v.ghostdagDataStore.Get(v.databaseContext, hash)
	if err != nil {
		return err
	}

	// Ensure the timestamp for the block header is not before the
	// median time of the last several blocks (medianTimeBlocks).
	pastMedianTime, err := v.pastMedianTimeManager.PastMedianTime(ghostdagData.SelectedParent)
	if err != nil {
		return err
	}

	if header.TimeInMilliseconds < pastMedianTime {
		return errors.Wrapf(ruleerrors.ErrTimeTooOld, "block timestamp of %d is not after expected %d",
			header.TimeInMilliseconds, pastMedianTime)
	}

	return nil
}

func (v *blockValidator) checkMergeSizeLimit(hash *externalapi.DomainHash) error {
	ghostdagData, err := v.ghostdagDataStore.Get(v.databaseContext, hash)
	if err != nil {
		return err
	}

	mergeSetSize := len(ghostdagData.MergeSetReds) + len(ghostdagData.MergeSetBlues)

	const mergeSetSizeLimit = 1000
	if mergeSetSize > mergeSetSizeLimit {
		return errors.Wrapf(ruleerrors.ErrViolatingMergeLimit,
			"The block merges %d blocks > %d merge set size limit", mergeSetSize, mergeSetSizeLimit)
	}

	return nil
}
