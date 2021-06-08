package model

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// IDToTransaction maps transactionID to a MempoolTransaction
type IDToTransaction map[externalapi.DomainTransactionID]*MempoolTransaction

// OutpointToUTXOEntry maps an outpoint to a UTXOEntry
type OutpointToUTXOEntry map[externalapi.DomainOutpoint]externalapi.UTXOEntry

// OutpointToTransaction maps an outpoint to a MempoolTransaction
type OutpointToTransaction map[externalapi.DomainOutpoint]*MempoolTransaction
