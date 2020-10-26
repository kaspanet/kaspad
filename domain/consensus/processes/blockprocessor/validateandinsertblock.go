package blockprocessor

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func (bp *blockProcessor) validateAndInsertBlock(block *externalapi.DomainBlock) error {
	err := bp.blockValidator.ValidateProofOfWork(block)
	if err != nil {
		// If the validation failed:
		//   Write in blockStatusStore that the block is invalid
		// return err
	}

	err = bp.validateBlockInIsolationAndInContext(block)
	if err != nil {
		return err
	}

	return nil
}

func (bp *blockProcessor) validateBlockInIsolationAndInContext(block *externalapi.DomainBlock) error {
	err := bp.blockValidator.ValidateHeaderInIsolation(block.Hash)
	if err != nil {
		return err
	}

	err = bp.blockValidator.ValidateHeaderInContext(block.Hash)
	if err != nil {
		return err
	}

	err = bp.blockValidator.ValidateBodyInIsolation(block.Hash)
	if err != nil {
		return err
	}

	err = bp.blockValidator.ValidateBodyInContext(block.Hash)
	if err != nil {
		return err
	}

	return nil
}

func (bp *blockProcessor) insertBlock(block *externalapi.DomainBlock) error {
	return nil
}

func (bp *blockProcessor) processNonValidBlock(block *externalapi.DomainBlock, blockStatus model.BlockStatus) error {
	return nil
}
