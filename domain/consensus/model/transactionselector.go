package model

import "github.com/kaspanet/kaspad/util"

// TransactionSelector is a function for selecting transaction from
// some ReadOnlyUTXOSet
type TransactionSelector func(readOnlyUTXOSet ReadOnlyUTXOSet) []*util.Tx
