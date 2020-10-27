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
	// If either in-isolation or proof-of-work validations fail, simply
	// return an error without writing anything in the database.
	// This is to prevent spamming attacks.
	err := bp.validateBlockInIsolationAndProofOfWork(block)
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

	return bp.validateInContext(block.Hash)
}

func (bp *blockProcessor) validateBlockInIsolationAndProofOfWork(block *externalapi.DomainBlock) error {
	err := bp.blockValidator.ValidateHeaderInIsolation(block.Hash)
	if err != nil {
		return err
	}
	err = bp.blockValidator.ValidateBodyInIsolation(block.Hash)
	if err != nil {
		return err
	}
	err = bp.blockValidator.ValidateProofOfWorkAndDifficulty(block.Hash)
	if err != nil {
		return err
	}
	return nil
}

func (bp *blockProcessor) validateInContext(blockHash *externalapi.DomainHash) error {
	err := bp.blockValidator.ValidateHeaderInContext(blockHash)
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
	for _, store := range bp.stores {
		store.Discard()
	}
}

func (bp *blockProcessor) commitAllChanges() error {
	dbTx, err := bp.databaseContext.NewTx()
	if err != nil {
		return err
	}

	for _, store := range bp.stores {
		err = store.Commit(dbTx)
		if err != nil {
			return err
		}
	}

	return dbTx.Commit()
}
