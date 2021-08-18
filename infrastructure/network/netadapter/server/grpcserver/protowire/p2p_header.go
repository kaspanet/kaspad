package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"
	"math"
	"math/big"
)

func (x *BlockHeader) toAppMessage() (*appmessage.MsgBlockHeader, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "BlockHeaderMessage is nil")
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
	pruningPoint, err := x.PruningPoint.toDomain()
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
		DAAScore:             x.DaaScore,
		BlueWork:             new(big.Int).SetBytes(x.BlueWork),
		PruningPoint:         pruningPoint,
	}, nil
}

func (x *BlockHeader) fromAppMessage(msgBlockHeader *appmessage.MsgBlockHeader) error {
	*x = BlockHeader{
		Version:              uint32(msgBlockHeader.Version),
		ParentHashes:         domainHashesToProto(msgBlockHeader.ParentHashes),
		HashMerkleRoot:       domainHashToProto(msgBlockHeader.HashMerkleRoot),
		AcceptedIdMerkleRoot: domainHashToProto(msgBlockHeader.AcceptedIDMerkleRoot),
		UtxoCommitment:       domainHashToProto(msgBlockHeader.UTXOCommitment),
		Timestamp:            msgBlockHeader.Timestamp.UnixMilliseconds(),
		Bits:                 msgBlockHeader.Bits,
		Nonce:                msgBlockHeader.Nonce,
		DaaScore:             msgBlockHeader.DAAScore,
		BlueWork:             msgBlockHeader.BlueWork.Bytes(),
		PruningPoint:         domainHashToProto(msgBlockHeader.PruningPoint),
	}
	return nil
}
