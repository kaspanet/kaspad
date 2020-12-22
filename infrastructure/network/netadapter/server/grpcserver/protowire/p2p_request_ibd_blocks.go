package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_RequestIBDBlocks) toAppMessage() (appmessage.Message, error) {
	hashes, err := protoHashesToDomain(x.RequestIBDBlocks.Hashes)
	if err != nil {
		return nil, err
	}
	return &appmessage.MsgRequestIBDBlocks{Hashes: hashes}, nil
}

func (x *KaspadMessage_RequestIBDBlocks) fromAppMessage(msgRequestIBDBlocks *appmessage.MsgRequestIBDBlocks) error {
	x.RequestIBDBlocks = &RequestIBDBlocksMessage{
		Hashes: domainHashesToProto(msgRequestIBDBlocks.Hashes),
	}

	return nil
}
