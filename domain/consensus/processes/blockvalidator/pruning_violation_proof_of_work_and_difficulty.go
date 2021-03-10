package blockvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/pow"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util/difficulty"
	"github.com/pkg/errors"
)

func (v *blockValidator) ValidatePruningPointViolationAndProofOfWorkAndDifficulty(blockHash *externalapi.DomainHash) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "ValidatePruningPointViolationAndProofOfWorkAndDifficulty")
	defer onEnd()

	header, err := v.blockHeaderStore.BlockHeader(v.databaseContext, blockHash)
	if err != nil {
		return err
	}

	err = v.checkParentHeadersExist(header)
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

	err = v.dagTopologyManager.SetParents(blockHash, header.ParentHashes())
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
	// We need to calculate GHOSTDAG for the block in order to check its difficulty
	err := v.ghostdagManager.GHOSTDAG(blockHash)
	if err != nil {
		return err
	}

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
	if header.Bits() != expectedBits {
		return errors.Wrapf(ruleerrors.ErrUnexpectedDifficulty, "block difficulty of %d is not the expected value of %d", header.Bits(), expectedBits)
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
func (v *blockValidator) checkProofOfWork(header externalapi.BlockHeader) error {
	// The target difficulty must be larger than zero.
	target := difficulty.CompactToBig(header.Bits())
	if target.Sign() <= 0 {
		return errors.Wrapf(ruleerrors.ErrNegativeTarget, "block target difficulty of %064x is too low",
			target)
	}

	// The target difficulty must be less than the maximum allowed.
	if target.Cmp(v.powMax) > 0 {
		return errors.Wrapf(ruleerrors.ErrTargetTooHigh, "block target difficulty of %064x is "+
			"higher than max of %064x", target, v.powMax)
	}

	// The block pow must be valid unless the flag to avoid proof of work checks is set.
	if !v.skipPoW {
		valid := pow.CheckProofOfWorkWithTarget(header.ToMutable(), target)
		if !valid {
			return errors.Wrap(ruleerrors.ErrInvalidPoW, "block has invalid proof of work")
		}
	}
	return nil
}

func (v *blockValidator) checkParentHeadersExist(header externalapi.BlockHeader) error {
	missingParentHashes := []*externalapi.DomainHash{}
	for _, parent := range header.ParentHashes() {
		parentHeaderExists, err := v.blockHeaderStore.HasBlockHeader(v.databaseContext, parent)
		if err != nil {
			return err
		}
		if !parentHeaderExists {
			parentStatus, err := v.blockStatusStore.Get(v.databaseContext, parent)
			if err != nil {
				if !database.IsNotFoundError(err) {
					return err
				}
			} else if parentStatus == externalapi.StatusInvalid {
				return errors.Wrapf(ruleerrors.ErrInvalidAncestorBlock, "parent %s is invalid", parent)
			}

			missingParentHashes = append(missingParentHashes, parent)
			continue
		}
	}

	if len(missingParentHashes) > 0 {
		return ruleerrors.NewErrMissingParents(missingParentHashes)
	}

	return nil
}
func (v *blockValidator) checkPruningPointViolation(header externalapi.BlockHeader) error {
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

	isAncestorOfAny, err := v.dagTopologyManager.IsAncestorOfAny(pruningPoint, header.ParentHashes())
	if err != nil {
		return err
	}
	if !isAncestorOfAny {
		return errors.Wrapf(ruleerrors.ErrPruningPointViolation,
			"expected pruning point %s to be in block %s past.", pruningPoint, consensushashing.HeaderHash(header))
	}
	return nil
}
