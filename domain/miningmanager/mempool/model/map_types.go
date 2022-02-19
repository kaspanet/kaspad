package model

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// IDToTransactionMap maps transactionID to a MempoolTransaction
type IDToTransactionMap map[externalapi.DomainTransactionID]*MempoolTransaction

// IDToTransactionsSliceMap maps transactionID to a slice MempoolTransaction
type IDToTransactionsSliceMap map[externalapi.DomainTransactionID][]*MempoolTransaction

// OutpointToUTXOEntryMap maps an outpoint to a UTXOEntry
type OutpointToUTXOEntryMap map[externalapi.DomainOutpoint]externalapi.UTXOEntry

// OutpointToTransactionMap maps an outpoint to a MempoolTransaction
type OutpointToTransactionMap map[externalapi.DomainOutpoint]*MempoolTransaction
