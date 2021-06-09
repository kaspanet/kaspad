package mempool

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
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

func (mp *mempool) ValidateAndInsertTransaction(transaction *externalapi.DomainTransaction, isHighPriority bool, allowOrphans bool) (
	acceptedTransactions []*externalapi.DomainTransaction, err error) {

	panic("mempool.ValidateAndInsertTransaction not implemented") // TODO (Mike)
}

func (mp *mempool) HandleNewBlockTransactions(transactions []*externalapi.DomainTransaction) (
	acceptedOrphans []*externalapi.DomainTransaction, err error) {

	panic("mempool.HandleNewBlockTransactions not implemented") // TODO (Mike)
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
