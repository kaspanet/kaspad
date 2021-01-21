package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/pkg/errors"
)

func (csm *consensusStateManager) ImportPruningPoint(newPruningPoint *externalapi.DomainBlock) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "ImportPruningPoint")
	defer onEnd()

	err := csm.importPruningPoint(newPruningPoint)
	if err != nil {
		csm.discardImportedPruningPointUTXOSetChanges()
		return err
	}

	return csm.applyImportedPruningPointUTXOSet()
}

func (csm *consensusStateManager) importPruningPoint(newPruningPoint *externalapi.DomainBlock) error {
	log.Debugf("importPruningPoint start")
	defer log.Debugf("importPruningPoint end")

	newPruningPointHash := consensushashing.BlockHash(newPruningPoint)

	// We ignore the shouldSendNotification return value because we always want to send finality conflict notification
	// in case the new pruning point violates finality
	isViolatingFinality, _, err := csm.isViolatingFinality(newPruningPointHash)
	if err != nil {
		return err
	}

	if isViolatingFinality {
		log.Warnf("Finality Violation Detected! The suggest pruning point %s violates finality!", newPruningPointHash)
		return errors.Wrapf(ruleerrors.ErrSuggestedPruningViolatesFinality, "%s cannot be a pruning point because "+
			"it violates finality", newPruningPointHash)
	}

	importedPruningPointMultiset, err := csm.pruningStore.ImportedPruningPointMultiset(csm.databaseContext)
	if err != nil {
		return err
	}

	newPruningPointHeader, err := csm.blockHeaderStore.BlockHeader(csm.databaseContext, newPruningPointHash)
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

	log.Debugf("Staging the the pruning point as the only DAG tip")
	newTips := []*externalapi.DomainHash{newPruningPointHash}
	csm.consensusStateStore.StageTips(newTips)

	log.Debugf("Setting the pruning point as the only virtual parent")
	err = csm.dagTopologyManager.SetParents(model.VirtualBlockHash, newTips)
	if err != nil {
		return err
	}

	log.Debugf("Calculating GHOSTDAG for the new virtual")
	err = csm.ghostdagManager.GHOSTDAG(model.VirtualBlockHash)
	if err != nil {
		return err
	}

	log.Debugf("Deleting all existing virtual diff parents")
	csm.consensusStateStore.StageVirtualDiffParents(nil)

	log.Debugf("Updating the new pruning point to be the new virtual diff parent with an empty diff")
	err = csm.stageDiff(newPruningPointHash, utxo.NewUTXODiff(), nil)
	if err != nil {
		return err
	}

	log.Debugf("Staging the new pruning point")
	csm.pruningStore.StagePruningPoint(newPruningPointHash)

	// Before we manually mark the new pruning point as valid, we validate that all of its transactions are valid
	// against the provided UTXO set.
	log.Debugf("Validating that the pruning point is UTXO valid")

	// validateBlockTransactionsAgainstPastUTXO pre-fills the block's transactions inputs, which
	// are assumed to not be pre-filled during further validations.
	// Therefore - clone newPruningPoint before passing it to validateBlockTransactionsAgainstPastUTXO
	err = csm.validateBlockTransactionsAgainstPastUTXO(newPruningPoint.Clone(), utxo.NewUTXODiff())
	if err != nil {
		return err
	}

	log.Debugf("Staging the new pruning point as %s", externalapi.StatusUTXOValid)
	csm.blockStatusStore.Stage(newPruningPointHash, externalapi.StatusUTXOValid)

	log.Debugf("Staging the new pruning point multiset")
	csm.multisetStore.Stage(newPruningPointHash, importedPruningPointMultiset)
	return nil
}

func (csm *consensusStateManager) discardImportedPruningPointUTXOSetChanges() {
	for _, store := range csm.stores {
		store.Discard()
	}
}

func (csm *consensusStateManager) applyImportedPruningPointUTXOSet() error {
	dbTx, err := csm.databaseContext.Begin()
	if err != nil {
		return err
	}

	for _, store := range csm.stores {
		err = store.Commit(dbTx)
		if err != nil {
			return err
		}
	}

	err = dbTx.Commit()
	if err != nil {
		return err
	}

	return csm.importVirtualUTXOSetAndPruningPointUTXOSet()
}

func (csm *consensusStateManager) importVirtualUTXOSetAndPruningPointUTXOSet() error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "importVirtualUTXOSetAndPruningPointUTXOSet")
	defer onEnd()

	log.Debugf("Starting to import virtual UTXO set and pruning point utxo set")
	err := csm.consensusStateStore.StartImportingPruningPointUTXOSet(csm.databaseContext)
	if err != nil {
		return err
	}

	log.Debugf("Getting an iterator into the imported pruning point utxo set")
	pruningPointUTXOSetIterator, err := csm.pruningStore.ImportedPruningPointUTXOIterator(csm.databaseContext)
	if err != nil {
		return err
	}

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
