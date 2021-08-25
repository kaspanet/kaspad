package blockvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/pkg/errors"
)

func (v *blockValidator) ValidateHeaderPruningPoint(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) error {
	if blockHash.Equal(v.genesisHash) {
		return nil
	}

	header, err := v.blockHeaderStore.BlockHeader(v.databaseContext, stagingArea, blockHash)
	if err != nil {
		return err
	}

	expectedPruningPoint, err := v.expectedHeaderPruningPoint(stagingArea, blockHash)
	if err != nil {
		return err
	}

	if !header.PruningPoint().Equal(expectedPruningPoint) {
		return errors.Wrapf(ruleerrors.ErrUnexpectedPruningPoint, "block pruning point of %s is not the expected hash of %s", header.PruningPoint(), expectedPruningPoint)
	}

	return nil
}

func (v *blockValidator) expectedHeaderPruningPoint(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	pruningPointIndex, err := v.pruningStore.PruningPointIndex(v.databaseContext, stagingArea)
	if err != nil {
		return nil, err
	}

	blockGHOSTDAGData, err := v.ghostdagDataStore.Get(v.databaseContext, stagingArea, blockHash, false)
	if err != nil {
		return nil, err
	}

	for i := pruningPointIndex; ; i-- {
		currentPruningPoint, err := v.pruningStore.PruningPointByIndex(v.databaseContext, stagingArea, i)
		if err != nil {
			return nil, err
		}

		currentPruningPointHeader, err := v.blockHeaderStore.BlockHeader(v.databaseContext, stagingArea, currentPruningPoint)
		if err != nil {
			return nil, err
		}

		if blockGHOSTDAGData.BlueScore() >= currentPruningPointHeader.BlueScore()+v.pruningDepth {
			return currentPruningPoint, nil
		}

		if i == 0 {
			break
		}
	}

	return v.genesisHash, nil
}
