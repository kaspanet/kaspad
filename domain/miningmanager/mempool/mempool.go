package mempool

import (
	"fmt"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool/model"
	"github.com/pkg/errors"
)

type mempool struct {
	config    *config
	consensus externalapi.Consensus

	mempoolUTXOSet   *mempoolUTXOSet
	transactionsPool *transactionsPool
	orphansPool      *orphansPool
}

func newMempool(consensus externalapi.Consensus, dagParams *dagconfig.Params) *mempool {
	mp := &mempool{
		config:    defaultConfig(dagParams),
		consensus: consensus,
	}

	mp.mempoolUTXOSet = newMempoolUTXOSet(mp)
	mp.transactionsPool = newTransactionsPool(mp)
	mp.orphansPool = newOrphansPool(mp)

	return mp
}

func (mp *mempool) ValidateAndInsertTransaction(transaction *externalapi.DomainTransaction, isHighPriority bool, allowOrphan bool) (
	acceptedTransactions []*externalapi.DomainTransaction, err error) {

	err = mp.validateTransactionInContext(transaction)
	if err != nil {
		return nil, err
	}

	parentsInPool, missingOutpoints, err := mp.fillInputsAndGetMissingParents(transaction)
	if err != nil {
		return nil, err
	}

	if len(missingOutpoints) > 0 {
		if !allowOrphan {
			str := fmt.Sprintf("Transaction %s is an orphan, where allowOrphans = false",
				consensushashing.TransactionID(transaction))
			return nil, transactionRuleError(RejectBadOrphan, str)
		}

		return nil, mp.orphansPool.maybeAddOrphan(transaction, isHighPriority)
	}

	err = mp.validateTransactionInContext(transaction)
	if err != nil {
		return nil, err
	}

	mempoolTransaction, err := mp.transactionsPool.addTransaction(transaction, parentsInPool, isHighPriority)
	if err != nil {
		return nil, err
	}

	acceptedOrphans, err := mp.orphansPool.processOrphansAfterAcceptedTransaction(mempoolTransaction.Transaction())
	if err != nil {
		return nil, err
	}

	acceptedTransactions = make([]*externalapi.DomainTransaction, 0, len(acceptedOrphans)+1)
	acceptedTransactions = append(acceptedTransactions, transaction)
	for _, acceptedOrphan := range acceptedOrphans {
		acceptedTransactions = append(acceptedTransactions, acceptedOrphan)
	}

	mp.transactionsPool.limitTransactionCount()

	return acceptedTransactions, nil
}

func (mp *mempool) HandleNewBlockTransactions(transactions []*externalapi.DomainTransaction) (
	acceptedOrphans []*externalapi.DomainTransaction, err error) {

	acceptedOrphans = []*externalapi.DomainTransaction{}
	for _, transaction := range transactions {
		transactionID := consensushashing.TransactionID(transaction)
		err := mp.RemoveTransaction(transactionID, false)
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

func (mp *mempool) RemoveTransaction(transactionID *externalapi.DomainTransactionID, removeRedeemers bool) error {
	if _, ok := mp.orphansPool.allOrphans[*transactionID]; ok {
		return mp.orphansPool.removeOrphan(transactionID, true)
	}

	mempoolTransaction, ok := mp.transactionsPool.allTransactions[*transactionID]
	if !ok {
		return nil
	}

	transactionsToRemove := []*model.MempoolTransaction{mempoolTransaction}
	if removeRedeemers {
		redeemers := mp.transactionsPool.getRedeemers(mempoolTransaction)
		transactionsToRemove = append(transactionsToRemove, redeemers...)
	}

	for _, transactionToRemove := range transactionsToRemove {
		err := mp.removeTransaction(transactionToRemove, removeRedeemers)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mp *mempool) removeTransaction(mempoolTransaction *model.MempoolTransaction, removeRedeemers bool) error {
	mp.mempoolUTXOSet.removeTransaction(mempoolTransaction)

	err := mp.transactionsPool.removeTransaction(mempoolTransaction)
	if err != nil {
		return err
	}

	err = mp.orphansPool.updateOrphansAfterTransactionRemoved(mempoolTransaction, removeRedeemers)
	if err != nil {
		return err
	}

	return nil
}

func (mp *mempool) BlockCandidateTransactions() []*externalapi.DomainTransaction {
	return mp.transactionsPool.allReadyTransactions()
}

func (mp *mempool) RevalidateHighPriorityTransactions() (validTransactions []*externalapi.DomainTransaction, err error) {
	panic("mempool.RevalidateHighPriorityTransactions not implemented") // TODO (Mike)
}

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

func (mp *mempool) removeDoubleSpends(transaction *externalapi.DomainTransaction) error {
	for _, input := range transaction.Inputs {
		if redeemer, ok := mp.mempoolUTXOSet.transactionByPreviousOutpoint[input.PreviousOutpoint]; ok {
			err := mp.RemoveTransaction(redeemer.TransactionID(), true)
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
