package protowire

import "github.com/kaspanet/kaspad/wire"

func (x *KaspadMessage_GetBlocks) toWireMessage() (*wire.MsgGetBlocks, error) {
	lowHash, err := x.GetBlocks.LowHash.toWire()
	if err != nil {
		return nil, err
	}

	highHash, err := x.GetBlocks.HighHash.toWire()
	if err != nil {
		return nil, err
	}

	return &wire.MsgGetBlocks{
		LowHash:  lowHash,
		HighHash: highHash,
	}, nil
}

func (x *KaspadMessage_GetBlocks) fromWireMessage(msgGetBlocks *wire.MsgGetBlocks) error {
	x.GetBlocks = &GetBlocksMessage{
		LowHash:  wireHashToProto(msgGetBlocks.LowHash),
		HighHash: wireHashToProto(msgGetBlocks.HighHash),
	}
	return nil
}
