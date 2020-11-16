package consensusstatemanager

import (
	"errors"

	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/multiset"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/processes/consensusstatemanager/utxoalgebra"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
)

func (csm *consensusStateManager) CalculatePastUTXOAndAcceptanceData(blockHash *externalapi.DomainHash) (
	*model.UTXODiff, model.AcceptanceData, model.Multiset, error) {

	// The genesis block has an empty UTXO diff, empty acceptance data, and a blank multiset
	if *blockHash == *csm.genesisHash {
		return &model.UTXODiff{}, model.AcceptanceData{}, multiset.New(), nil
	}

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
	for {
		utxoDiff, err := csm.utxoDiffStore.UTXODiff(csm.databaseContext, nextBlockHash)
		if err != nil {
			return nil, err
		}
		utxoDiffs = append(utxoDiffs, utxoDiff)

		exists, err := csm.utxoDiffStore.HasUTXODiffChild(csm.databaseContext, nextBlockHash)
		if err != nil {
			return nil, err
		}
		if !exists {
			break
		}

		nextBlockHash, err = csm.utxoDiffStore.UTXODiffChild(csm.databaseContext, nextBlockHash)
		if err != nil {
			return nil, err
		}
		if nextBlockHash == nil {
			break
		}
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

	selectedParentMedianTime, err := csm.pastMedianTimeManager.PastMedianTime(blockHash)
	if err != nil {
		return nil, nil, err
	}

	multiblockAcceptanceData := make(model.AcceptanceData, len(blueBlocks))
	accumulatedUTXODiff := utxoalgebra.DiffClone(selectedParentPastUTXODiff)
	accumulatedMass := uint64(0)

	for i, blueBlock := range blueBlocks {
		blockAcceptanceData := &model.BlockAcceptanceData{
			TransactionAcceptanceData: make([]*model.TransactionAcceptanceData, len(blueBlock.Transactions)),
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

			blockAcceptanceData.TransactionAcceptanceData[j] = &model.TransactionAcceptanceData{
				Transaction: transaction,
				Fee:         fee,
				IsAccepted:  isAccepted,
			}
		}
		multiblockAcceptanceData[i] = blockAcceptanceData
	}

	return multiblockAcceptanceData, accumulatedUTXODiff, nil
}

func (csm *consensusStateManager) maybeAcceptTransaction(transaction *externalapi.DomainTransaction,
	blockHash *externalapi.DomainHash, isSelectedParent bool, accumulatedUTXODiff *model.UTXODiff,
	accumulatedMassBefore uint64, selectedParentPastMedianTime int64, blockBlueScore uint64) (
	isAccepted bool, accumulatedMassAfter uint64, err error) {

	err = csm.populateTransactionWithUTXOEntriesFromVirtualOrDiff(transaction, accumulatedUTXODiff)
	if err != nil {
		if !errors.As(err, &(ruleerrors.RuleError{})) {
			return false, 0, err
		}

		return false, accumulatedMassBefore, nil
	}

	// Coinbase transaction outputs are added to the UTXO-set only if they are in the selected parent chain.
	if transactionhelper.IsCoinBase(transaction) {
		if !isSelectedParent {
			return false, accumulatedMassBefore, nil
		}
	} else {
		err = csm.transactionValidator.ValidateTransactionInContextAndPopulateMassAndFee(
			transaction, blockHash, selectedParentPastMedianTime)
		if err != nil {
			if !errors.As(err, &(ruleerrors.RuleError{})) {
				return false, 0, err
			}

			return false, accumulatedMassBefore, nil
		}

		isAccepted, accumulatedMassAfter = csm.checkTransactionMass(transaction, accumulatedMassBefore)
	}

	err = utxoalgebra.DiffAddTransaction(accumulatedUTXODiff, transaction, blockBlueScore)
	if err != nil {
		return false, 0, err
	}

	return true, accumulatedMassAfter, nil
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

func (csm *consensusStateManager) RestorePastUTXOSetIterator(blockHash *externalapi.DomainHash) (model.ReadOnlyUTXOSetIterator, error) {
	diff, _, _, err := csm.CalculatePastUTXOAndAcceptanceData(blockHash)
	if err != nil {
		return nil, err
	}

	virtualUTXOSetIterator, err := csm.consensusStateStore.VirtualUTXOSetIterator(csm.databaseContext)
	if err != nil {
		return nil, err
	}

	pastUTXO := model.NewUTXODiff()
	for virtualUTXOSetIterator.Next() {
		outpoint, utxoEntry, err := virtualUTXOSetIterator.Get()
		if err != nil {
			return nil, err
		}
		pastUTXO.ToAdd[*outpoint] = utxoEntry
	}

	diff, err = utxoalgebra.WithDiff(pastUTXO, diff)
	if err != nil {
		return nil, err
	}

	if len(diff.ToRemove) > 0 {
		return nil, errors.New("diff.ToRemove is expected to be empty")
	}

	return newUTXOSetIterator(diff.ToAdd), nil
}

type utxoOutpointEntryPair struct {
	outpoint externalapi.DomainOutpoint
	entry    *externalapi.UTXOEntry
}

type utxoSetIterator struct {
	index int
	pairs []utxoOutpointEntryPair
}

func newUTXOSetIterator(collection model.UTXOCollection) *utxoSetIterator {
	pairs := make([]utxoOutpointEntryPair, len(collection))
	i := 0
	for outpoint, entry := range collection {
		pairs[i] = utxoOutpointEntryPair{
			outpoint: outpoint,
			entry:    entry,
		}
		i++
	}
	return &utxoSetIterator{index: -1, pairs: pairs}
}

func (u *utxoSetIterator) Next() bool {
	u.index++
	return u.index < len(u.pairs)
}

func (u *utxoSetIterator) Get() (outpoint *externalapi.DomainOutpoint, utxoEntry *externalapi.UTXOEntry, err error) {
	pair := u.pairs[u.index]
	return &pair.outpoint, pair.entry, nil
}
