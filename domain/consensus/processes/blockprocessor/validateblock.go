package blockprocessor

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/pkg/errors"
)

func (bp *blockProcessor) validateBlockAndDiscardChanges(block *externalapi.DomainBlock, isPruningPoint bool) error {
	return bp.validateBlock(model.NewStagingArea(), block, isPruningPoint)
}

func (bp *blockProcessor) validateBlock(stagingArea *model.StagingArea, block *externalapi.DomainBlock, isPruningPoint bool) error {
	blockHash := consensushashing.HeaderHash(block.Header)
	log.Debugf("Validating block %s", blockHash)

	err := bp.checkBlockStatus(stagingArea, block)
	if err != nil {
		return err
	}

	hasValidatedHeader, err := bp.hasValidatedHeader(stagingArea, blockHash)
	if err != nil {
		return err
	}

	if !hasValidatedHeader {
		log.Debugf("Staging block %s header", blockHash)
		bp.blockHeaderStore.Stage(stagingArea, blockHash, block.Header)
	} else {
		log.Debugf("Block %s header is already known, so no need to stage it", blockHash)
	}

	// If any validation until (included) proof-of-work fails, simply
	// return an error without writing anything in the database.
	// This is to prevent spamming attacks.
	err = bp.validatePreProofOfWork(stagingArea, block)
	if err != nil {
		return err
	}

	if !hasValidatedHeader {
		err = bp.blockValidator.ValidatePruningPointViolationAndProofOfWorkAndDifficulty(stagingArea, blockHash)
		if err != nil {
			return err
		}
	}

	// If in-context validations fail, discard all changes and store the
	// block with StatusInvalid.
	err = bp.validatePostProofOfWork(stagingArea, block, isPruningPoint)
	if err != nil {
		if errors.As(err, &ruleerrors.RuleError{}) {
			// We mark invalid blocks with status externalapi.StatusInvalid except in the
			// case of the following errors:
			// ErrMissingParents - If we got ErrMissingParents the block shouldn't be
			// considered as invalid because it could be added later on when its
			// parents are present.
			// ErrBadMerkleRoot - if we get ErrBadMerkleRoot we shouldn't mark the
			// block as invalid because later on we can get the block with
			// transactions that fits the merkle root.
			// ErrPrunedBlock - ErrPrunedBlock is an error that rejects a block body and
			// not the block as a whole, so we shouldn't mark it as invalid.
			if !errors.As(err, &ruleerrors.ErrMissingParents{}) &&
				!errors.Is(err, ruleerrors.ErrBadMerkleRoot) &&
				!errors.Is(err, ruleerrors.ErrPrunedBlock) {
				// Use a new stagingArea so we save only the block status
				stagingArea := model.NewStagingArea()
				hash := consensushashing.BlockHash(block)
				bp.blockStatusStore.Stage(stagingArea, hash, externalapi.StatusInvalid)
				commitErr := bp.commitAllChanges(stagingArea)
				if commitErr != nil {
					return commitErr
				}
			}
		}
		return err
	}
	return nil
}
