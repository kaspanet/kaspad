package mempool

import (
	"sync"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	miningmanagermodel "github.com/kaspanet/kaspad/domain/miningmanager/model"
)

type mempool struct {
	mtx sync.RWMutex

	config             *Config
	consensusReference externalapi.ConsensusReference

	mempoolUTXOSet   *mempoolUTXOSet
	transactionsPool *transactionsPool
	orphansPool      *orphansPool
}

// New constructs a new mempool
func New(config *Config, consensusReference externalapi.ConsensusReference) miningmanagermodel.Mempool {
	mp := &mempool{
		config:             config,
		consensusReference: consensusReference,
	}

	mp.mempoolUTXOSet = newMempoolUTXOSet(mp)
	mp.transactionsPool = newTransactionsPool(mp)
	mp.orphansPool = newOrphansPool(mp)

	return mp
}

func (mp *mempool) ValidateAndInsertTransaction(transaction *externalapi.DomainTransaction, isHighPriority bool, allowOrphan bool) (
	acceptedTransactions []*externalapi.DomainTransaction, err error) {

	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	return mp.validateAndInsertTransaction(transaction, isHighPriority, allowOrphan)
}

func (mp *mempool) GetTransaction(transactionID *externalapi.DomainTransactionID) (*externalapi.DomainTransaction, bool) {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	return mp.transactionsPool.getTransaction(transactionID)
}

func (mp *mempool) AllTransactions() []*externalapi.DomainTransaction {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	return mp.transactionsPool.getAllTransactions()
}

func (mp *mempool) TransactionCount() int {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	return mp.transactionsPool.transactionCount()
}

func (mp *mempool) HandleNewBlockTransactions(transactions []*externalapi.DomainTransaction) (
	acceptedOrphans []*externalapi.DomainTransaction, err error) {

	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	return mp.handleNewBlockTransactions(transactions)
}

func (mp *mempool) BlockCandidateTransactions() []*externalapi.DomainTransaction {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	return mp.transactionsPool.allReadyTransactions()
}

func (mp *mempool) RevalidateHighPriorityTransactions() (validTransactions []*externalapi.DomainTransaction, err error) {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	return mp.revalidateHighPriorityTransactions()
}

func (mp *mempool) RemoveTransactions(transactions []*externalapi.DomainTransaction, removeRedeemers bool) error {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	return mp.removeTransactions(transactions, removeRedeemers)
}

func (mp *mempool) RemoveTransaction(transactionID *externalapi.DomainTransactionID, removeRedeemers bool) error {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	return mp.removeTransaction(transactionID, removeRedeemers)
}
