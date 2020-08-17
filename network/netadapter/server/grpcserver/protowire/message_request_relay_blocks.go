package protowire

import (
	"github.com/kaspanet/kaspad/network/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_RequestRelayBlocks) toDomainMessage() (appmessage.Message, error) {
	if len(x.RequestRelayBlocks.Hashes) > appmessage.MsgRequestRelayBlocksHashes {
		return nil, errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(x.RequestRelayBlocks.Hashes), appmessage.MsgRequestRelayBlocksHashes)
	}
	hashes, err := protoHashesToWire(x.RequestRelayBlocks.Hashes)
	if err != nil {
		return nil, err
	}
	return &appmessage.MsgRequestRelayBlocks{Hashes: hashes}, nil
}

func (x *KaspadMessage_RequestRelayBlocks) fromDomainMessage(msgGetRelayBlocks *appmessage.MsgRequestRelayBlocks) error {
	if len(msgGetRelayBlocks.Hashes) > appmessage.MsgRequestRelayBlocksHashes {
		return errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(msgGetRelayBlocks.Hashes), appmessage.MsgRequestRelayBlocksHashes)
	}

	x.RequestRelayBlocks = &RequestRelayBlocksMessage{
		Hashes: wireHashesToProto(msgGetRelayBlocks.Hashes),
	}
	return nil
}
