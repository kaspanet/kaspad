package blockvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/pkg/errors"
)

func (v *blockValidator) validateHeaderPruningPoint(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) error {
	if blockHash.Equal(v.genesisHash) {
		return nil
	}

	header, err := v.blockHeaderStore.BlockHeader(v.databaseContext, stagingArea, blockHash)
	if err != nil {
		return err
	}

	expectedPruningPoint, err := v.pruningManager.ExpectedHeaderPruningPoint(stagingArea, blockHash)
	if err != nil {
		return err
	}

	if !header.PruningPoint().Equal(expectedPruningPoint) {
		return errors.Wrapf(ruleerrors.ErrUnexpectedPruningPoint, "block pruning point of %s is not the expected hash of %s", header.PruningPoint(), expectedPruningPoint)
	}

	return nil
}
