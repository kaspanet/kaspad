package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_IbdBlockLocator) toAppMessage() (appmessage.Message, error) {
	targetHash, err := x.IbdBlockLocator.TargetHash.toDomain()
	if err != nil {
		return nil, err
	}
	blockLocatorHash, err := protoHashesToDomain(x.IbdBlockLocator.BlockLocatorHashes)
	if err != nil {
		return nil, err
	}
	return &appmessage.MsgIBDBlockLocator{
		TargetHash:         targetHash,
		BlockLocatorHashes: blockLocatorHash,
	}, nil
}

func (x *KaspadMessage_IbdBlockLocator) fromAppMessage(message *appmessage.MsgIBDBlockLocator) error {
	x.IbdBlockLocator = &IbdBlockLocatorMessage{
		TargetHash:         domainHashToProto(message.TargetHash),
		BlockLocatorHashes: domainHashesToProto(message.BlockLocatorHashes),
	}
	return nil
}
