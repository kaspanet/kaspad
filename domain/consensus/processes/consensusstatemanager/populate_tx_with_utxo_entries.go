package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/processes/consensusstatemanager/utxoalgebra"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
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

	missingOutpoints := []*externalapi.DomainOutpoint{}

	for _, transactionInput := range transaction.Inputs {
		// skip all inputs that have a pre-filled utxo entry
		if transactionInput.UTXOEntry != nil {
			continue
		}

		// check if utxoDiff says anything about the input's outpoint
		if utxoDiff != nil {
			if utxoEntry, ok := utxoalgebra.CollectionGet(utxoDiff.ToAdd, &transactionInput.PreviousOutpoint); ok {
				transactionInput.UTXOEntry = utxoEntry
				continue
			}

			if utxoalgebra.CollectionContains(utxoDiff.ToRemove, &transactionInput.PreviousOutpoint) {
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
			missingOutpoints = append(missingOutpoints, &transactionInput.PreviousOutpoint)
			continue
		}

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
