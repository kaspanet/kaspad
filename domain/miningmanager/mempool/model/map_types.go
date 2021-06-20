package model

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// IDsToTransactions maps transactionID to a MempoolTransaction
type IDsToTransactions map[externalapi.DomainTransactionID]*MempoolTransaction

// OutpointsToUTXOEntries maps an outpoint to a UTXOEntry
type OutpointsToUTXOEntries map[externalapi.DomainOutpoint]externalapi.UTXOEntry

// OutpointsToTransactions maps an outpoint to a MempoolTransaction
type OutpointsToTransactions map[externalapi.DomainOutpoint]*MempoolTransaction
