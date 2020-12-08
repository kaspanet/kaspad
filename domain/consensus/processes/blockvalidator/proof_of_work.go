package blockvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
)

func (v *blockValidator) ValidatePruningPointViolationAndProofOfWorkAndDifficulty(blockHash *externalapi.DomainHash) error {
	header, err := v.blockHeaderStore.BlockHeader(v.databaseContext, blockHash)
	if err != nil {
		return err
	}

	err = v.checkParentsExist(blockHash, header)
	if err != nil {
		return err
	}

	err = v.checkParentsIncest(header)
	if err != nil {
		return err
	}

	err = v.checkPruningPointViolation(header)
	if err != nil {
		return err
	}

	err = v.checkProofOfWork(header)
	if err != nil {
		return err
	}

	err = v.dagTopologyManager.SetParents(blockHash, header.ParentHashes)
	if err != nil {
		return err
	}

	err = v.validateDifficulty(blockHash)
	if err != nil {
		return err
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

	header, err := v.blockHeaderStore.BlockHeader(v.databaseContext, blockHash)
	if err != nil {
		return err
	}
	if header.Bits != expectedBits {
		return errors.Wrapf(ruleerrors.ErrUnexpectedDifficulty, "block difficulty of %d is not the expected value of %d", header.Bits, expectedBits)
	}

	return nil
}

// checkProofOfWork ensures the block header bits which indicate the target
// difficulty is in min/max range and that the block hash is less than the
// target difficulty as claimed.
//
// The flags modify the behavior of this function as follows:
//  - BFNoPoWCheck: The check to ensure the block hash is less than the target
//    difficulty is not performed.
func (v *blockValidator) checkProofOfWork(header *externalapi.DomainBlockHeader) error {
	// The target difficulty must be larger than zero.
	target := util.CompactToBig(header.Bits)
	if target.Sign() <= 0 {
		return errors.Wrapf(ruleerrors.ErrUnexpectedDifficulty, "block target difficulty of %064x is too low",
			target)
	}

	// The target difficulty must be less than the maximum allowed.
	if target.Cmp(v.powMax) > 0 {
		return errors.Wrapf(ruleerrors.ErrUnexpectedDifficulty, "block target difficulty of %064x is "+
			"higher than max of %064x", target, v.powMax)
	}

	// The block hash must be less than the claimed target unless the flag
	// to avoid proof of work checks is set.
	if !v.skipPoW {
		// The block hash must be less than the claimed target.
		hash := consensushashing.HeaderHash(header)
		hashNum := hashes.ToBig(hash)
		if hashNum.Cmp(target) > 0 {
			return errors.Wrapf(ruleerrors.ErrUnexpectedDifficulty, "block hash of %064x is higher than "+
				"expected max of %064x", hashNum, target)
		}
	}

	return nil
}

func (v *blockValidator) checkParentsExist(blockHash *externalapi.DomainHash, header *externalapi.DomainBlockHeader) error {
	missingParentHashes := []*externalapi.DomainHash{}

	isFullBlock, err := v.blockStore.HasBlock(v.databaseContext, blockHash)
	if err != nil {
		return err
	}

	for _, parent := range header.ParentHashes {
		parentHeaderExists, err := v.blockHeaderStore.HasBlockHeader(v.databaseContext, parent)
		if err != nil {
			return err
		}
		if !parentHeaderExists {
			missingParentHashes = append(missingParentHashes, parent)
			continue
		}

		if isFullBlock {
			parentStatus, err := v.blockStatusStore.Get(v.databaseContext, parent)
			if err != nil {
				return err
			}

			if parentStatus != externalapi.StatusHeaderOnly {
				continue
			}

			headerTipsPruningPoint, err := v.consensusStateManager.HeaderTipsPruningPoint()
			if err != nil {
				return err
			}

			isParentInPastOfHeaderTipsPruningPoint, err := v.dagTopologyManager.IsAncestorOf(headerTipsPruningPoint, parent)
			if err != nil {
				return err
			}

			if isParentInPastOfHeaderTipsPruningPoint {
				continue
			}

			missingParentHashes = append(missingParentHashes, parent)
		}
	}

	if len(missingParentHashes) > 0 {
		return ruleerrors.NewErrMissingParents(missingParentHashes)
	}

	return nil
}
func (v *blockValidator) checkPruningPointViolation(header *externalapi.DomainBlockHeader) error {
	// check if the pruning point is on past of at least one parent of the header's parents.

	hasPruningPoint, err := v.pruningStore.HasPruningPoint(v.databaseContext)
	if err != nil {
		return err
	}

	//If hasPruningPoint has a false value, it means that it's the genesis - so no violation can exist.
	if !hasPruningPoint {
		return nil
	}

	pruningPoint, err := v.pruningStore.PruningPoint(v.databaseContext)
	if err != nil {
		return err
	}

	isAncestorOfAny, err := v.dagTopologyManager.IsAncestorOfAny(pruningPoint, header.ParentHashes)
	if err != nil {
		return err
	}
	if isAncestorOfAny {
		return nil
	}
	return errors.Wrapf(ruleerrors.ErrPruningPointViolation,
		"expected pruning point to be in block %d past.", header.Bits)
}
