package mempool

import (
	"sync"

	"github.com/kaspanet/kaspad/domain/consensusreference"
	"github.com/kaspanet/kaspad/domain/dagconfig"

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

func (mp *mempool) GetTransaction(transactionID *externalapi.DomainTransactionID,
	includeTransactionPool bool,
	includeOrphanPool bool) (
	transactionPoolTransaction *externalapi.DomainTransaction,
	orphanPoolTransaction *externalapi.DomainTransaction,
	found bool) {

	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	var transactionfound bool
	var orphanTransactionFound bool

	if includeTransactionPool {
		transactionPoolTransaction, transactionfound = mp.transactionsPool.getTransaction(transactionID, true)
	}
	if includeOrphanPool {
		orphanPoolTransaction, orphanTransactionFound = mp.orphansPool.getOrphanTransaction(transactionID, true)
	}

	return transactionPoolTransaction, orphanPoolTransaction, transactionfound || orphanTransactionFound
}

func (mp *mempool) GetTransactionsByAddresses(includeTransactionPool bool, includeOrphanPool bool) (
	sendingInTransactionPool map[string]*externalapi.DomainTransaction,
	receivingInTransactionPool map[string]*externalapi.DomainTransaction,
	sendingInOrphanPool map[string]*externalapi.DomainTransaction,
	receivingInOrphanPool map[string]*externalapi.DomainTransaction,
	err error) {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	if includeTransactionPool {
		sendingInTransactionPool, receivingInTransactionPool, err = mp.transactionsPool.getTransactionsByAddresses(true)
		if err != nil {
			return nil, nil, nil, nil, err
		}
	}

	if includeOrphanPool {
		sendingInTransactionPool, receivingInOrphanPool, err = mp.orphansPool.getOrphanTransactionsByAddresses(true)
		if err != nil {
			return nil, nil, nil, nil, err
		}
	}

	return sendingInTransactionPool, receivingInTransactionPool, sendingInTransactionPool, receivingInOrphanPool, nil
}

func (mp *mempool) AllTransactions(includeTransactionPool bool, includeOrphanPool bool) (
	transactionPoolTransactions []*externalapi.DomainTransaction,
	orphanPoolTransactions []*externalapi.DomainTransaction) {

	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	if includeTransactionPool {
		transactionPoolTransactions = mp.transactionsPool.getAllTransactions(true)
	}

	if includeOrphanPool {
		orphanPoolTransactions = mp.orphansPool.getAllOrphanTransactions(true)
	}

	return transactionPoolTransactions, orphanPoolTransactions
}

func (mp *mempool) TransactionCount(includeTransactionPool bool, includeOrphanPool bool) int {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	transactionCount := 0

	if includeOrphanPool {
		transactionCount += mp.orphansPool.orphanTransactionCount()
	}
	if includeTransactionPool {
		transactionCount += mp.transactionsPool.transactionCount()
	}

	return transactionCount
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
