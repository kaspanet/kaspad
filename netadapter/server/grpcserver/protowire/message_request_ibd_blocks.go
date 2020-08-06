package protowire

import "github.com/kaspanet/kaspad/domainmessage"

func (x *KaspadMessage_RequestIBDBlocks) toWireMessage() (domainmessage.Message, error) {
	lowHash, err := x.RequestIBDBlocks.LowHash.toWire()
	if err != nil {
		return nil, err
	}

	highHash, err := x.RequestIBDBlocks.HighHash.toWire()
	if err != nil {
		return nil, err
	}

	return &domainmessage.MsgRequestIBDBlocks{
		LowHash:  lowHash,
		HighHash: highHash,
	}, nil
}

func (x *KaspadMessage_RequestIBDBlocks) fromWireMessage(msgGetBlocks *domainmessage.MsgRequestIBDBlocks) error {
	x.RequestIBDBlocks = &RequestIBDBlocksMessage{
		LowHash:  wireHashToProto(msgGetBlocks.LowHash),
		HighHash: wireHashToProto(msgGetBlocks.HighHash),
	}
	return nil
}
