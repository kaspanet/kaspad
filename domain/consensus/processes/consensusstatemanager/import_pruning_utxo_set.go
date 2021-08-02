package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util/staging"
	"github.com/pkg/errors"
)

func (csm *consensusStateManager) ImportPruningPoint(stagingArea *model.StagingArea, newPruningPoint *externalapi.DomainHash) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "ImportPruningPoint")
	defer onEnd()

	err := csm.importPruningPoint(stagingArea, newPruningPoint)
	if err != nil {
		return err
	}

	err = csm.applyImportedPruningPointUTXOSet(stagingArea)
	if err != nil {
		return err
	}

	return nil
}

func (csm *consensusStateManager) importPruningPoint(
	stagingArea *model.StagingArea, newPruningPoint *externalapi.DomainHash) error {

	log.Debugf("importPruningPoint start")
	defer log.Debugf("importPruningPoint end")

	// TODO: We should validate the imported pruning point doesn't violate finality as part of the headers proof.

	importedPruningPointMultiset, err := csm.pruningStore.ImportedPruningPointMultiset(csm.databaseContext)
	if err != nil {
		return err
	}

	newPruningPointHeader, err := csm.blockHeaderStore.BlockHeader(csm.databaseContext, stagingArea, newPruningPoint)
	if err != nil {
		return err
	}
	log.Debugf("The UTXO commitment of the pruning point: %s",
		newPruningPointHeader.UTXOCommitment())

	if !newPruningPointHeader.UTXOCommitment().Equal(importedPruningPointMultiset.Hash()) {
		return errors.Wrapf(ruleerrors.ErrBadPruningPointUTXOSet, "the expected multiset hash of the pruning "+
			"point UTXO set is %s but got %s", newPruningPointHeader.UTXOCommitment(), *importedPruningPointMultiset.Hash())
	}
	log.Debugf("The new pruning point UTXO commitment validation passed")

	log.Debugf("Setting the pruning point as the only virtual parent")
	err = csm.dagTopologyManager.SetParents(stagingArea, model.VirtualBlockHash, []*externalapi.DomainHash{newPruningPoint})
	if err != nil {
		return err
	}

	log.Debugf("Calculating GHOSTDAG for the new virtual")
	err = csm.ghostdagManager.GHOSTDAG(stagingArea, model.VirtualBlockHash)
	if err != nil {
		return err
	}

	log.Debugf("Updating the new pruning point to be the new virtual diff parent with an empty diff")
	csm.stageDiff(stagingArea, newPruningPoint, utxo.NewUTXODiff(), nil)

	log.Debugf("Staging the new pruning point %s", newPruningPoint)
	csm.pruningStore.StagePruningPoint(stagingArea, newPruningPoint)
	csm.pruningStore.StagePruningPointCandidate(stagingArea, newPruningPoint)

	log.Debugf("Populating the pruning point with UTXO entries")
	importedPruningPointUTXOIterator, err := csm.pruningStore.ImportedPruningPointUTXOIterator(csm.databaseContext)
	if err != nil {
		return err
	}
	defer importedPruningPointUTXOIterator.Close()

	newPruningPointBlock, err := csm.blockStore.Block(csm.databaseContext, stagingArea, newPruningPoint)
	if err != nil {
		return err
	}

	err = csm.populateTransactionWithUTXOEntriesFromUTXOSet(newPruningPointBlock, importedPruningPointUTXOIterator)
	if err != nil {
		return err
	}

	// Before we manually mark the new pruning point as valid, we validate that all of its transactions are valid
	// against the provided UTXO set.
	log.Debugf("Validating that the pruning point is UTXO valid")
	newPruningPointSelectedParentMedianTime, err := csm.pastMedianTimeManager.PastMedianTime(stagingArea, newPruningPoint)
	if err != nil {
		return err
	}
	log.Tracef("The past median time of pruning block %s is %d",
		newPruningPoint, newPruningPointSelectedParentMedianTime)

	for i, transaction := range newPruningPointBlock.Transactions {
		transactionID := consensushashing.TransactionID(transaction)
		log.Tracef("Validating transaction %s in pruning block %s against "+
			"the pruning point's past UTXO", transactionID, newPruningPoint)
		if i == transactionhelper.CoinbaseTransactionIndex {
			log.Tracef("Skipping transaction %s because it is the coinbase", transactionID)
			continue
		}
		log.Tracef("Validating transaction %s and populating it with mass and fee", transactionID)
		err = csm.transactionValidator.ValidateTransactionInContextAndPopulateFee(
			stagingArea, transaction, newPruningPoint)
		if err != nil {
			return err
		}
		log.Tracef("Validation against the pruning point's past UTXO "+
			"passed for transaction %s", transactionID)
	}

	log.Debugf("Staging the new pruning point as %s", externalapi.StatusUTXOValid)
	csm.blockStatusStore.Stage(stagingArea, newPruningPoint, externalapi.StatusUTXOValid)

	log.Debugf("Staging the new pruning point multiset")
	csm.multisetStore.Stage(stagingArea, newPruningPoint, importedPruningPointMultiset)

	_, err = csm.difficultyManager.StageDAADataAndReturnRequiredDifficulty(stagingArea, model.VirtualBlockHash, false)
	if err != nil {
		return err
	}

	return nil
}

func (csm *consensusStateManager) applyImportedPruningPointUTXOSet(stagingArea *model.StagingArea) error {
	dbTx, err := csm.databaseContext.Begin()
	if err != nil {
		return err
	}

	err = stagingArea.Commit(dbTx)
	if err != nil {
		return err
	}

	log.Debugf("Starting to import virtual UTXO set and pruning point utxo set")
	err = csm.consensusStateStore.StartImportingPruningPointUTXOSet(dbTx)
	if err != nil {
		return err
	}

	log.Debugf("Committing all staged data for imported pruning point")
	err = dbTx.Commit()
	if err != nil {
		return err
	}

	return csm.importVirtualUTXOSetAndPruningPointUTXOSet()
}

func (csm *consensusStateManager) importVirtualUTXOSetAndPruningPointUTXOSet() error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "importVirtualUTXOSetAndPruningPointUTXOSet")
	defer onEnd()

	log.Debugf("Getting an iterator into the imported pruning point utxo set")
	pruningPointUTXOSetIterator, err := csm.pruningStore.ImportedPruningPointUTXOIterator(csm.databaseContext)
	if err != nil {
		return err
	}
	defer pruningPointUTXOSetIterator.Close()

	log.Debugf("Importing the virtual UTXO set")
	err = csm.consensusStateStore.ImportPruningPointUTXOSetIntoVirtualUTXOSet(csm.databaseContext, pruningPointUTXOSetIterator)
	if err != nil {
		return err
	}

	log.Debugf("Importing the new pruning point UTXO set")
	err = csm.pruningStore.CommitImportedPruningPointUTXOSet(csm.databaseContext)
	if err != nil {
		return err
	}

	// Run update virtual to create acceptance data and any other missing data.
	updateVirtualStagingArea := model.NewStagingArea()
	pruningPoint, err := csm.pruningStore.PruningPoint(csm.databaseContext, updateVirtualStagingArea)
	if err != nil {
		return err
	}

	_, _, err = csm.updateVirtual(updateVirtualStagingArea, pruningPoint, []*externalapi.DomainHash{pruningPoint})
	if err != nil {
		return err
	}

	err = staging.CommitAllChanges(csm.databaseContext, updateVirtualStagingArea)
	if err != nil {
		return err
	}

	log.Debugf("Finishing to import virtual UTXO set and pruning point UTXO set")
	return csm.consensusStateStore.FinishImportingPruningPointUTXOSet(csm.databaseContext)
}

func (csm *consensusStateManager) RecoverUTXOIfRequired() error {
	hadStartedImportingPruningPointUTXOSet, err := csm.consensusStateStore.HadStartedImportingPruningPointUTXOSet(csm.databaseContext)
	if err != nil {
		return err
	}
	if !hadStartedImportingPruningPointUTXOSet {
		return nil
	}

	log.Warnf("Unimported pruning point UTXO set detected. Attempting to recover...")
	err = csm.importVirtualUTXOSetAndPruningPointUTXOSet()
	if err != nil {
		return err
	}
	log.Warnf("Unimported UTXO set successfully recovered")
	return nil
}
