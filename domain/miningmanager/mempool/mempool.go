package mempool

import (
	"sync"

	"github.com/zoomy-network/zoomyd/domain/consensus/utils/consensushashing"
	"github.com/zoomy-network/zoomyd/domain/consensus/utils/constants"

	"github.com/zoomy-network/zoomyd/domain/consensusreference"

	"github.com/zoomy-network/zoomyd/domain/consensus/model/externalapi"
	miningmanagermodel "github.com/zoomy-network/zoomyd/domain/miningmanager/model"
)

type mempool struct {
	mtx sync.RWMutex

	config             *Config
	consensusReference consensusreference.ConsensusReference

	mempoolUTXOSet   *mempoolUTXOSet
	transactionsPool *transactionsPool
	orphansPool      *orphansPool
}

// New constructs a new mempool
func New(config *Config, consensusReference consensusreference.ConsensusReference) miningmanagermodel.Mempool {
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

func (mp *mempool) GetTransaction(transactionID *externalapi.DomainTransactionID,
	includeTransactionPool bool,
	includeOrphanPool bool) (
	transaction *externalapi.DomainTransaction,
	isOrphan bool,
	found bool) {

	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	var transactionfound bool
	isOrphan = false

	if includeTransactionPool {
		transaction, transactionfound = mp.transactionsPool.getTransaction(transactionID, true)
		isOrphan = false
	}
	if !transactionfound && includeOrphanPool {
		transaction, transactionfound = mp.orphansPool.getOrphanTransaction(transactionID)
		isOrphan = true
	}

	return transaction, isOrphan, transactionfound
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
		sendingInTransactionPool, receivingInTransactionPool, err = mp.transactionsPool.getTransactionsByAddresses()
		if err != nil {
			return nil, nil, nil, nil, err
		}
	}

	if includeOrphanPool {
		sendingInTransactionPool, receivingInOrphanPool, err = mp.orphansPool.getOrphanTransactionsByAddresses()
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
		transactionPoolTransactions = mp.transactionsPool.getAllTransactions()
	}

	if includeOrphanPool {
		orphanPoolTransactions = mp.orphansPool.getAllOrphanTransactions()
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

	return mp.handleNewBlockTransactions(transactions)
}

func (mp *mempool) BlockCandidateTransactions() []*externalapi.DomainTransaction {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()

	readyTxs := mp.transactionsPool.allReadyTransactions()
	var candidateTxs []*externalapi.DomainTransaction
	var spamTx *externalapi.DomainTransaction
	var spamTxNewestUTXODaaScore uint64
	for _, tx := range readyTxs {
		if len(tx.Outputs) > len(tx.Inputs) {
			hasCoinbaseInput := false
			for _, input := range tx.Inputs {
				if input.UTXOEntry.IsCoinbase() {
					hasCoinbaseInput = true
					break
				}
			}

			numExtraOuts := len(tx.Outputs) - len(tx.Inputs)
			if !hasCoinbaseInput && numExtraOuts > 2 && tx.Fee < uint64(numExtraOuts)*constants.SompiPerKaspa {
				log.Debugf("Filtered spam tx %s", consensushashing.TransactionID(tx))
				continue
			}

			if hasCoinbaseInput || tx.Fee > uint64(numExtraOuts)*constants.SompiPerKaspa {
				candidateTxs = append(candidateTxs, tx)
			} else {
				txNewestUTXODaaScore := tx.Inputs[0].UTXOEntry.BlockDAAScore()
				for _, input := range tx.Inputs {
					if input.UTXOEntry.BlockDAAScore() > txNewestUTXODaaScore {
						txNewestUTXODaaScore = input.UTXOEntry.BlockDAAScore()
					}
				}

				if spamTx != nil {
					if txNewestUTXODaaScore < spamTxNewestUTXODaaScore {
						spamTx = tx
						spamTxNewestUTXODaaScore = txNewestUTXODaaScore
					}
				} else {
					spamTx = tx
					spamTxNewestUTXODaaScore = txNewestUTXODaaScore
				}
			}
		} else {
			candidateTxs = append(candidateTxs, tx)
		}
	}

	if spamTx != nil {
		log.Debugf("Adding spam tx candidate %s", consensushashing.TransactionID(spamTx))
		candidateTxs = append(candidateTxs, spamTx)
	}

	return candidateTxs
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
