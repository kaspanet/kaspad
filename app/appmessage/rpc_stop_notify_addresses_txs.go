package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_StopNotifyAddressesTxsRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_StopNotifyAddressesTxsRequest is nil")
	}
	return x.StopNotifyAddressesTxsRequest.toAppMessage()
}

func (x *KaspadMessage_StopNotifyAddressesTxsRequest) fromAppMessage(message *appmessage.StopNotifyAddressesTxsRequestMessage) error {
	x.StopNotifyAddressesTxsRequest = &StopNotifyAddressesTxsRequestMessage{
		Addresses: message.Addresses,
	}
	return nil
}

func (x *StopNotifyAddressesTxsRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "StopNotifyAddressesTxsRequestMessage is nil")
	}
	return &appmessage.StopNotifyAddressesTxsRequestMessage{
		Addresses: x.Addresses,
	}, nil
}

func (x *KaspadMessage_StopNotifyAddressesTxsResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "StopNotifyAddressesTxsResponseMessage is nil")
	}
	return x.StopNotifyAddressesTxsResponse.toAppMessage()
}

func (x *KaspadMessage_StopNotifyAddressesTxsResponse) fromAppMessage(message *appmessage.StopNotifyAddressesTxsResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.StopNotifyAddressesTxsResponse = &StopNotifytTxsConfirmationChangedResponseMessage{
		Error: err,
	}
	return nil
}