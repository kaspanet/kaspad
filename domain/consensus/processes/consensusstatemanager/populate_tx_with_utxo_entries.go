package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/processes/consensusstatemanager/utxoalgebra"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
)

// PopulateTransactionWithUTXOEntries populates the transaction UTXO entries with data from the virtual's UTXO set.
func (csm *consensusStateManager) PopulateTransactionWithUTXOEntries(transaction *externalapi.DomainTransaction) error {
	return csm.populateTransactionWithUTXOEntriesFromVirtualOrDiff(transaction, nil)
}

// populateTransactionWithUTXOEntriesFromVirtualOrDiff populates the transaction UTXO entries with data
// from the virtual's UTXO set combined with the provided utxoDiff.
// If utxoDiff == nil UTXO entries are taken from the virtual's UTXO set only
func (csm *consensusStateManager) populateTransactionWithUTXOEntriesFromVirtualOrDiff(
	transaction *externalapi.DomainTransaction, utxoDiff *model.UTXODiff) error {

	transactionID := consensusserialization.TransactionID(transaction)
	log.Tracef("populateTransactionWithUTXOEntriesFromVirtualOrDiff start for transaction %s", transactionID)
	defer log.Tracef("populateTransactionWithUTXOEntriesFromVirtualOrDiff end for transaction %s", transactionID)

	var missingOutpoints []*externalapi.DomainOutpoint
	for _, transactionInput := range transaction.Inputs {
		// skip all inputs that have a pre-filled utxo entry
		if transactionInput.UTXOEntry != nil {
			log.Tracef("Skipping outpoint %s:%d because it is already populated",
				transactionInput.PreviousOutpoint.TransactionID, transactionInput.PreviousOutpoint.Index)
			continue
		}

		// check if utxoDiff says anything about the input's outpoint
		if utxoDiff != nil {
			if utxoEntry, ok := utxoalgebra.CollectionGet(utxoDiff.ToAdd, &transactionInput.PreviousOutpoint); ok {
				log.Tracef("Populating outpoint %s:%d from the given utxoDiff",
					transactionInput.PreviousOutpoint.TransactionID, transactionInput.PreviousOutpoint.Index)
				transactionInput.UTXOEntry = utxoEntry
				continue
			}

			if utxoalgebra.CollectionContains(utxoDiff.ToRemove, &transactionInput.PreviousOutpoint) {
				log.Tracef("Outpoint %s:%d is missing in the given utxoDiff",
					transactionInput.PreviousOutpoint.TransactionID, transactionInput.PreviousOutpoint.Index)
				missingOutpoints = append(missingOutpoints, &transactionInput.PreviousOutpoint)
				continue
			}
		}

		// Check for the input's outpoint in virtual's UTXO set.
		hasUTXOEntry, err := csm.consensusStateStore.HasUTXOByOutpoint(csm.databaseContext, &transactionInput.PreviousOutpoint)
		if err != nil {
			return err
		}
		if !hasUTXOEntry {
			log.Tracef("Outpoint %s:%d is missing in the database",
				transactionInput.PreviousOutpoint.TransactionID, transactionInput.PreviousOutpoint.Index)
			missingOutpoints = append(missingOutpoints, &transactionInput.PreviousOutpoint)
			continue
		}

		log.Tracef("Populating outpoint %s:%d from the database",
			transactionInput.PreviousOutpoint.TransactionID, transactionInput.PreviousOutpoint.Index)
		utxoEntry, err := csm.consensusStateStore.UTXOByOutpoint(csm.databaseContext, &transactionInput.PreviousOutpoint)
		if err != nil {
			return err
		}
		transactionInput.UTXOEntry = utxoEntry
	}

	if len(missingOutpoints) > 0 {
		return ruleerrors.NewErrMissingTxOut(missingOutpoints)
	}

	return nil
}
