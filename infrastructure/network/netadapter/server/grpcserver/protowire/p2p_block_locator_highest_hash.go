package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_BlockLocatorHighestHash) toAppMessage() (appmessage.Message, error) {
	highestHash, err := x.BlockLocatorHighestHash.HighestHash.toDomain()
	if err != nil {
		return nil, err
	}

	return &appmessage.MsgBlockLocatorHighestHash{
		HighestHash: highestHash,
	}, nil
}

func (x *KaspadMessage_BlockLocatorHighestHash) fromAppMessage(message *appmessage.MsgBlockLocatorHighestHash) error {
	x.BlockLocatorHighestHash = &BlockLocatorHighestHashMessage{
		HighestHash: domainHashToProto(message.HighestHash),
	}
	return nil
}
