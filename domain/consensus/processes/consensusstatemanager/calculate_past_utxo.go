package consensusstatemanager

import (
	"errors"

	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/processes/consensusstatemanager/utxoalgebra"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
)

func (csm *consensusStateManager) calculatePastUTXOAndAcceptanceData(blockHash *externalapi.DomainHash) (
	*model.UTXODiff, model.AcceptanceData, model.Multiset, error) {

	blockGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, blockHash)
	if err != nil {
		return nil, nil, nil, err
	}

	selectedParentPastUTXO, err := csm.restorePastUTXO(blockGHOSTDAGData.SelectedParent)
	if err != nil {
		return nil, nil, nil, err
	}

	acceptanceData, utxoDiff, err := csm.applyBlueBlocks(blockHash, selectedParentPastUTXO, blockGHOSTDAGData)
	if err != nil {
		return nil, nil, nil, err
	}

	multiset, err := csm.calculateMultiset(acceptanceData, blockGHOSTDAGData)
	if err != nil {
		return nil, nil, nil, err
	}

	return utxoDiff, acceptanceData, multiset, nil
}

func (csm *consensusStateManager) restorePastUTXO(blockHash *externalapi.DomainHash) (*model.UTXODiff, error) {
	var err error

	// collect the UTXO diffs
	var utxoDiffs []*model.UTXODiff
	nextBlockHash := blockHash
	for nextBlockHash != nil {
		utxoDiff, err := csm.utxoDiffStore.UTXODiff(csm.databaseContext, nextBlockHash)
		if err != nil {
			return nil, err
		}
		utxoDiffs = append(utxoDiffs, utxoDiff)

		nextBlockHash, err = csm.utxoDiffStore.UTXODiffChild(csm.databaseContext, nextBlockHash)
	}

	// apply the diffs in reverse order
	accumulatedDiff := model.NewUTXODiff()
	for i := len(utxoDiffs) - 1; i >= 0; i-- {
		accumulatedDiff, err = utxoalgebra.WithDiff(accumulatedDiff, utxoDiffs[i])
		if err != nil {
			return nil, err
		}
	}

	return accumulatedDiff, nil
}

func (csm *consensusStateManager) applyBlueBlocks(blockHash *externalapi.DomainHash,
	selectedParentPastUTXODiff *model.UTXODiff, ghostdagData *model.BlockGHOSTDAGData) (
	model.AcceptanceData, *model.UTXODiff, error) {

	blueBlocks, err := csm.blockStore.Blocks(csm.databaseContext, ghostdagData.MergeSetBlues)
	if err != nil {
		return nil, nil, err
	}

	selectedParentMedianTime, err := csm.pastMedianTimeManager.PastMedianTime(ghostdagData.SelectedParent)
	if err != nil {
		return nil, nil, err
	}

	multiblockAcceptanceData := make(model.AcceptanceData, len(blueBlocks))
	accumulatedUTXODiff := utxoalgebra.DiffClone(selectedParentPastUTXODiff)
	accumulatedMass := uint64(0)

	for i, blueBlock := range blueBlocks {
		blockAccepanceData := &model.BlockAcceptanceData{
			TransactionAcceptanceData: []*model.TransactionAcceptanceData{},
		}
		isSelectedParent := i == 0

		for j, transaction := range blueBlock.Transactions {
			var isAccepted bool
			var fee uint64

			isAccepted, accumulatedMass, err = csm.maybeAcceptTransaction(transaction, blockHash, isSelectedParent,
				accumulatedUTXODiff, accumulatedMass, selectedParentMedianTime, ghostdagData.BlueScore)
			if err != nil {
				return nil, nil, err
			}

			blockAccepanceData.TransactionAcceptanceData[j] = &model.TransactionAcceptanceData{
				Transaction: transaction,
				Fee:         fee,
				IsAccepted:  isAccepted,
			}
		}
		multiblockAcceptanceData[i] = blockAccepanceData
	}

	return multiblockAcceptanceData, accumulatedUTXODiff, nil
}

func (csm *consensusStateManager) maybeAcceptTransaction(transaction *externalapi.DomainTransaction,
	blockHash *externalapi.DomainHash, isSelectedParent bool, accumulatedUTXODiff *model.UTXODiff,
	accumulatedMassBefore uint64, selectedParentPastMedianTime int64, blockBlueScore uint64) (
	isAccepted bool, accumulatedMassAfter uint64, err error) {

	err = csm.populateTransactionWithUTXOEntriesFromVirtualOrDiff(transaction, accumulatedUTXODiff)
	if err != nil {
		return false, 0, err
	}

	// Coinbase transaction outputs are added to the UTXO-set only if they are in the selected parent chain.
	if transactionhelper.IsCoinBase(transaction) {
		if !isSelectedParent {
			return false, accumulatedMassBefore, nil
		}

		err := utxoalgebra.DiffAddTransaction(accumulatedUTXODiff, transaction, blockBlueScore)
		if err != nil {
			return false, 0, err
		}

		return true, accumulatedMassBefore, nil
	}

	err = csm.transactionValidator.ValidateTransactionInContextAndPopulateMassAndFee(
		transaction, blockHash, selectedParentPastMedianTime)
	if err != nil {
		if !errors.As(err, &(ruleerrors.RuleError{})) {
			return false, 0, err
		}

		return false, accumulatedMassBefore, nil
	}

	isAccepted, accumulatedMassAfter = csm.checkTransactionMass(transaction, accumulatedMassBefore)

	return isAccepted, accumulatedMassAfter, nil
}

func (csm *consensusStateManager) checkTransactionMass(
	transaction *externalapi.DomainTransaction, accumulatedMassBefore uint64) (
	isAccepted bool, accumulatedMassAfter uint64) {

	accumulatedMassAfter = accumulatedMassBefore + transaction.Mass

	// We could potentially overflow the accumulator so check for
	// overflow as well.
	if accumulatedMassAfter < transaction.Mass || accumulatedMassAfter > constants.MaxMassAcceptedByBlock {
		return false, 0
	}

	return true, accumulatedMassAfter
}
