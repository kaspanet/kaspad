package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/multiset"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
)

func (csm *consensusStateManager) CalculatePastUTXOAndAcceptanceData(stagingArea *model.StagingArea,
	blockHash *externalapi.DomainHash) (externalapi.UTXODiff, externalapi.AcceptanceData, model.Multiset, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "CalculatePastUTXOAndAcceptanceData")
	defer onEnd()

	log.Debugf("CalculatePastUTXOAndAcceptanceData start for block %s", blockHash)

	if blockHash.Equal(csm.genesisHash) {
		log.Debugf("Block %s is the genesis. By definition, "+
			"it has an empty UTXO diff, empty acceptance data, and a blank multiset", blockHash)
		return utxo.NewUTXODiff(), externalapi.AcceptanceData{}, multiset.New(), nil
	}

	blockGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, nil, nil, err
	}

	log.Debugf("Restoring the past UTXO of block %s with selectedParent %s",
		blockHash, blockGHOSTDAGData.SelectedParent())
	selectedParentPastUTXO, err := csm.restorePastUTXO(stagingArea, blockGHOSTDAGData.SelectedParent())
	if err != nil {
		return nil, nil, nil, err
	}

	log.Debugf("Restored the past UTXO of block %s with selectedParent %s. "+
		"Diff toAdd length: %d, toRemove length: %d", blockHash, blockGHOSTDAGData.SelectedParent(),
		selectedParentPastUTXO.ToAdd().Len(), selectedParentPastUTXO.ToRemove().Len())

	return csm.calculatePastUTXOAndAcceptanceDataWithSelectedParentUTXO(stagingArea, blockHash, selectedParentPastUTXO)
}

func (csm *consensusStateManager) calculatePastUTXOAndAcceptanceDataWithSelectedParentUTXO(stagingArea *model.StagingArea,
	blockHash *externalapi.DomainHash, selectedParentPastUTXO externalapi.UTXODiff) (
	externalapi.UTXODiff, externalapi.AcceptanceData, model.Multiset, error) {

	blockGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, nil, nil, err
	}

	daaScore, err := csm.daaBlocksStore.DAAScore(csm.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, nil, nil, err
	}

	log.Debugf("Applying blue blocks to the selected parent past UTXO of block %s", blockHash)
	acceptanceData, utxoDiff, err := csm.applyMergeSetBlocks(
		stagingArea, blockHash, selectedParentPastUTXO, blockGHOSTDAGData, daaScore)
	if err != nil {
		return nil, nil, nil, err
	}

	log.Debugf("Calculating the multiset of %s", blockHash)
	multiset, err := csm.calculateMultiset(stagingArea, acceptanceData, blockGHOSTDAGData, daaScore)
	if err != nil {
		return nil, nil, nil, err
	}
	log.Debugf("The multiset of block %s resolved to: %s", blockHash, multiset.Hash())

	return utxoDiff.ToImmutable(), acceptanceData, multiset, nil
}

func (csm *consensusStateManager) restorePastUTXO(
	stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (externalapi.UTXODiff, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "restorePastUTXO")
	defer onEnd()

	log.Debugf("restorePastUTXO start for block %s", blockHash)

	var err error

	log.Debugf("Collecting UTXO diffs for block %s", blockHash)
	var utxoDiffs []externalapi.UTXODiff
	nextBlockHash := blockHash
	for {
		log.Debugf("Collecting UTXO diff for block %s", nextBlockHash)
		utxoDiff, err := csm.utxoDiffStore.UTXODiff(csm.databaseContext, stagingArea, nextBlockHash)
		if err != nil {
			return nil, err
		}
		utxoDiffs = append(utxoDiffs, utxoDiff)
		log.Debugf("Collected UTXO diff for block %s: toAdd: %d, toRemove: %d",
			nextBlockHash, utxoDiff.ToAdd().Len(), utxoDiff.ToRemove().Len())

		exists, err := csm.utxoDiffStore.HasUTXODiffChild(csm.databaseContext, stagingArea, nextBlockHash)
		if err != nil {
			return nil, err
		}
		if !exists {
			log.Debugf("Block %s does not have a UTXO diff child, "+
				"meaning we reached the virtual", nextBlockHash)
			break
		}

		nextBlockHash, err = csm.utxoDiffStore.UTXODiffChild(csm.databaseContext, stagingArea, nextBlockHash)
		if err != nil {
			return nil, err
		}
		if nextBlockHash == nil {
			log.Debugf("Block %s does not have a UTXO diff child, "+
				"meaning we reached the virtual", nextBlockHash)
			break
		}
	}

	// apply the diffs in reverse order
	log.Debugf("Applying the collected UTXO diffs for block %s in reverse order", blockHash)
	accumulatedDiff := utxo.NewMutableUTXODiff()
	for i := len(utxoDiffs) - 1; i >= 0; i-- {
		err = accumulatedDiff.WithDiffInPlace(utxoDiffs[i])
		if err != nil {
			return nil, err
		}
	}
	log.Tracef("The accumulated diff for block %s is: %s", blockHash, accumulatedDiff)

	return accumulatedDiff.ToImmutable(), nil
}

func (csm *consensusStateManager) applyMergeSetBlocks(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash,
	selectedParentPastUTXODiff externalapi.UTXODiff, ghostdagData *model.BlockGHOSTDAGData, daaScore uint64) (
	externalapi.AcceptanceData, externalapi.MutableUTXODiff, error) {

	log.Debugf("applyMergeSetBlocks start for block %s", blockHash)
	defer log.Debugf("applyMergeSetBlocks end for block %s", blockHash)

	mergeSetHashes := ghostdagData.MergeSet()
	log.Debugf("Merge set for block %s is %v", blockHash, mergeSetHashes)
	mergeSetBlocks, err := csm.blockStore.Blocks(csm.databaseContext, stagingArea, mergeSetHashes)
	if err != nil {
		return nil, nil, err
	}

	selectedParentMedianTime, err := csm.pastMedianTimeManager.PastMedianTime(stagingArea, blockHash)
	if err != nil {
		return nil, nil, err
	}
	log.Tracef("The past median time for block %s is: %d", blockHash, selectedParentMedianTime)

	multiblockAcceptanceData := make(externalapi.AcceptanceData, len(mergeSetBlocks))
	accumulatedUTXODiff := selectedParentPastUTXODiff.CloneMutable()
	accumulatedMass := uint64(0)

	for i, mergeSetBlock := range mergeSetBlocks {
		mergeSetBlockHash := consensushashing.BlockHash(mergeSetBlock)
		log.Tracef("Applying merge set block %s", mergeSetBlockHash)
		blockAcceptanceData := &externalapi.BlockAcceptanceData{
			BlockHash:                 mergeSetBlockHash,
			TransactionAcceptanceData: make([]*externalapi.TransactionAcceptanceData, len(mergeSetBlock.Transactions)),
		}
		isSelectedParent := i == 0
		log.Tracef("Is merge set block %s the selected parent: %t", mergeSetBlockHash, isSelectedParent)

		for j, transaction := range mergeSetBlock.Transactions {
			var isAccepted bool

			transactionID := consensushashing.TransactionID(transaction)
			log.Tracef("Attempting to accept transaction %s in block %s",
				transactionID, mergeSetBlockHash)

			isAccepted, accumulatedMass, err = csm.maybeAcceptTransaction(stagingArea, transaction, blockHash,
				isSelectedParent, accumulatedUTXODiff, accumulatedMass, selectedParentMedianTime, daaScore)
			if err != nil {
				return nil, nil, err
			}
			log.Tracef("Transaction %s in block %s isAccepted: %t, fee: %d",
				transactionID, mergeSetBlockHash, isAccepted, transaction.Fee)

			var transactionInputUTXOEntries []externalapi.UTXOEntry
			if isAccepted {
				transactionInputUTXOEntries = make([]externalapi.UTXOEntry, len(transaction.Inputs))
				for k, input := range transaction.Inputs {
					transactionInputUTXOEntries[k] = input.UTXOEntry
				}
			}

			blockAcceptanceData.TransactionAcceptanceData[j] = &externalapi.TransactionAcceptanceData{
				Transaction:                 transaction,
				Fee:                         transaction.Fee,
				IsAccepted:                  isAccepted,
				TransactionInputUTXOEntries: transactionInputUTXOEntries,
			}
		}
		multiblockAcceptanceData[i] = blockAcceptanceData
	}

	return multiblockAcceptanceData, accumulatedUTXODiff, nil
}

func (csm *consensusStateManager) maybeAcceptTransaction(stagingArea *model.StagingArea,
	transaction *externalapi.DomainTransaction, blockHash *externalapi.DomainHash, isSelectedParent bool,
	accumulatedUTXODiff externalapi.MutableUTXODiff, accumulatedMassBefore uint64, selectedParentPastMedianTime int64,
	blockDAAScore uint64) (isAccepted bool, accumulatedMassAfter uint64, err error) {

	transactionID := consensushashing.TransactionID(transaction)
	log.Tracef("maybeAcceptTransaction start for transaction %s in block %s", transactionID, blockHash)
	defer log.Tracef("maybeAcceptTransaction end for transaction %s in block %s", transactionID, blockHash)

	log.Tracef("Populating transaction %s with UTXO entries", transactionID)
	err = csm.populateTransactionWithUTXOEntriesFromVirtualOrDiff(stagingArea, transaction, accumulatedUTXODiff.ToImmutable())
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
			stagingArea, transaction, blockHash, selectedParentPastMedianTime)
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
	err = accumulatedUTXODiff.AddTransaction(transaction, blockDAAScore)
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

// RestorePastUTXOSetIterator restores the given block's UTXOSet iterator, and returns it as a externalapi.ReadOnlyUTXOSetIterator
func (csm *consensusStateManager) RestorePastUTXOSetIterator(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (externalapi.ReadOnlyUTXOSetIterator, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "RestorePastUTXOSetIterator")
	defer onEnd()

	blockStatus, err := csm.resolveBlockStatus(stagingArea, blockHash)
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

	log.Debugf("Calculating UTXO diff for block %s", blockHash)
	blockDiff, err := csm.restorePastUTXO(stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	virtualUTXOSetIterator, err := csm.consensusStateStore.VirtualUTXOSetIterator(csm.databaseContext, stagingArea)
	if err != nil {
		return nil, err
	}

	return utxo.IteratorWithDiff(virtualUTXOSetIterator, blockDiff)
}
