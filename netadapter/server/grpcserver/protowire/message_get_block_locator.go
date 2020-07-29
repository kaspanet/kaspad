package protowire

import "github.com/kaspanet/kaspad/wire"

func (x *KaspadMessage_GetBlockLocator) toWireMessage() (wire.Message, error) {
	lowHash, err := x.GetBlockLocator.LowHash.toWire()
	if err != nil {
		return nil, err
	}

	highHash, err := x.GetBlockLocator.HighHash.toWire()
	if err != nil {
		return nil, err
	}

	return &wire.MsgGetBlockLocator{
		LowHash:  lowHash,
		HighHash: highHash,
	}, nil
}

func (x *KaspadMessage_GetBlockLocator) fromWireMessage(msgGetBlockLocator *wire.MsgGetBlockLocator) error {
	x.GetBlockLocator = &GetBlockLocatorMessage{
		LowHash:  wireHashToProto(msgGetBlockLocator.LowHash),
		HighHash: wireHashToProto(msgGetBlockLocator.HighHash),
	}
	return nil
}
