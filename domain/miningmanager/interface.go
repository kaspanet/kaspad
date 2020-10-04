package miningmanager

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util"
)

// Mempool maintains a set of known transactions that
// have no yet been added to any block
type Mempool interface {
	HandleNewBlock(block *appmessage.MsgBlock)
	Transactions() []*util.Tx
	ValidateAndInsertTransaction(transaction *appmessage.MsgTx) error
}

// BlockTemplateBuilder builds block templates for miners to consume
type BlockTemplateBuilder interface {
	GetBlockTemplate() *appmessage.MsgBlock
}
