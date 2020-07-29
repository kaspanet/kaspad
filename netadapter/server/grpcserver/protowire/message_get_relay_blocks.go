package protowire

import (
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetRelayBlocks) toWireMessage() (wire.Message, error) {
	if len(x.GetRelayBlocks.Hashes) > wire.MsgGetRelayBlocksHashes {
		return nil, errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(x.GetRelayBlocks.Hashes), wire.MsgGetRelayBlocksHashes)
	}
	hashes, err := protoHashesToWire(x.GetRelayBlocks.Hashes)
	if err != nil {
		return nil, err
	}
	return &wire.MsgGetRelayBlocks{Hashes: hashes}, nil
}

func (x *KaspadMessage_GetRelayBlocks) fromWireMessage(msgGetRelayBlocks *wire.MsgGetRelayBlocks) error {
	if len(msgGetRelayBlocks.Hashes) > wire.MsgGetRelayBlocksHashes {
		return errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(msgGetRelayBlocks.Hashes), wire.MsgGetRelayBlocksHashes)
	}

	x.GetRelayBlocks = &GetRelayBlocksMessage{
		Hashes: wireHashesToProto(msgGetRelayBlocks.Hashes),
	}
	return nil
}
