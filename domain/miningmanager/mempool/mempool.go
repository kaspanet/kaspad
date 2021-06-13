package mempool

import (
	"sync"

	"github.com/kaspanet/kaspad/domain/dagconfig"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool/model"
	miningmanagermodel "github.com/kaspanet/kaspad/domain/miningmanager/model"
	"github.com/pkg/errors"
)

type mempool struct {
	mtx sync.RWMutex

	config    *config
	consensus externalapi.Consensus

	mempoolUTXOSet   *mempoolUTXOSet
	transactionsPool *transactionsPool
	orphansPool      *orphansPool
}

// New constructs a new mempool
func New(consensus externalapi.Consensus, dagParams *dagconfig.Params) miningmanagermodel.Mempool {
	mp := &mempool{
		config:    defaultConfig(dagParams),
		consensus: consensus,
	}

	mp.mempoolUTXOSet = newMempoolUTXOSet(mp)
	mp.transactionsPool = newTransactionsPool(mp)
	mp.orphansPool = newOrphansPool(mp)

	return mp
}

func (mp *mempool) GetTransaction(transactionID *externalapi.DomainTransactionID) (*externalapi.DomainTransaction, bool) {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	if mempoolTransaction, ok := mp.transactionsPool.allTransactions[*transactionID]; ok {
		return mempoolTransaction.Transaction(), true
	}
	return nil, false
}

func (mp *mempool) AllTransactions() []*externalapi.DomainTransaction {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	allTransactions := make([]*externalapi.DomainTransaction, 0, len(mp.transactionsPool.allTransactions))
	for _, mempoolTransaction := range mp.transactionsPool.allTransactions {
		allTransactions = append(allTransactions, mempoolTransaction.Transaction())
	}
	return allTransactions
}

func (mp *mempool) TransactionCount() int {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	return len(mp.transactionsPool.allTransactions)
}

func (mp *mempool) HandleNewBlockTransactions(transactions []*externalapi.DomainTransaction) (
	acceptedOrphans []*externalapi.DomainTransaction, err error) {

	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	acceptedOrphans = []*externalapi.DomainTransaction{}
	for _, transaction := range transactions {
		transactionID := consensushashing.TransactionID(transaction)
		err := mp.removeTransaction(transactionID, false)
		if err != nil {
			return nil, err
		}

		err = mp.removeDoubleSpends(transaction)
		if err != nil {
			return nil, err
		}

		err = mp.orphansPool.removeOrphan(transactionID, false)
		if err != nil {
			return nil, err
		}

		acceptedOrphansFromThisTransaction, err := mp.orphansPool.processOrphansAfterAcceptedTransaction(transaction)
		if err != nil {
			return nil, err
		}

		acceptedOrphans = append(acceptedOrphans, acceptedOrphansFromThisTransaction...)
	}
	err = mp.orphansPool.expireOrphanTransactions()
	if err != nil {
		return nil, err
	}
	err = mp.transactionsPool.expireOldTransactions()
	if err != nil {
		return nil, err
	}

	return acceptedOrphans, nil
}

func (mp *mempool) BlockCandidateTransactions() []*externalapi.DomainTransaction {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	return mp.transactionsPool.allReadyTransactions()
}

func (mp *mempool) RevalidateHighPriorityTransactions() (validTransactions []*externalapi.DomainTransaction, err error) {
	validTransactions = []*externalapi.DomainTransaction{}

	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	for _, transaction := range mp.transactionsPool.highPriorityTransactions {
		clearInputs(transaction)
		_, missingParents, err := mp.fillInputsAndGetMissingParents(transaction.Transaction())
		if err != nil {
			return nil, err
		}
		if len(missingParents) > 0 {
			err := mp.removeTransaction(transaction.TransactionID(), true)
			if err != nil {
				return nil, err
			}
			continue
		}
		validTransactions = append(validTransactions, transaction.Transaction())
	}

	return validTransactions, nil
}

func clearInputs(transaction *model.MempoolTransaction) {
	for _, input := range transaction.Transaction().Inputs {
		input.UTXOEntry = nil
	}
}

// this function MUST be called with the mempool mutex locked for reads
func (mp *mempool) fillInputsAndGetMissingParents(transaction *externalapi.DomainTransaction) (
	parents model.OutpointToTransaction, missingOutpoints []*externalapi.DomainOutpoint, err error) {

	parentsInPool := mp.transactionsPool.getParentTransactionsInPool(transaction)

	fillInputs(transaction, parentsInPool)

	err = mp.consensus.ValidateTransactionAndPopulateWithConsensusData(transaction)
	if err != nil {
		errMissingOutpoints := ruleerrors.ErrMissingTxOut{}
		if errors.As(err, &errMissingOutpoints) {
			return parentsInPool, errMissingOutpoints.MissingOutpoints, nil
		}
		if errors.Is(err, ruleerrors.ErrImmatureSpend) {
			return nil, nil, transactionRuleError(
				RejectImmatureSpend, "one of the transaction inputs spends an immature UTXO")
		}
		if errors.As(err, &ruleerrors.RuleError{}) {
			return nil, nil, newRuleError(err)
		}
		return nil, nil, err
	}

	return parentsInPool, nil, nil
}

// this function MUST be called with the mempool mutex locked for writes
func (mp *mempool) removeDoubleSpends(transaction *externalapi.DomainTransaction) error {
	for _, input := range transaction.Inputs {
		if redeemer, ok := mp.mempoolUTXOSet.transactionByPreviousOutpoint[input.PreviousOutpoint]; ok {
			err := mp.removeTransaction(redeemer.TransactionID(), true)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func fillInputs(transaction *externalapi.DomainTransaction, parentsInPool model.OutpointToTransaction) {
	for _, input := range transaction.Inputs {
		parent, ok := parentsInPool[input.PreviousOutpoint]
		if !ok {
			continue
		}
		relevantOutput := parent.Transaction().Outputs[input.PreviousOutpoint.Index]
		input.UTXOEntry = utxo.NewUTXOEntry(relevantOutput.Value, relevantOutput.ScriptPublicKey,
			transactionhelper.IsCoinBase(transaction), model.UnacceptedDAAScore)
	}
}
