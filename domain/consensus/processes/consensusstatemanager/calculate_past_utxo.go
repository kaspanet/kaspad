package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/multiset"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
)

func (csm *consensusStateManager) CalculatePastUTXOAndAcceptanceData(blockHash *externalapi.DomainHash) (
	model.UTXODiff, model.AcceptanceData, model.Multiset, error) {

	log.Tracef("CalculatePastUTXOAndAcceptanceData start for block %s", blockHash)
	defer log.Tracef("CalculatePastUTXOAndAcceptanceData end for block %s", blockHash)

	if *blockHash == *csm.genesisHash {
		log.Tracef("Block %s is the genesis. By definition, "+
			"it has an empty UTXO diff, empty acceptance data, and a blank multiset", blockHash)
		return utxo.NewUTXODiff(), model.AcceptanceData{}, multiset.New(), nil
	}

	blockGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, blockHash)
	if err != nil {
		return nil, nil, nil, err
	}

	log.Tracef("Restoring the past UTXO of block %s with selectedParent %s",
		blockHash, blockGHOSTDAGData.SelectedParent())
	selectedParentPastUTXO, err := csm.restorePastUTXO(blockGHOSTDAGData.SelectedParent())
	if err != nil {
		return nil, nil, nil, err
	}

	log.Tracef("Applying blue blocks to the selected parent past UTXO of block %s", blockHash)
	acceptanceData, utxoDiff, err := csm.applyBlueBlocks(blockHash, selectedParentPastUTXO, blockGHOSTDAGData)
	if err != nil {
		return nil, nil, nil, err
	}

	log.Tracef("Calculating the multiset of %s", blockHash)
	multiset, err := csm.calculateMultiset(acceptanceData, blockGHOSTDAGData)
	if err != nil {
		return nil, nil, nil, err
	}
	log.Tracef("The multiset of block %s resolved to: %s", blockHash, multiset.Hash())

	return utxoDiff.ToImmutable(), acceptanceData, multiset, nil
}

func (csm *consensusStateManager) restorePastUTXO(blockHash *externalapi.DomainHash) (model.MutableUTXODiff, error) {
	log.Tracef("restorePastUTXO start for block %s", blockHash)
	defer log.Tracef("restorePastUTXO end for block %s", blockHash)

	var err error

	log.Tracef("Collecting UTXO diffs for block %s", blockHash)
	var utxoDiffs []model.UTXODiff
	nextBlockHash := blockHash
	for {
		log.Tracef("Collecting UTXO diff for block %s", nextBlockHash)
		utxoDiff, err := csm.utxoDiffStore.UTXODiff(csm.databaseContext, nextBlockHash)
		if err != nil {
			return nil, err
		}
		utxoDiffs = append(utxoDiffs, utxoDiff)
		log.Tracef("Collected UTXO diff for block %s: %s", nextBlockHash, utxoDiff)

		exists, err := csm.utxoDiffStore.HasUTXODiffChild(csm.databaseContext, nextBlockHash)
		if err != nil {
			return nil, err
		}
		if !exists {
			log.Tracef("Block %s does not have a UTXO diff child, "+
				"meaning we reached the virtual. Returning the collected "+
				"UTXO diffs: %s", nextBlockHash, utxoDiffs)
			break
		}

		nextBlockHash, err = csm.utxoDiffStore.UTXODiffChild(csm.databaseContext, nextBlockHash)
		if err != nil {
			return nil, err
		}
		if nextBlockHash == nil {
			log.Tracef("Block %s does not have a UTXO diff child, "+
				"meaning we reached the virtual. Returning the collected "+
				"UTXO diffs: %s", nextBlockHash, utxoDiffs)
			break
		}
	}

	// apply the diffs in reverse order
	log.Tracef("Applying the collected UTXO diffs for block %s in reverse order", blockHash)
	accumulatedDiff := utxo.NewMutableUTXODiff()
	for i := len(utxoDiffs) - 1; i >= 0; i-- {
		err = accumulatedDiff.WithDiffInPlace(utxoDiffs[i])
		if err != nil {
			return nil, err
		}
	}
	log.Tracef("The accumulated diff for block %s is: %s", blockHash, accumulatedDiff)

	return accumulatedDiff, nil
}

func (csm *consensusStateManager) applyBlueBlocks(blockHash *externalapi.DomainHash,
	selectedParentPastUTXODiff model.MutableUTXODiff, ghostdagData model.BlockGHOSTDAGData) (
	model.AcceptanceData, model.MutableUTXODiff, error) {

	log.Tracef("applyBlueBlocks start for block %s", blockHash)
	defer log.Tracef("applyBlueBlocks end for block %s", blockHash)

	blueBlocks, err := csm.blockStore.Blocks(csm.databaseContext, ghostdagData.MergeSetBlues())
	if err != nil {
		return nil, nil, err
	}

	selectedParentMedianTime, err := csm.pastMedianTimeManager.PastMedianTime(blockHash)
	if err != nil {
		return nil, nil, err
	}
	log.Tracef("The past median time for block %s is: %d", blockHash, selectedParentMedianTime)

	multiblockAcceptanceData := make(model.AcceptanceData, len(blueBlocks))
	accumulatedUTXODiff := selectedParentPastUTXODiff
	accumulatedMass := uint64(0)

	for i, blueBlock := range blueBlocks {
		blueBlockHash := consensushashing.BlockHash(blueBlock)
		log.Tracef("Applying blue block %s", blueBlockHash)
		blockAcceptanceData := &model.BlockAcceptanceData{
			TransactionAcceptanceData: make([]*model.TransactionAcceptanceData, len(blueBlock.Transactions)),
		}
		isSelectedParent := i == 0
		log.Tracef("Is blue block %s the selected parent: %t", blueBlockHash, isSelectedParent)

		for j, transaction := range blueBlock.Transactions {
			var isAccepted bool

			transactionID := consensushashing.TransactionID(transaction)
			log.Tracef("Attempting to accept transaction %s in block %s",
				transactionID, blueBlockHash)

			isAccepted, accumulatedMass, err = csm.maybeAcceptTransaction(transaction, blockHash, isSelectedParent,
				accumulatedUTXODiff, accumulatedMass, selectedParentMedianTime, ghostdagData.BlueScore())
			if err != nil {
				return nil, nil, err
			}
			log.Tracef("Transaction %s in block %s isAccepted: %t, fee: %d",
				transactionID, blueBlockHash, isAccepted, transaction.Fee)

			blockAcceptanceData.TransactionAcceptanceData[j] = &model.TransactionAcceptanceData{
				Transaction: transaction,
				Fee:         transaction.Fee,
				IsAccepted:  isAccepted,
			}
		}
		multiblockAcceptanceData[i] = blockAcceptanceData
	}

	return multiblockAcceptanceData, accumulatedUTXODiff, nil
}

func (csm *consensusStateManager) maybeAcceptTransaction(transaction *externalapi.DomainTransaction,
	blockHash *externalapi.DomainHash, isSelectedParent bool, accumulatedUTXODiff model.MutableUTXODiff,
	accumulatedMassBefore uint64, selectedParentPastMedianTime int64, blockBlueScore uint64) (
	isAccepted bool, accumulatedMassAfter uint64, err error) {

	transactionID := consensushashing.TransactionID(transaction)
	log.Tracef("maybeAcceptTransaction start for transaction %s in block %s", transactionID, blockHash)
	defer log.Tracef("maybeAcceptTransaction end for transaction %s in block %s", transactionID, blockHash)

	log.Tracef("Populating transaction %s with UTXO entries", transactionID)
	err = csm.populateTransactionWithUTXOEntriesFromVirtualOrDiff(transaction, accumulatedUTXODiff.ToImmutable())
	if err != nil {
		if !errors.As(err, &(ruleerrors.RuleError{})) {
			return false, 0, err
		}

		return false, accumulatedMassBefore, nil
	}

	// Coinbase transaction outputs are added to the UTXO-set only if they are in the selected parent chain.
	if transactionhelper.IsCoinBase(transaction) {
		if !isSelectedParent {
			log.Tracef("Transaction %s is the coinbase of block %s "+
				"but said block is not in the selected parent chain. "+
				"As such, it is not accepted", transactionID, blockHash)
			return false, accumulatedMassBefore, nil
		}
		log.Tracef("Transaction %s is the coinbase of block %s", transactionID, blockHash)
	} else {
		log.Tracef("Validating transaction %s in block %s", transactionID, blockHash)
		err = csm.transactionValidator.ValidateTransactionInContextAndPopulateMassAndFee(
			transaction, blockHash, selectedParentPastMedianTime)
		if err != nil {
			if !errors.As(err, &(ruleerrors.RuleError{})) {
				return false, 0, err
			}

			log.Tracef("Validation failed for transaction %s "+
				"in block %s: %s", transactionID, blockHash, err)
			return false, accumulatedMassBefore, nil
		}
		log.Tracef("Validation passed for transaction %s in block %s", transactionID, blockHash)

		log.Tracef("Check mass for transaction %s in block %s", transactionID, blockHash)
		isAccepted, accumulatedMassAfter = csm.checkTransactionMass(transaction, accumulatedMassBefore)
		if !isAccepted {
			log.Tracef("Transaction %s in block %s has too much mass, "+
				"and cannot be accepted", transactionID, blockHash)
			return false, accumulatedMassBefore, nil
		}
	}

	log.Tracef("Adding transaction %s in block %s to the accumulated diff", transactionID, blockHash)
	err = accumulatedUTXODiff.AddTransaction(transaction, blockBlueScore)
	if err != nil {
		return false, 0, err
	}

	return true, accumulatedMassAfter, nil
}

func (csm *consensusStateManager) checkTransactionMass(
	transaction *externalapi.DomainTransaction, accumulatedMassBefore uint64) (
	isAccepted bool, accumulatedMassAfter uint64) {

	transactionID := consensushashing.TransactionID(transaction)
	log.Tracef("checkTransactionMass start for transaction %s", transactionID)
	defer log.Tracef("checkTransactionMass end for transaction %s", transactionID)

	log.Tracef("Adding transaction %s with mass %d to the "+
		"so-far accumulated mass of %d", transactionID, transaction.Mass, accumulatedMassBefore)
	accumulatedMassAfter = accumulatedMassBefore + transaction.Mass
	log.Tracef("Accumulated mass including transaction %s: %d", transactionID, accumulatedMassAfter)

	// We could potentially overflow the accumulator so check for
	// overflow as well.
	if accumulatedMassAfter < transaction.Mass || accumulatedMassAfter > csm.maxMassAcceptedByBlock {
		return false, 0
	}

	return true, accumulatedMassAfter
}

// RestorePastUTXOSetIterator restores the given block's UTXOSet iterator, and returns it as a model.ReadOnlyUTXOSetIterator
func (csm *consensusStateManager) RestorePastUTXOSetIterator(blockHash *externalapi.DomainHash) (
	model.ReadOnlyUTXOSetIterator, error) {

	blockStatus, err := csm.resolveBlockStatus(blockHash)
	if err != nil {
		return nil, err
	}
	if blockStatus != externalapi.StatusUTXOValid {
		return nil, errors.Errorf(
			"block %s, has status '%s', and therefore can't restore it's UTXO set. Only blocks with status '%s' can be restored.",
			blockHash, blockStatus, externalapi.StatusUTXOValid)
	}

	log.Tracef("RestorePastUTXOSetIterator start for block %s", blockHash)
	defer log.Tracef("RestorePastUTXOSetIterator end for block %s", blockHash)

	log.Tracef("Calculating UTXO diff for block %s", blockHash)
	blockDiff, err := csm.restorePastUTXO(blockHash)
	if err != nil {
		return nil, err
	}

	virtualUTXOSetIterator, err := csm.consensusStateStore.VirtualUTXOSetIterator(csm.databaseContext)
	if err != nil {
		return nil, err
	}

	return utxo.IteratorWithDiff(virtualUTXOSetIterator, blockDiff.ToImmutable())
}
