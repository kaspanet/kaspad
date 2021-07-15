package blockprocessor

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/util/staging"
	"github.com/pkg/errors"
)

func (bp *blockProcessor) validateBlock(stagingArea *model.StagingArea, block *externalapi.DomainBlock, isBlockWithTrustedData bool) error {
	blockHash := consensushashing.HeaderHash(block.Header)
	log.Debugf("Validating block %s", blockHash)

	// Since genesis has a lot of special cases validation rules, we make sure it's not added unintentionally
	// on uninitialized node.
	if blockHash.Equal(bp.genesisHash) && bp.blockStore.Count(stagingArea) != 0 {
		return errors.Wrapf(ruleerrors.ErrGenesisOnInitializedConsensus, "Cannot add genesis to an initialized consensus")
	}

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
		err = bp.blockValidator.ValidatePruningPointViolationAndProofOfWorkAndDifficulty(stagingArea, blockHash, isBlockWithTrustedData)
		if err != nil {
			return err
		}
	}

	// If in-context validations fail, discard all changes and store the
	// block with StatusInvalid.
	err = bp.validatePostProofOfWork(stagingArea, block, isBlockWithTrustedData)
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
				commitErr := staging.CommitAllChanges(bp.databaseContext, stagingArea)
				if commitErr != nil {
					return commitErr
				}
			}
		}
		return err
	}
	return nil
}
