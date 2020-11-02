package blockprocessor

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/pkg/errors"
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

	// If in-context validations fail, discard all changes and store the
	// block with StatusInvalid.
	err = bp.validateInContext(block)
	if err != nil {
		if errors.As(err, &ruleerrors.RuleError{}) {
			bp.discardAllChanges()
			bp.blockStatusStore.Stage(block.Hash, externalapi.StatusInvalid)
			commitErr := bp.commitAllChanges()
			if commitErr != nil {
				return commitErr
			}
		}
		return err
	}
	return nil
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

func (bp *blockProcessor) validateInContext(block *externalapi.DomainBlock) error {
	err := bp.dagTopologyManager.SetParents(block.Hash, block.Header.ParentHashes)
	if err != nil {
		return err
	}

	bp.blockStore.Stage(block.Hash, block)
	bp.blockHeaderStore.Stage(block.Hash, block.Header)

	err = bp.blockValidator.ValidateHeaderInContext(block.Hash)
	if err != nil {
		return err
	}
	err = bp.blockValidator.ValidateBodyInContext(block.Hash)
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
	dbTx, err := bp.databaseContext.Begin()
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
