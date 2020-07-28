package protowire

import (
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_Block) toWireMessage() (*wire.MsgBlock, error) {
	return x.Block.toWireMessage()
}

func (x *KaspadMessage_Block) fromWireMessage(msgBlock *wire.MsgBlock) error {
	x.Block = new(BlockMessage)
	return x.Block.fromWireMessage(msgBlock)
}

func (x *BlockMessage) toWireMessage() (*wire.MsgBlock, error) {
	if len(x.Transactions) > wire.MaxTxPerBlock {
		return nil, errors.Errorf("too many transactions to fit into a block "+
			"[count %d, max %d]", len(x.Transactions), wire.MaxTxPerBlock)
	}

	protoBlockHeader := x.Header
	parentHashes, err := protoHashesToWire(protoBlockHeader.ParentHashes)
	if err != nil {
		return nil, err
	}

	hashMerkleRoot, err := protoBlockHeader.HashMerkleRoot.toWire()
	if err != nil {
		return nil, err
	}

	acceptedIDMerkleRoot, err := protoBlockHeader.AcceptedIDMerkleRoot.toWire()
	if err != nil {
		return nil, err
	}

	utxoCommitment, err := protoBlockHeader.UtxoCommitment.toWire()
	if err != nil {
		return nil, err
	}

	header := wire.BlockHeader{
		Version:              protoBlockHeader.Version,
		ParentHashes:         parentHashes,
		HashMerkleRoot:       hashMerkleRoot,
		AcceptedIDMerkleRoot: acceptedIDMerkleRoot,
		UTXOCommitment:       utxoCommitment,
		Timestamp:            mstime.UnixMilliseconds(protoBlockHeader.Timestamp),
		Bits:                 protoBlockHeader.Bits,
		Nonce:                protoBlockHeader.Nonce,
	}

	transactions := make([]*wire.MsgTx, len(x.Transactions))
	for i, protoTx := range x.Transactions {
		transactions[i], err = protoTx.toWireMessage()
		if err != nil {
			return nil, err
		}
	}

	return &wire.MsgBlock{
		Header:       header,
		Transactions: transactions,
	}, nil
}

func (x *BlockMessage) fromWireMessage(msgBlock *wire.MsgBlock) error {
	if len(msgBlock.Transactions) > wire.MaxTxPerBlock {
		return errors.Errorf("too many transactions to fit into a block "+
			"[count %d, max %d]", len(msgBlock.Transactions), wire.MaxTxPerBlock)
	}

	header := msgBlock.Header
	protoHeader := &BlockHeader{
		Version:              header.Version,
		ParentHashes:         wireHashesToProto(header.ParentHashes),
		HashMerkleRoot:       wireHashToProto(header.HashMerkleRoot),
		AcceptedIDMerkleRoot: wireHashToProto(header.AcceptedIDMerkleRoot),
		UtxoCommitment:       wireHashToProto(header.UTXOCommitment),
		Timestamp:            header.Timestamp.UnixMilliseconds(),
		Bits:                 header.Bits,
		Nonce:                header.Nonce,
	}
	protoTransactions := make([]*TransactionMessage, len(msgBlock.Transactions))
	for i, tx := range msgBlock.Transactions {
		protoTransactions[i].fromWireMessage(tx)
	}
	*x = BlockMessage{
		Header:       protoHeader,
		Transactions: protoTransactions,
	}
	return nil
}
