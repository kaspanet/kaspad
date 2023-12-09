package blockvalidator

import (
	"github.com/pkg/errors"
	"github.com/zoomy-network/zoomyd/domain/consensus/model"
	"github.com/zoomy-network/zoomyd/domain/consensus/model/externalapi"
	"github.com/zoomy-network/zoomyd/domain/consensus/ruleerrors"
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
