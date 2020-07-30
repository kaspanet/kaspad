package protowire

import (
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_RequestRelayBlocks) toWireMessage() (wire.Message, error) {
	if len(x.RequestRelayBlocks.Hashes) > wire.MsgGetRelayBlocksHashes {
		return nil, errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(x.RequestRelayBlocks.Hashes), wire.MsgGetRelayBlocksHashes)
	}
	hashes, err := protoHashesToWire(x.RequestRelayBlocks.Hashes)
	if err != nil {
		return nil, err
	}
	return &wire.MsgRequestRelayBlocks{Hashes: hashes}, nil
}

func (x *KaspadMessage_RequestRelayBlocks) fromWireMessage(msgGetRelayBlocks *wire.MsgRequestRelayBlocks) error {
	if len(msgGetRelayBlocks.Hashes) > wire.MsgGetRelayBlocksHashes {
		return errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(msgGetRelayBlocks.Hashes), wire.MsgGetRelayBlocksHashes)
	}

	x.RequestRelayBlocks = &RequestRelayBlocksMessage{
		Hashes: wireHashesToProto(msgGetRelayBlocks.Hashes),
	}
	return nil
}
