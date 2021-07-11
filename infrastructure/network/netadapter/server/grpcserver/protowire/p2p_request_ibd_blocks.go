package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_RequestIbdBlocks) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_RequestIBDBlocks is nil")
	}
	lowHash, err := x.RequestIbdBlocks.LowHash.toDomain()
	if err != nil {
		return nil, err
	}

	highHash, err := x.RequestIbdBlocks.HighHash.toDomain()
	if err != nil {
		return nil, err
	}

	return &appmessage.MsgRequestIBDBlocks{
		LowHash:  lowHash,
		HighHash: highHash,
	}, nil
}

func (x *RequestIbdBlocksMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RequestIBDBlocksMessage is nil")
	}
	lowHash, err := x.LowHash.toDomain()
	if err != nil {
		return nil, err
	}

	highHash, err := x.HighHash.toDomain()
	if err != nil {
		return nil, err
	}

	return &appmessage.MsgRequestIBDBlocks{
		LowHash:  lowHash,
		HighHash: highHash,
	}, nil

}

func (x *KaspadMessage_RequestIbdBlocks) fromAppMessage(msgRequestmsgRequestIBDBlocks *appmessage.MsgRequestIBDBlocks) error {
	x.RequestIbdBlocks = &RequestIbdBlocksMessage{
		LowHash:  domainHashToProto(msgRequestmsgRequestIBDBlocks.LowHash),
		HighHash: domainHashToProto(msgRequestmsgRequestIBDBlocks.HighHash),
	}
	return nil
}
