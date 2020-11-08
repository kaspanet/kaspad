package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_Block) toAppMessage() (appmessage.Message, error) {
	return x.Block.toAppMessage()
}

func (x *KaspadMessage_Block) fromAppMessage(msgBlock *appmessage.MsgBlock) error {
	x.Block = new(BlockMessage)
	return x.Block.fromAppMessage(msgBlock)
}

func (x *BlockMessage) toAppMessage() (appmessage.Message, error) {
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

	parentHashes, err := protoHashesToDomain(protoBlockHeader.ParentHashes)
	if err != nil {
		return nil, err
	}

	hashMerkleRoot, err := protoBlockHeader.HashMerkleRoot.toDomain()
	if err != nil {
		return nil, err
	}

	acceptedIDMerkleRoot, err := protoBlockHeader.AcceptedIDMerkleRoot.toDomain()
	if err != nil {
		return nil, err
	}

	utxoCommitment, err := protoBlockHeader.UtxoCommitment.toDomain()
	if err != nil {
		return nil, err
	}

	header := appmessage.BlockHeader{
		Version:              protoBlockHeader.Version,
		ParentHashes:         parentHashes,
		HashMerkleRoot:       hashMerkleRoot,
		AcceptedIDMerkleRoot: acceptedIDMerkleRoot,
		UTXOCommitment:       utxoCommitment,
		Timestamp:            mstime.UnixMilliseconds(protoBlockHeader.Timestamp),
		Bits:                 protoBlockHeader.Bits,
		Nonce:                protoBlockHeader.Nonce,
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
		Header:       header,
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

	header := msgBlock.Header
	protoHeader := &BlockHeader{
		Version:              header.Version,
		ParentHashes:         domainHashesToProto(header.ParentHashes),
		HashMerkleRoot:       domainHashToProto(header.HashMerkleRoot),
		AcceptedIDMerkleRoot: domainHashToProto(header.AcceptedIDMerkleRoot),
		UtxoCommitment:       domainHashToProto(header.UTXOCommitment),
		Timestamp:            header.Timestamp.UnixMilliseconds(),
		Bits:                 header.Bits,
		Nonce:                header.Nonce,
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
