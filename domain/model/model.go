package model

import "github.com/kaspanet/kaspad/util"

type ReadOnlyUTXOSet interface {
}

type TransactionSelector func(readOnlyUTXOSet ReadOnlyUTXOSet) []*util.Tx
