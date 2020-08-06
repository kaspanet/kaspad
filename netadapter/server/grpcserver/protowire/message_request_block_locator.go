package protowire

import "github.com/kaspanet/kaspad/domainmessage"

func (x *KaspadMessage_RequestBlockLocator) toWireMessage() (domainmessage.Message, error) {
	lowHash, err := x.RequestBlockLocator.LowHash.toWire()
	if err != nil {
		return nil, err
	}

	highHash, err := x.RequestBlockLocator.HighHash.toWire()
	if err != nil {
		return nil, err
	}

	return &domainmessage.MsgRequestBlockLocator{
		LowHash:  lowHash,
		HighHash: highHash,
	}, nil
}

func (x *KaspadMessage_RequestBlockLocator) fromWireMessage(msgGetBlockLocator *domainmessage.MsgRequestBlockLocator) error {
	x.RequestBlockLocator = &RequestBlockLocatorMessage{
		LowHash:  wireHashToProto(msgGetBlockLocator.LowHash),
		HighHash: wireHashToProto(msgGetBlockLocator.HighHash),
	}
	return nil
}
