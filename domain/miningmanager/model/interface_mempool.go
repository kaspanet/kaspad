package model

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// Mempool maintains a set of known transactions that
// are intended to be mined into new blocks
type Mempool interface {
	HandleNewBlockTransactions(txs []*externalapi.DomainTransaction) ([]*externalapi.DomainTransaction, error)
	BlockCandidateTransactions() []*externalapi.DomainTransaction
	ValidateAndInsertTransaction(transaction *externalapi.DomainTransaction, isHighPriority bool, allowOrphan bool) (
		acceptedTransactions []*externalapi.DomainTransaction, err error)
	RemoveTransactions(txs []*externalapi.DomainTransaction, removeRedeemers bool) error
	GetTransaction(
		transactionID *externalapi.DomainTransactionID,
		includeTransactionPool bool,
		includeOrphanPool bool,
	) (
		transactionPoolTransaction *externalapi.DomainTransaction,
		orphanPoolTransaction *externalapi.DomainTransaction,
		found bool)
	GetTransactionsByAddresses(
		includeTransactionPool bool,
		includeOrphanPool bool) (
		sendingInTransactionPool map[string]*externalapi.DomainTransaction,
		receivingInTransactionPool map[string]*externalapi.DomainTransaction,
		sendingInOrphanPool map[string]*externalapi.DomainTransaction,
		receivingInOrphanPool map[string]*externalapi.DomainTransaction,
		err error)
	AllTransactions(
		includeTransactionPool bool,
		includeOrphanPool bool,
	) (
		transactionPoolTransactions []*externalapi.DomainTransaction,
		orphanPoolTransactions []*externalapi.DomainTransaction)
	TransactionCount(
		includeTransactionPool bool,
		includeOrphanPool bool) int
	RevalidateHighPriorityTransactions() (validTransactions []*externalapi.DomainTransaction, err error)
	IsTransactionOutputDust(output *externalapi.DomainTransactionOutput) bool
}
