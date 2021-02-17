package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_Block) toAppMessage() (appmessage.Message, error) {
	return x.Block.toAppMessage()
}

func (x *KaspadMessage_Block) fromAppMessage(msgBlock *appmessage.MsgBlock) error {
	x.Block = new(BlockMessage)
	return x.Block.fromAppMessage(msgBlock)
}

func (x *BlockMessage) toAppMessage() (*appmessage.MsgBlock, error) {
	if len(x.Transactions) > appmessage.MaxTxPerBlock {
		return nil, errors.Errorf("too many transactions to fit into a block "+
			"[count %d, max %d]", len(x.Transactions), appmessage.MaxTxPerBlock)
	}

	protoBlockHeader := x.Header
	if protoBlockHeader == nil {
		return nil, errors.New("block header field cannot be nil")
	}

	if len(protoBlockHeader.ParentHashes) > appmessage.MaxBlockParents {
		return nil, errors.Errorf("block header has %d parents, but the maximum allowed amount "+
			"is %d", len(protoBlockHeader.ParentHashes), appmessage.MaxBlockParents)
	}

	header, err := x.Header.toAppMessage()
	if err != nil {
		return nil, err
	}

	transactions := make([]*appmessage.MsgTx, len(x.Transactions))
	for i, protoTx := range x.Transactions {
		msgTx, err := protoTx.toAppMessage()
		if err != nil {
			return nil, err
		}
		transactions[i] = msgTx.(*appmessage.MsgTx)
	}

	return &appmessage.MsgBlock{
		Header:       *header,
		Transactions: transactions,
	}, nil
}

func (x *BlockMessage) fromAppMessage(msgBlock *appmessage.MsgBlock) error {
	if len(msgBlock.Transactions) > appmessage.MaxTxPerBlock {
		return errors.Errorf("too many transactions to fit into a block "+
			"[count %d, max %d]", len(msgBlock.Transactions), appmessage.MaxTxPerBlock)
	}

	if len(msgBlock.Header.ParentHashes) > appmessage.MaxBlockParents {
		return errors.Errorf("block header has %d parents, but the maximum allowed amount "+
			"is %d", len(msgBlock.Header.ParentHashes), appmessage.MaxBlockParents)
	}

	protoHeader := new(BlockHeaderMessage)
	err := protoHeader.fromAppMessage(&msgBlock.Header)
	if err != nil {
		return err
	}

	protoTransactions := make([]*TransactionMessage, len(msgBlock.Transactions))
	for i, tx := range msgBlock.Transactions {
		protoTx := new(TransactionMessage)
		protoTx.fromAppMessage(tx)
		protoTransactions[i] = protoTx
	}
	*x = BlockMessage{
		Header:       protoHeader,
		Transactions: protoTransactions,
	}
	return nil
}
