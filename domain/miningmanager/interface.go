package miningmanager

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util"
)

// Mempool maintains a set of known transactions that
// are intended to be mined into new blocks
type Mempool interface {
	HandleNewBlock(block *appmessage.MsgBlock)
	Transactions() []*util.Tx
	ValidateAndInsertTransaction(transaction *appmessage.MsgTx) error
}

// BlockTemplateBuilder builds block templates for miners to consume
type BlockTemplateBuilder interface {
	GetBlockTemplate(payAddress util.Address, extraData []byte) *appmessage.MsgBlock
}
