package blockprocessor

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/pkg/errors"
)

func (bp *blockProcessor) validateAndInsertPruningPoint(newPruningPoint *externalapi.DomainBlock, serializedUTXOSet []byte) error {
	log.Info("Checking that the given pruning point is the expected pruning point")
	selectedTip, err := bp.headersSelectedTipStore.HeadersSelectedTip(bp.databaseContext)
	if err != nil {
		return err
	}

	expectedNewPruningPointHash, err := bp.pruningManager.CalculateIndependentPruningPoint(selectedTip)
	if err != nil {
		return err
	}

	newPruningPointHash := consensushashing.BlockHash(newPruningPoint)

	if *expectedNewPruningPointHash != *newPruningPointHash {
		return errors.Wrapf(ruleerrors.ErrUnexpectedPruningPoint, "expected pruning point %s but got %s",
			expectedNewPruningPointHash, newPruningPointHash)
	}

	// We have to validate the pruning point block before we set the new pruning point in consensusStateManager.
	log.Infof("Validating the new pruning point %s", newPruningPointHash)
	err = bp.validateBlockAndDiscardChanges(newPruningPoint)
	if err != nil {
		return err
	}

	err = bp.consensusStateManager.UpdatePruningPoint(expectedNewPruningPointHash, serializedUTXOSet)
	if err != nil {
		return err
	}

	return bp.ValidateAndInsertBlock(newPruningPoint)
}
