package blockvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/pow"
	"github.com/kaspanet/kaspad/domain/consensus/utils/virtual"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util/difficulty"
	"github.com/pkg/errors"
)

func (v *blockValidator) ValidatePruningPointViolationAndProofOfWorkAndDifficulty(stagingArea *model.StagingArea,
	blockHash *externalapi.DomainHash, isBlockWithTrustedData bool) error {

	onEnd := logger.LogAndMeasureExecutionTime(log, "ValidatePruningPointViolationAndProofOfWorkAndDifficulty")
	defer onEnd()

	header, err := v.blockHeaderStore.BlockHeader(v.databaseContext, stagingArea, blockHash)
	if err != nil {
		return err
	}

	err = v.checkParentNotVirtualGenesis(header)
	if err != nil {
		return err
	}

	err = v.checkParentHeadersExist(stagingArea, header, isBlockWithTrustedData)
	if err != nil {
		return err
	}

	err = v.setParents(stagingArea, blockHash, header, isBlockWithTrustedData)
	if err != nil {
		return err
	}

	err = v.checkParentsIncest(stagingArea, blockHash)
	if err != nil {
		return err
	}

	err = v.checkPruningPointViolation(stagingArea, blockHash)
	if err != nil {
		return err
	}

	err = v.checkProofOfWork(header)
	if err != nil {
		return err
	}

	err = v.validateDifficulty(stagingArea, blockHash, isBlockWithTrustedData)
	if err != nil {
		return err
	}

	return nil
}

func (v *blockValidator) setParents(stagingArea *model.StagingArea,
	blockHash *externalapi.DomainHash,
	header externalapi.BlockHeader,
	isBlockWithTrustedData bool) error {

	for level := 0; level <= pow.BlockLevel(header); level++ {
		var parents []*externalapi.DomainHash
		for _, parent := range header.ParentsAtLevel(level) {
			exists, err := v.blockStatusStore.Exists(v.databaseContext, stagingArea, parent)
			if err != nil {
				return err
			}

			if !exists {
				if level == 0 && !isBlockWithTrustedData {
					return errors.Errorf("direct parent %s is missing: only block with prefilled information can have some missing parents", parent)
				}
				continue
			}

			parents = append(parents, parent)
		}

		if len(parents) == 0 {
			parents = append(parents, model.VirtualGenesisBlockHash)
		}

		err := v.dagTopologyManagers[level].SetParents(stagingArea, blockHash, parents)
		if err != nil {
			return err
		}
	}

	return nil
}

func (v *blockValidator) validateDifficulty(stagingArea *model.StagingArea,
	blockHash *externalapi.DomainHash,
	isBlockWithTrustedData bool) error {

	if !isBlockWithTrustedData {
		// We need to calculate GHOSTDAG for the block in order to check its difficulty and blue work
		err := v.ghostdagManagers[0].GHOSTDAG(stagingArea, blockHash)
		if err != nil {
			return err
		}
	}

	header, err := v.blockHeaderStore.BlockHeader(v.databaseContext, stagingArea, blockHash)
	if err != nil {
		return err
	}

	blockLevel := pow.BlockLevel(header)
	for i := 1; i <= blockLevel; i++ {
		err = v.ghostdagManagers[i].GHOSTDAG(stagingArea, blockHash)
		if err != nil {
			return err
		}
	}

	// Ensure the difficulty specified in the block header matches
	// the calculated difficulty based on the previous block and
	// difficulty retarget rules.
	expectedBits, err := v.difficultyManager.StageDAADataAndReturnRequiredDifficulty(stagingArea, blockHash, isBlockWithTrustedData)
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

func (v *blockValidator) checkParentNotVirtualGenesis(header externalapi.BlockHeader) error {
	for _, parent := range header.DirectParents() {
		if parent.Equal(model.VirtualGenesisBlockHash) {
			return errors.Wrapf(ruleerrors.ErrVirtualGenesisParent, "block header cannot have the virtual genesis as parent")
		}
	}

	return nil
}

func (v *blockValidator) checkParentHeadersExist(stagingArea *model.StagingArea,
	header externalapi.BlockHeader,
	isBlockWithTrustedData bool) error {

	if isBlockWithTrustedData {
		return nil
	}

	missingParentHashes := []*externalapi.DomainHash{}
	for _, parent := range header.DirectParents() {
		parentHeaderExists, err := v.blockHeaderStore.HasBlockHeader(v.databaseContext, stagingArea, parent)
		if err != nil {
			return err
		}
		if !parentHeaderExists {
			parentStatus, err := v.blockStatusStore.Get(v.databaseContext, stagingArea, parent)
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
func (v *blockValidator) checkPruningPointViolation(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) error {
	// check if the pruning point is on past of at least one parent of the header's parents.

	hasPruningPoint, err := v.pruningStore.HasPruningPoint(v.databaseContext, stagingArea)
	if err != nil {
		return err
	}

	//If hasPruningPoint has a false value, it means that it's the genesis - so no violation can exist.
	if !hasPruningPoint {
		return nil
	}

	pruningPoint, err := v.pruningStore.PruningPoint(v.databaseContext, stagingArea)
	if err != nil {
		return err
	}

	parents, err := v.dagTopologyManagers[0].Parents(stagingArea, blockHash)
	if err != nil {
		return err
	}

	if virtual.ContainsOnlyVirtualGenesis(parents) {
		return nil
	}

	isAncestorOfAny, err := v.dagTopologyManagers[0].IsAncestorOfAny(stagingArea, pruningPoint, parents)
	if err != nil {
		return err
	}

	if !isAncestorOfAny {
		return errors.Wrapf(ruleerrors.ErrPruningPointViolation,
			"expected pruning point %s to be in block %s past.", pruningPoint, blockHash)
	}
	return nil
}
