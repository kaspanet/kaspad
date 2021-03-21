package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
)

// PopulateTransactionWithUTXOEntries populates the transaction UTXO entries with data from the virtual's UTXO set.
func (csm *consensusStateManager) PopulateTransactionWithUTXOEntries(transaction *externalapi.DomainTransaction) error {
	return csm.populateTransactionWithUTXOEntriesFromVirtualOrDiff(transaction, nil)
}

// populateTransactionWithUTXOEntriesFromVirtualOrDiff populates the transaction UTXO entries with data
// from the virtual's UTXO set combined with the provided utxoDiff.
// If utxoDiff == nil UTXO entries are taken from the virtual's UTXO set only
func (csm *consensusStateManager) populateTransactionWithUTXOEntriesFromVirtualOrDiff(
	transaction *externalapi.DomainTransaction, utxoDiff externalapi.UTXODiff) error {

	transactionID := consensushashing.TransactionID(transaction)
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
			if utxoEntry, ok := utxoDiff.ToAdd().Get(&transactionInput.PreviousOutpoint); ok {
				log.Tracef("Populating outpoint %s:%d from the given utxoDiff",
					transactionInput.PreviousOutpoint.TransactionID, transactionInput.PreviousOutpoint.Index)
				transactionInput.UTXOEntry = utxoEntry
				continue
			}

			if utxoDiff.ToRemove().Contains(&transactionInput.PreviousOutpoint) {
				log.Tracef("Outpoint %s:%d is missing in the given utxoDiff",
					transactionInput.PreviousOutpoint.TransactionID, transactionInput.PreviousOutpoint.Index)
				missingOutpoints = append(missingOutpoints, &transactionInput.PreviousOutpoint)
				continue
			}
		}

		// Check for the input's outpoint in virtual's UTXO set.
		hasUTXOEntry, err := csm.consensusStateStore.HasUTXOByOutpoint(csm.databaseContext, nil, &transactionInput.PreviousOutpoint)
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
		utxoEntry, err := csm.consensusStateStore.UTXOByOutpoint(csm.databaseContext, nil, &transactionInput.PreviousOutpoint)
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

func (csm *consensusStateManager) populateTransactionWithUTXOEntriesFromUTXOSet(
	pruningPoint *externalapi.DomainBlock, iterator externalapi.ReadOnlyUTXOSetIterator) error {

	// Collect the required outpoints from the block
	outpointsForPopulation := make(map[externalapi.DomainOutpoint]interface{})
	for _, transaction := range pruningPoint.Transactions {
		for _, input := range transaction.Inputs {
			outpointsForPopulation[input.PreviousOutpoint] = struct{}{}
		}
	}

	// Collect the UTXO entries from the iterator
	outpointsToUTXOEntries := make(map[externalapi.DomainOutpoint]externalapi.UTXOEntry, len(outpointsForPopulation))
	for ok := iterator.First(); ok; ok = iterator.Next() {
		outpoint, utxoEntry, err := iterator.Get()
		if err != nil {
			return err
		}
		outpointValue := *outpoint
		if _, ok := outpointsForPopulation[outpointValue]; ok {
			outpointsToUTXOEntries[outpointValue] = utxoEntry
		}
		if len(outpointsForPopulation) == len(outpointsToUTXOEntries) {
			break
		}
	}

	// Populate the block with the collected UTXO entries
	var missingOutpoints []*externalapi.DomainOutpoint
	for _, transaction := range pruningPoint.Transactions {
		for _, input := range transaction.Inputs {
			utxoEntry, ok := outpointsToUTXOEntries[input.PreviousOutpoint]
			if !ok {
				missingOutpoints = append(missingOutpoints, &input.PreviousOutpoint)
				continue
			}
			input.UTXOEntry = utxoEntry
		}
	}

	if len(missingOutpoints) > 0 {
		return ruleerrors.NewErrMissingTxOut(missingOutpoints)
	}
	return nil
}
