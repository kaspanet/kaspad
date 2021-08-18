package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
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

	var expectedPruningPoint *externalapi.DomainHash
	currentPruningPoint, err := csm.pruningStore.PruningPoint(csm.databaseContext, stagingArea)
	if err != nil {
		return err
	}

	currentPruningPointGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, stagingArea, currentPruningPoint, false)
	if err != nil {
		return err
	}

	blockGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, stagingArea, blockHash, false)
	if err != nil {
		return err
	}

	// We can't validate the pruning point of blocks that their expected pruning point is not the current one.
	if currentPruningPoint.Equal(csm.genesisHash) || blockGHOSTDAGData.BlueScore()-currentPruningPointGHOSTDAGData.BlueScore() > csm.pruningDepth {
		expectedPruningPoint = currentPruningPoint
	} else {
		// TODO: Import the previous pruning points from the syncer
		expectedPruningPoint, err = csm.pruningStore.PreviousPruningPoint(csm.databaseContext, stagingArea)
		if database.IsNotFoundError(err) {
			expectedPruningPoint = csm.genesisHash
		} else if err != nil {
			return err
		}
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
