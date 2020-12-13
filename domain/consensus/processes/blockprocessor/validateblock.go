package blockprocessor

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/pkg/errors"
)

func (bp *blockProcessor) validateBlockAndDiscardChanges(block *externalapi.DomainBlock) error {
	defer bp.discardAllChanges()
	return bp.validateBlock(block)
}

func (bp *blockProcessor) validateBlock(block *externalapi.DomainBlock) error {
	blockHash := consensushashing.HeaderHash(block.Header)
	log.Debugf("Validating block %s", blockHash)

	err := bp.checkBlockStatus(block)
	if err != nil {
		return err
	}

	hasValidatedHeader, err := bp.hasValidatedOnlyHeader(blockHash)
	if err != nil {
		return err
	}

	if !hasValidatedHeader {
		log.Tracef("Staging block %s header")
		bp.blockHeaderStore.Stage(blockHash, block.Header)
	} else {
		log.Tracef("Block %s header is already known, so no need to stage it")
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
	err = bp.validatePostProofOfWork(block)
	if err != nil {
		if errors.As(err, &ruleerrors.RuleError{}) {
			bp.discardAllChanges()
			hash := consensushashing.BlockHash(block)
			bp.blockStatusStore.Stage(hash, externalapi.StatusInvalid)
			commitErr := bp.commitAllChanges()
			if commitErr != nil {
				return commitErr
			}
		}
		return err
	}
	return nil
}
