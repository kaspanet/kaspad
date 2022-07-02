package mempool

import (
	"sync"

	"github.com/kaspanet/kaspad/domain/consensusreference"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	miningmanagermodel "github.com/kaspanet/kaspad/domain/miningmanager/model"
)

type mempool struct {
	mtx sync.RWMutex

	config             *Config
	params             *dagconfig.Params
	consensusReference consensusreference.ConsensusReference

	mempoolUTXOSet   *mempoolUTXOSet
	transactionsPool *transactionsPool
	orphansPool      *orphansPool
}

// New constructs a new mempool
func New(config *Config, params *dagconfig.Params, consensusReference consensusreference.ConsensusReference) miningmanagermodel.Mempool {
	mp := &mempool{
		config:             config,
		params:             params,
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

	return mp.validateAndInsertTransaction(transaction, isHighPriority, allowOrphan, true)
}

func (mp *mempool) GetTransaction(transactionID *externalapi.DomainTransactionID) (*externalapi.DomainTransaction, bool) {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	return mp.transactionsPool.getTransaction(transactionID, true)
}

func (mp *mempool) GetTransactionsByAddresses() (
	sending map[util.Address]*externalapi.DomainTransaction,
	receiving map[util.Address]*externalapi.DomainTransaction,
	err error,
) {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	return mp.transactionsPool.getTransactionsByAddresses(true)
}

func (mp *mempool) AllTransactions() []*externalapi.DomainTransaction {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	return mp.transactionsPool.getAllTransactions(true)
}

func (mp *mempool) GetOrphanTransaction(transactionID *externalapi.DomainTransactionID) (*externalapi.DomainTransaction, bool) {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	return mp.orphansPool.getOrphanTransaction(transactionID, true)
}

func (mp *mempool) GetOrphanTransactionsByAddresses() (
	sending map[util.Address]*externalapi.DomainTransaction,
	receiving map[util.Address]*externalapi.DomainTransaction,
	err error,
) {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	return mp.orphansPool.getOrphanTransactionsByAddresses(true)
}

func (mp *mempool) AllOrphanTransactions() []*externalapi.DomainTransaction {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	return mp.orphansPool.getAllOrphanTransactions(true)
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

	return mp.handleNewBlockTransactions(transactions, true)
}

func (mp *mempool) BlockCandidateTransactions() []*externalapi.DomainTransaction {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	return mp.transactionsPool.allReadyTransactions(true)
}

func (mp *mempool) RevalidateHighPriorityTransactions() (validTransactions []*externalapi.DomainTransaction, err error) {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	return mp.revalidateHighPriorityTransactions(true)
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
