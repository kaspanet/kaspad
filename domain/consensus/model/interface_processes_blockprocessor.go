package model

import "github.com/kaspanet/kaspad/app/appmessage"

// BlockProcessor is responsible for processing incoming blocks
// and creating blocks from the current state
type BlockProcessor interface {
	BuildBlock(coinbaseScriptPublicKey []byte, coinbaseExtraData []byte, transactionSelector TransactionSelector) *appmessage.MsgBlock
	ValidateAndInsertBlock(block *appmessage.MsgBlock) error
}
