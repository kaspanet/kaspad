package model

import "github.com/kaspanet/kaspad/util"

// TransactionSelector ...
type TransactionSelector func(readOnlyUTXOSet ReadOnlyUTXOSet) []*util.Tx
