package protowire

import "github.com/kaspanet/kaspad/wire"

func (x *KaspadMessage_RequestBlockLocator) toWireMessage() (wire.Message, error) {
	lowHash, err := x.RequestBlockLocator.LowHash.toWire()
	if err != nil {
		return nil, err
	}

	highHash, err := x.RequestBlockLocator.HighHash.toWire()
	if err != nil {
		return nil, err
	}

	return &wire.MsgRequestBlockLocator{
		LowHash:  lowHash,
		HighHash: highHash,
	}, nil
}

func (x *KaspadMessage_RequestBlockLocator) fromWireMessage(msgGetBlockLocator *wire.MsgRequestBlockLocator) error {
	x.RequestBlockLocator = &RequestBlockLocatorMessage{
		LowHash:  wireHashToProto(msgGetBlockLocator.LowHash),
		HighHash: wireHashToProto(msgGetBlockLocator.HighHash),
	}
	return nil
}
