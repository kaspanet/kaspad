package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/pkg/errors"
)

func (csm *consensusStateManager) verifyChainBlock(stagingArea *model.StagingArea,
	blockHash *externalapi.DomainHash, pastUTXOSet externalapi.UTXODiff, acceptanceData externalapi.AcceptanceData,
	multiset model.Multiset) error {

	err := csm.verifyHeaderPruningPoint(stagingArea, blockHash)
	if err != nil {
		return err
	}

	log.Tracef("verifying the UTXO of block %s", blockHash)
	err = csm.verifyUTXO(stagingArea, blockHash, pastUTXOSet, acceptanceData, multiset)
	if err != nil {
		return err
	}

	return nil
}

func (csm *consensusStateManager) verifyHeaderPruningPoint(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) error {
	if blockHash.Equal(csm.genesisHash) {
		return nil
	}

	expectedPruningPoint, err := csm.expectedHeaderPruningPoint(stagingArea, blockHash)
	if err != nil {
		return err
	}

	header, err := csm.blockHeaderStore.BlockHeader(csm.databaseContext, stagingArea, blockHash)
	if err != nil {
		return err
	}

	if !header.PruningPoint().Equal(expectedPruningPoint) {
		return errors.Wrapf(ruleerrors.ErrUnexpectedPruningPoint, "block pruning point of %s is not the expected hash of %s", header.PruningPoint(), expectedPruningPoint)
	}

	return nil
}

func (csm *consensusStateManager) expectedHeaderPruningPoint(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	pruningPointIndex, err := csm.pruningStore.PruningPointIndex(csm.databaseContext, stagingArea)
	if err != nil {
		return nil, err
	}

	for i := pruningPointIndex; ; i-- {
		currentPruningPoint, err := csm.pruningStore.PruningPointByIndex(csm.databaseContext, stagingArea, i)
		if err != nil {
			return nil, err
		}

		currentPruningPointGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, stagingArea, currentPruningPoint, false)
		if err != nil {
			return nil, err
		}

		blockGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, stagingArea, blockHash, false)
		if err != nil {
			return nil, err
		}

		if blockGHOSTDAGData.BlueScore()-currentPruningPointGHOSTDAGData.BlueScore() > csm.pruningDepth {
			return currentPruningPoint, nil
		}

		if i == 0 {
			break
		}
	}

	return csm.genesisHash, nil
}
