package blockprocessor

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/pkg/errors"
)

func (bp *blockProcessor) validateAndInsertImportedPruningPoint(
	stagingArea *model.StagingArea, newPruningPointHash *externalapi.DomainHash) error {

	log.Info("Checking that the given pruning point is the expected pruning point")

	isValidPruningPoint, err := bp.pruningManager.IsValidPruningPoint(stagingArea, newPruningPointHash)
	if err != nil {
		return err
	}

	if !isValidPruningPoint {
		return errors.Wrapf(ruleerrors.ErrUnexpectedPruningPoint, "%s is not a valid pruning point",
			newPruningPointHash)
	}

	arePruningPointsInValidChain, err := bp.pruningManager.ArePruningPointsInValidChain(stagingArea)
	if err != nil {
		return err
	}

	if !arePruningPointsInValidChain {
		return errors.Wrapf(ruleerrors.ErrInvalidPruningPointsChain, "pruning points do not compose a valid "+
			"chain to genesis")
	}

	log.Infof("Updating consensus state manager according to the new pruning point %s", newPruningPointHash)
	err = bp.consensusStateManager.ImportPruningPointUTXOSet(stagingArea, newPruningPointHash)
	if err != nil {
		return err
	}

	err = bp.updateVirtualAcceptanceDataAfterImportingPruningPoint(stagingArea)
	if err != nil {
		return err
	}

	return nil
}
