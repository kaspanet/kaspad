package blockprocessor

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func (bp *blockProcessor) validateAndInsertBlock(block *externalapi.DomainBlock) error {
	err := bp.validateBlock(block)
	if err != nil {
		bp.discardAllChanges()
		return err
	}

	// Block validations passed, save whatever DAG data was
	// collected so far
	err = bp.commitAllChanges()
	if err != nil {
		return err
	}

	// Attempt to add the block to the virtual
	err = bp.consensusStateManager.AddBlockToVirtual(block.Hash)
	if err != nil {
		return err
	}

	return bp.commitAllChanges()
}

func (bp *blockProcessor) validateBlock(block *externalapi.DomainBlock) error {
	// If either header of proof-of-work validation fails, simply
	// return an error without writing anything in the database.
	// This is to prevent spamming attacks.
	err := bp.validateHeaderAndProofOfWork(block)
	if err != nil {
		return err
	}

	// It's now safe to store the block in the database
	bp.blockStatusStore.Stage(block.Hash, model.StatusInvalid)
	bp.blockStore.Stage(block.Hash, block)
	err = bp.commitAllChanges()
	if err != nil {
		return err
	}

	return bp.validateBody(block.Hash)
}

func (bp *blockProcessor) validateHeaderAndProofOfWork(block *externalapi.DomainBlock) error {
	err := bp.blockValidator.ValidateHeaderInIsolation(block.Hash)
	if err != nil {
		return err
	}
	err = bp.blockValidator.ValidateHeaderInContext(block.Hash)
	if err != nil {
		return err
	}
	err = bp.blockValidator.ValidateProofOfWorkAndDifficulty(block.Hash)
	if err != nil {
		return err
	}
	return nil
}

func (bp *blockProcessor) validateBody(blockHash *externalapi.DomainHash) error {
	err := bp.blockValidator.ValidateBodyInIsolation(blockHash)
	if err != nil {
		return err
	}
	err = bp.blockValidator.ValidateBodyInContext(blockHash)
	if err != nil {
		return err
	}
	return nil
}

func (bp *blockProcessor) discardAllChanges() {

}

func (bp *blockProcessor) commitAllChanges() error {
	return nil
}
