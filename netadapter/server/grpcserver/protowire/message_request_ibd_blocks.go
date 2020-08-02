package protowire

import "github.com/kaspanet/kaspad/wire"

func (x *KaspadMessage_RequestIBDBlocks) toWireMessage() (wire.Message, error) {
	lowHash, err := x.RequestIBDBlocks.LowHash.toWire()
	if err != nil {
		return nil, err
	}

	highHash, err := x.RequestIBDBlocks.HighHash.toWire()
	if err != nil {
		return nil, err
	}

	return &wire.MsgRequestIBDBlocks{
		LowHash:  lowHash,
		HighHash: highHash,
	}, nil
}

func (x *KaspadMessage_RequestIBDBlocks) fromWireMessage(msgGetBlocks *wire.MsgRequestIBDBlocks) error {
	x.RequestIBDBlocks = &RequestIBDBlocksMessage{
		LowHash:  wireHashToProto(msgGetBlocks.LowHash),
		HighHash: wireHashToProto(msgGetBlocks.HighHash),
	}
	return nil
}
