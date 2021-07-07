package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_RequestIBDBlocks) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_RequestIBDBlocks is nil")
	}
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

func (x *RequestIBDBlocksMessage) toAppMessage() (appmessage.Message, error) {
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

func (x *KaspadMessage_RequestIBDBlocks) fromAppMessage(msgRequestmsgRequestIBDBlocks *appmessage.MsgRequestIBDBlocks) error {
	x.RequestIBDBlocks = &RequestIBDBlocksMessage{
		LowHash:  domainHashToProto(msgRequestmsgRequestIBDBlocks.LowHash),
		HighHash: domainHashToProto(msgRequestmsgRequestIBDBlocks.HighHash),
	}
	return nil
}
