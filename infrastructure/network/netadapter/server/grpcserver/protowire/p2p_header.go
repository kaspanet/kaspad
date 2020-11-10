package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_BlockHeader) toAppMessage() (appmessage.Message, error) {
	return x.BlockHeader.toAppMessage()
}

func (x *KaspadMessage_BlockHeader) fromAppMessage(msgBlockHeader *appmessage.MsgBlockHeader) error {
	x.BlockHeader = new(BlockHeaderMessage)
	return x.BlockHeader.fromAppMessage(msgBlockHeader)
}

func (x *BlockHeaderMessage) toAppMessage() (*appmessage.MsgBlockHeader, error) {
	if len(x.ParentHashes) > appmessage.MaxBlockParents {
		return nil, errors.Errorf("block header has %d parents, but the maximum allowed amount "+
			"is %d", len(x.ParentHashes), appmessage.MaxBlockParents)
	}

	parentHashes, err := protoHashesToDomain(x.ParentHashes)
	if err != nil {
		return nil, err
	}

	hashMerkleRoot, err := x.HashMerkleRoot.toDomain()
	if err != nil {
		return nil, err
	}

	acceptedIDMerkleRoot, err := x.AcceptedIDMerkleRoot.toDomain()
	if err != nil {
		return nil, err
	}

	utxoCommitment, err := x.UtxoCommitment.toDomain()
	if err != nil {
		return nil, err
	}

	return &appmessage.MsgBlockHeader{
		Version:              x.Version,
		ParentHashes:         parentHashes,
		HashMerkleRoot:       hashMerkleRoot,
		AcceptedIDMerkleRoot: acceptedIDMerkleRoot,
		UTXOCommitment:       utxoCommitment,
		Timestamp:            mstime.UnixMilliseconds(x.Timestamp),
		Bits:                 x.Bits,
		Nonce:                x.Nonce,
	}, nil
}

func (x *BlockHeaderMessage) fromAppMessage(msgBlockHeader *appmessage.MsgBlockHeader) error {
	if len(msgBlockHeader.ParentHashes) > appmessage.MaxBlockParents {
		return errors.Errorf("block header has %d parents, but the maximum allowed amount "+
			"is %d", len(msgBlockHeader.ParentHashes), appmessage.MaxBlockParents)
	}

	*x = BlockHeaderMessage{
		Version:              msgBlockHeader.Version,
		ParentHashes:         domainHashesToProto(msgBlockHeader.ParentHashes),
		HashMerkleRoot:       domainHashToProto(msgBlockHeader.HashMerkleRoot),
		AcceptedIDMerkleRoot: domainHashToProto(msgBlockHeader.AcceptedIDMerkleRoot),
		UtxoCommitment:       domainHashToProto(msgBlockHeader.UTXOCommitment),
		Timestamp:            msgBlockHeader.Timestamp.UnixMilliseconds(),
		Bits:                 msgBlockHeader.Bits,
		Nonce:                msgBlockHeader.Nonce,
	}
	return nil
}
