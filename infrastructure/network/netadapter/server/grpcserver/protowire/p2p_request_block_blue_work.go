package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_RequestBlockBlueWork) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_RequestBlockBlueWork is nil")
	}
	return x.RequestBlockBlueWork.toAppMessage()
}

func (x *RequestBlockBlueWorkMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RequestBlockBlueWorkMessage is nil")
	}
	hash, err := x.Block.toDomain()
	if err != nil {
		return nil, err
	}

	return &appmessage.MsgRequestBlockBlueWork{Hash: hash}, nil

}

func (x *KaspadMessage_RequestBlockBlueWork) fromAppMessage(msgRequestBlockBlueWork *appmessage.MsgRequestBlockBlueWork) error {
	x.RequestBlockBlueWork = &RequestBlockBlueWorkMessage{
		Block: domainHashToProto(msgRequestBlockBlueWork.Hash),
	}
	return nil
}
