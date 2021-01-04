package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_IbdBlockLocatorHighestHash) toAppMessage() (appmessage.Message, error) {
	highestHash, err := x.IbdBlockLocatorHighestHash.HighestHash.toDomain()
	if err != nil {
		return nil, err
	}

	return &appmessage.MsgIBDBlockLocatorHighestHash{
		HighestHash: highestHash,
	}, nil
}

func (x *KaspadMessage_IbdBlockLocatorHighestHash) fromAppMessage(message *appmessage.MsgIBDBlockLocatorHighestHash) error {
	x.IbdBlockLocatorHighestHash = &IbdBlockLocatorHighestHashMessage{
		HighestHash: domainHashToProto(message.HighestHash),
	}
	return nil
}
