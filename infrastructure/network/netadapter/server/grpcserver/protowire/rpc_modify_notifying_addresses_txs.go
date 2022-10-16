package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_ModifyNotifyingAddressesTxsRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_ModifyNotifyingAddressesTxsRequest is nil")
	}
	return x.ModifyNotifyingAddressesTxsRequest.toAppMessage()
}

func (x *KaspadMessage_ModifyNotifyingAddressesTxsRequest) fromAppMessage(message *appmessage.ModifyNotifyingAddressesTxsRequestMessage) error {
	x.ModifyNotifyingAddressesTxsRequest = &ModifyNotifyingAddressesTxsRequestMessage{
		RequiredConfirmations: message.RequiredConfirmations,
		IncludePending: message.IncludePending,
		IncludeSending: message.IncludeSending,
		IncludeReceiving: message.IncludeReceiving,
	}
	return nil
}

func (x *ModifyNotifyingAddressesTxsRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "ModifyNotifyingAddressesTxsRequestMessage is nil")
	}
	return &appmessage.ModifyNotifyingAddressesTxsRequestMessage{
		RequiredConfirmations: x.RequiredConfirmations,
		IncludePending: x.IncludePending,
		IncludeSending: x.IncludeSending,
		IncludeReceiving: x.IncludeReceiving,
	}, nil
}

func (x *KaspadMessage_ModifyNotifyingAddressesTxsResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "ModifyNotifyingAddressesTxsResponseMessage is nil")
	}
	return x.ModifyNotifyingAddressesTxsResponse.toAppMessage()
}

func (x *KaspadMessage_ModifyNotifyingAddressesTxsResponse) fromAppMessage(message *appmessage.ModifyNotifyingAddressesTxsResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.ModifyNotifyingAddressesTxsResponse = &ModifyNotifytTxsConfirmingationChangedResponseMessage{
		Error: err,
	}
	return nil
}
