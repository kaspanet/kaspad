package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_ModifyNotifyAddressesTxsParamsRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_ModifyNotifyAddressesTxsParamsRequest is nil")
	}
	return x.ModifyNotifyAddressesTxsParamsRequest.toAppMessage()
}

func (x *KaspadMessage_ModifyNotifyAddressesTxsParamsRequest) fromAppMessage(message *appmessage.ModifyNotifyAddressesTxsParamsRequestMessage) error {
	x.ModifyNotifyAddressesTxsParamsRequest = &ModifyNotifyAddressesTxsParamsRequestMessage{
		RequiredConfirmations: message.RequiredConfirmations,
		IncludePending: message.IncludePending,
		IncludeSending: message.IncludeSending,
		IncludeReceiving: message.IncludeReceiving,
	}
	return nil
}

func (x *ModifyNotifyAddressesTxsParamsRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "ModifyNotifyAddressesTxsParamsRequestMessage is nil")
	}
	return &appmessage.ModifyNotifyAddressesTxsParamsRequestMessage{
		RequiredConfirmations: x.RequiredConfirmations,
		IncludePending: x.IncludePending,
		IncludeSending: x.IncludeSending,
		IncludeReceiving: x.IncludeReceiving,
	}, nil
}

func (x *KaspadMessage_ModifyNotifyAddressesTxsParamsResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "ModifyNotifyAddressesTxsParamsResponseMessage is nil")
	}
	return x.ModifyNotifyAddressesTxsParamsResponse.toAppMessage()
}

func (x *KaspadMessage_ModifyNotifyAddressesTxsParamsResponse) fromAppMessage(message *appmessage.ModifyNotifyAddressesTxsParamsResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.ModifyNotifyAddressesTxsParamsResponse = &ModifyNotifytTxsConfirmationChangedParamsResponseMessage{
		Error: err,
	}
	return nil
}
