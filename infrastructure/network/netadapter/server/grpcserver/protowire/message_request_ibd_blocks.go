package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_RequestIBDBlocks) toAppMessage() (appmessage.Message, error) {
	lowHash, err := x.RequestIBDBlocks.LowHash.toWire()
	if err != nil {
		return nil, err
	}

	highHash, err := x.RequestIBDBlocks.HighHash.toWire()
	if err != nil {
		return nil, err
	}

	return &appmessage.MsgRequestIBDBlocks{
		LowHash:  lowHash,
		HighHash: highHash,
	}, nil
}

func (x *KaspadMessage_RequestIBDBlocks) fromAppMessage(msgGetBlocks *appmessage.MsgRequestIBDBlocks) error {
	x.RequestIBDBlocks = &RequestIBDBlocksMessage{
		LowHash:  wireHashToProto(msgGetBlocks.LowHash),
		HighHash: wireHashToProto(msgGetBlocks.HighHash),
	}
	return nil
}
