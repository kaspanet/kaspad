package protowire

import (
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_RequestRelayBlocks) toWireMessage() (wire.Message, error) {
	if len(x.RequestRelayBlocks.Hashes) > wire.MsgRequestRelayBlocksHashes {
		return nil, errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(x.RequestRelayBlocks.Hashes), wire.MsgRequestRelayBlocksHashes)
	}
	hashes, err := protoHashesToWire(x.RequestRelayBlocks.Hashes)
	if err != nil {
		return nil, err
	}
	return &wire.MsgRequestRelayBlocks{Hashes: hashes}, nil
}

func (x *KaspadMessage_RequestRelayBlocks) fromWireMessage(msgGetRelayBlocks *wire.MsgRequestRelayBlocks) error {
	if len(msgGetRelayBlocks.Hashes) > wire.MsgRequestRelayBlocksHashes {
		return errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(msgGetRelayBlocks.Hashes), wire.MsgRequestRelayBlocksHashes)
	}

	x.RequestRelayBlocks = &RequestRelayBlocksMessage{
		Hashes: wireHashesToProto(msgGetRelayBlocks.Hashes),
	}
	return nil
}
