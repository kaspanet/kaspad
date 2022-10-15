package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_StartNotifyAddressesTxsRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_StartNotifyAddressesTxsRequest is nil")
	}
	return x.StartNotifyAddressesTxsRequest.toAppMessage()
}

func (x *KaspadMessage_StartNotifyAddressesTxsRequest) fromAppMessage(message *appmessage.StartNotifyAddressesTxsRequestMessage) error {
	x.StartNotifyAddressesTxsRequest = &StartNotifyAddressesTxsRequestMessage{
		Addresses: message.Addresses,
	}
	return nil
}

func (x *StartNotifyAddressesTxsRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "StartNotifyAddressesTxsRequestMessage is nil")
	}
	return &appmessage.StartNotifyAddressesTxsRequestMessage{
		Addresses: x.Addresses,
	}, nil
}

func (x *KaspadMessage_StartNotifyAddressesTxsResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "StartNotifyAddressesTxsResponseMessage is nil")
	}
	return x.StartNotifyAddressesTxsResponse.toAppMessage()
}

func (x *KaspadMessage_StartNotifyAddressesTxsResponse) fromAppMessage(message *appmessage.StartNotifyAddressesTxsResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.StartNotifyAddressesTxsResponse = &StartNotifytTxsConfirmationChangedResponseMessage{
		Error: err,
	}
	return nil
}
