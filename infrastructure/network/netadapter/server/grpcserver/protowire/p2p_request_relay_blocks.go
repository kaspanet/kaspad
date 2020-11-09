package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_RequestRelayBlocks) toAppMessage() (appmessage.Message, error) {
	if len(x.RequestRelayBlocks.Hashes) > appmessage.MsgRequestRelayBlocksHashes {
		return nil, errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(x.RequestRelayBlocks.Hashes), appmessage.MsgRequestRelayBlocksHashes)
	}
	hashes, err := protoHashesToDomain(x.RequestRelayBlocks.Hashes)
	if err != nil {
		return nil, err
	}
	return &appmessage.MsgRequestRelayBlocks{Hashes: hashes}, nil
}

func (x *KaspadMessage_RequestRelayBlocks) fromAppMessage(msgGetRelayBlocks *appmessage.MsgRequestRelayBlocks) error {
	if len(msgGetRelayBlocks.Hashes) > appmessage.MsgRequestRelayBlocksHashes {
		return errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(msgGetRelayBlocks.Hashes), appmessage.MsgRequestRelayBlocksHashes)
	}

	x.RequestRelayBlocks = &RequestRelayBlocksMessage{
		Hashes: domainHashesToProto(msgGetRelayBlocks.Hashes),
	}
	return nil
}
