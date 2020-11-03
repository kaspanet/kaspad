package domainconverters

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/util/mstime"
)

func DomainBlockToMsgBlock(domainBlock externalapi.DomainBlock) *appmessage.MsgBlock {
	msgTxs := make([]*appmessage.MsgTx, 0, len(domainBlock.Transactions))
	for _, domainTransaction := range domainBlock.Transactions {
		msgTxs = append(msgTxs, DomainTransactionToMsgTx(domainTransaction))
	}
	return &appmessage.MsgBlock{
		Header:       *domainBlockHeaderToBlockHeader(domainBlock.Header),
		Transactions: msgTxs,
	}
}

func domainBlockHeaderToBlockHeader(domainBlockHeader *externalapi.DomainBlockHeader) *appmessage.BlockHeader {
	return &appmessage.BlockHeader{
		Version:              domainBlockHeader.Version,
		ParentHashes:         domainBlockHeader.ParentHashes,
		HashMerkleRoot:       &domainBlockHeader.HashMerkleRoot,
		AcceptedIDMerkleRoot: &domainBlockHeader.AcceptedIDMerkleRoot,
		UTXOCommitment:       &domainBlockHeader.UTXOCommitment,
		Timestamp:            mstime.UnixMilliseconds(domainBlockHeader.TimeInMilliseconds),
		Bits:                 domainBlockHeader.Bits,
		Nonce:                domainBlockHeader.Nonce,
	}
}

func MsgBlockToDomainBlock(msgBlock *appmessage.MsgBlock) *externalapi.DomainBlock {
	transactions := make([]*externalapi.DomainTransaction, 0, len(msgBlock.Transactions))
	for _, msgTx := range msgBlock.Transactions {
		transactions = append(transactions, MsgTxToDomainTransaction(msgTx))
	}

	return &externalapi.DomainBlock{
		Header:       blockHeaderToDomainBlockHeader(&msgBlock.Header),
		Transactions: transactions,
	}
}

func blockHeaderToDomainBlockHeader(blockHeader *appmessage.BlockHeader) *externalapi.DomainBlockHeader {
	return &externalapi.DomainBlockHeader{
		Version:              blockHeader.Version,
		ParentHashes:         blockHeader.ParentHashes,
		HashMerkleRoot:       *blockHeader.HashMerkleRoot,
		AcceptedIDMerkleRoot: *blockHeader.AcceptedIDMerkleRoot,
		UTXOCommitment:       *blockHeader.UTXOCommitment,
		TimeInMilliseconds:   blockHeader.Timestamp.UnixMilliseconds(),
		Bits:                 blockHeader.Bits,
		Nonce:                blockHeader.Nonce,
	}
}
