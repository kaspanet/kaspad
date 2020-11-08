package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_RequestIBDBlocks) toAppMessage() (appmessage.Message, error) {
	lowHash, err := x.RequestIBDBlocks.LowHash.toDomain()
	if err != nil {
		return nil, err
	}

	highHash, err := x.RequestIBDBlocks.HighHash.toDomain()
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
		LowHash:  domainHashToProto(msgGetBlocks.LowHash),
		HighHash: domainHashToProto(msgGetBlocks.HighHash),
	}
	return nil
}
