package model

import "github.com/kaspanet/kaspad/util"

type TransactionSelector func(readOnlyUTXOSet ReadOnlyUTXOSet) []*util.Tx
