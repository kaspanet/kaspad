package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"
	"math"
)

func (x *BlockHeaderMessage) toAppMessage() (*appmessage.MsgBlockHeader, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "BlockHeaderMessage is nil")
	}
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
	acceptedIDMerkleRoot, err := x.AcceptedIdMerkleRoot.toDomain()
	if err != nil {
		return nil, err
	}
	utxoCommitment, err := x.UtxoCommitment.toDomain()
	if err != nil {
		return nil, err
	}
	if x.Version > math.MaxUint16 {
		return nil, errors.Errorf("Invalid block header version - bigger then uint16")
	}
	return &appmessage.MsgBlockHeader{
		Version:              uint16(x.Version),
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
		Version:              uint32(msgBlockHeader.Version),
		ParentHashes:         domainHashesToProto(msgBlockHeader.ParentHashes),
		HashMerkleRoot:       domainHashToProto(msgBlockHeader.HashMerkleRoot),
		AcceptedIdMerkleRoot: domainHashToProto(msgBlockHeader.AcceptedIDMerkleRoot),
		UtxoCommitment:       domainHashToProto(msgBlockHeader.UTXOCommitment),
		Timestamp:            msgBlockHeader.Timestamp.UnixMilliseconds(),
		Bits:                 msgBlockHeader.Bits,
		Nonce:                msgBlockHeader.Nonce,
	}
	return nil
}
