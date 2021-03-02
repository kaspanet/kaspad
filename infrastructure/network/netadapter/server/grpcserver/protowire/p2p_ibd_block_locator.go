package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_IbdBlockLocator) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_IbdBlockLocator is nil")
	}
	return x.IbdBlockLocator.toAppMessage()
}

func (x *IbdBlockLocatorMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "IbdBlockLocatorMessage is nil")
	}
	targetHash, err := x.TargetHash.toDomain()
	if err != nil {
		return nil, err
	}
	blockLocatorHash, err := protoHashesToDomain(x.BlockLocatorHashes)
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
