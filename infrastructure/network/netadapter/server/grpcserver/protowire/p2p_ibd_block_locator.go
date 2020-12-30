package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_IbdBlockLocator) toAppMessage() (appmessage.Message, error) {
	hashes, err := protoHashesToDomain(x.IbdBlockLocator.Hashes)
	if err != nil {
		return nil, err
	}
	return &appmessage.MsgIBDBlockLocator{Hashes: hashes}, nil
}

func (x *KaspadMessage_IbdBlockLocator) fromAppMessage(message *appmessage.MsgIBDBlockLocator) error {
	x.IbdBlockLocator = &IbdBlockLocatorMessage{
		Hashes: domainHashesToProto(message.Hashes),
	}
	return nil
}
