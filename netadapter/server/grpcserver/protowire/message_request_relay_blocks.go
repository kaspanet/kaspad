package protowire

import (
	"github.com/kaspanet/kaspad/domainmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_RequestRelayBlocks) toWireMessage() (domainmessage.Message, error) {
	if len(x.RequestRelayBlocks.Hashes) > domainmessage.MsgRequestRelayBlocksHashes {
		return nil, errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(x.RequestRelayBlocks.Hashes), domainmessage.MsgRequestRelayBlocksHashes)
	}
	hashes, err := protoHashesToWire(x.RequestRelayBlocks.Hashes)
	if err != nil {
		return nil, err
	}
	return &domainmessage.MsgRequestRelayBlocks{Hashes: hashes}, nil
}

func (x *KaspadMessage_RequestRelayBlocks) fromWireMessage(msgGetRelayBlocks *domainmessage.MsgRequestRelayBlocks) error {
	if len(msgGetRelayBlocks.Hashes) > domainmessage.MsgRequestRelayBlocksHashes {
		return errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(msgGetRelayBlocks.Hashes), domainmessage.MsgRequestRelayBlocksHashes)
	}

	x.RequestRelayBlocks = &RequestRelayBlocksMessage{
		Hashes: wireHashesToProto(msgGetRelayBlocks.Hashes),
	}
	return nil
}
