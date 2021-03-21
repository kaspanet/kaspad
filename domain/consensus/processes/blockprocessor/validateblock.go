package blockprocessor

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/pkg/errors"
)

func (bp *blockProcessor) validateBlockAndDiscardChanges(block *externalapi.DomainBlock, isPruningPoint bool) error {
	defer bp.discardAllChanges()
	return bp.validateBlock(block, isPruningPoint)
}

func (bp *blockProcessor) validateBlock(block *externalapi.DomainBlock, isPruningPoint bool) error {
	blockHash := consensushashing.HeaderHash(block.Header)
	log.Debugf("Validating block %s", blockHash)

	err := bp.checkBlockStatus(block)
	if err != nil {
		return err
	}

	hasValidatedHeader, err := bp.hasValidatedHeader(blockHash)
	if err != nil {
		return err
	}

	if !hasValidatedHeader {
		log.Debugf("Staging block %s header", blockHash)
		bp.blockHeaderStore.Stage(nil, blockHash, block.Header)
	} else {
		log.Debugf("Block %s header is already known, so no need to stage it", blockHash)
	}

	// If any validation until (included) proof-of-work fails, simply
	// return an error without writing anything in the database.
	// This is to prevent spamming attacks.
	err = bp.validatePreProofOfWork(block)
	if err != nil {
		return err
	}

	if !hasValidatedHeader {
		err = bp.blockValidator.ValidatePruningPointViolationAndProofOfWorkAndDifficulty(blockHash)
		if err != nil {
			return err
		}
	}

	// If in-context validations fail, discard all changes and store the
	// block with StatusInvalid.
	err = bp.validatePostProofOfWork(block, isPruningPoint)
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
				// Discard all changes so we save only the block status
				bp.discardAllChanges()
				hash := consensushashing.BlockHash(block)
				bp.blockStatusStore.Stage(nil, hash, externalapi.StatusInvalid)
				commitErr := bp.commitAllChanges()
				if commitErr != nil {
					return commitErr
				}
			}
		}
		return err
	}
	return nil
}
