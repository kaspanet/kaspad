package blockvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashserialization"
	"github.com/pkg/errors"
)

// ValidateHeaderInContext validates block headers in the context of the current
// consensus state
func (v *blockValidator) ValidateHeaderInContext(blockHash *externalapi.DomainHash) error {
	header, err := v.blockHeaderStore.BlockHeader(v.databaseContext, blockHash)
	if err != nil {
		return err
	}

	err = v.checkParentsIncest(header)
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

	err = v.checkBoundedMergeDepth(blockHash)
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

	if mergeSetSize > constants.MergeSetSizeLimit {
		return errors.Wrapf(ruleerrors.ErrViolatingMergeLimit,
			"The block merges %d blocks > %d merge set size limit", mergeSetSize, constants.MergeSetSizeLimit)
	}

	return nil
}
