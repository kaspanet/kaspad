package model

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/util"
)

// Mempool maintains a set of known transactions that
// are intended to be mined into new blocks
type Mempool interface {
	HandleNewBlockTransactions(txs []*externalapi.DomainTransaction) ([]*externalapi.DomainTransaction, error)
	BlockCandidateTransactions() []*externalapi.DomainTransaction
	ValidateAndInsertTransaction(transaction *externalapi.DomainTransaction, isHighPriority bool, allowOrphan bool) (
		acceptedTransactions []*externalapi.DomainTransaction, err error)
	RemoveTransactions(txs []*externalapi.DomainTransaction, removeRedeemers bool) error
	GetTransaction(transactionID *externalapi.DomainTransactionID) (*externalapi.DomainTransaction, bool)
	GetTransactionsByAddresses() (
		sending map[util.Address]*externalapi.DomainTransaction,
		receiving map[util.Address]*externalapi.DomainTransaction,
		err error)
	AllTransactions() []*externalapi.DomainTransaction
	GetOrphanTransaction(transactionID *externalapi.DomainTransactionID) (*externalapi.DomainTransaction, bool)
	GetOrphanTransactionsByAddresses() (
		sending map[util.Address]*externalapi.DomainTransaction,
		receiving map[util.Address]*externalapi.DomainTransaction,
		err error)
	AllOrphanTransactions() []*externalapi.DomainTransaction
	TransactionCount() int
	RevalidateHighPriorityTransactions() (validTransactions []*externalapi.DomainTransaction, err error)
	IsTransactionOutputDust(output *externalapi.DomainTransactionOutput) bool
}
