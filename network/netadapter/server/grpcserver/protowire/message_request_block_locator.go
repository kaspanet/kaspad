package protowire

import "github.com/kaspanet/kaspad/network/appmessage"

func (x *KaspadMessage_RequestBlockLocator) toDomainMessage() (appmessage.Message, error) {
	lowHash, err := x.RequestBlockLocator.LowHash.toWire()
	if err != nil {
		return nil, err
	}

	highHash, err := x.RequestBlockLocator.HighHash.toWire()
	if err != nil {
		return nil, err
	}

	return &appmessage.MsgRequestBlockLocator{
		LowHash:  lowHash,
		HighHash: highHash,
	}, nil
}

func (x *KaspadMessage_RequestBlockLocator) fromDomainMessage(msgGetBlockLocator *appmessage.MsgRequestBlockLocator) error {
	x.RequestBlockLocator = &RequestBlockLocatorMessage{
		LowHash:  wireHashToProto(msgGetBlockLocator.LowHash),
		HighHash: wireHashToProto(msgGetBlockLocator.HighHash),
	}
	return nil
}
