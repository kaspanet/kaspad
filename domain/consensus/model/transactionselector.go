package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// TransactionSelector is a function for selecting transaction from
// some ReadOnlyUTXOSet
type TransactionSelector func(readOnlyUTXOSet ReadOnlyUTXOSet) []*externalapi.DomainTransaction
