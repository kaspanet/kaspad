package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_ModifyNotifyTxsConfirmationChangedParamsRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_ModifyNotifyTxsConfirmationChangedParamsRequest is nil")
	}
	return x.ModifyNotifyTxsConfirmationChangedParamsRequest.toAppMessage()
}

func (x *KaspadMessage_ModifyNotifyTxsConfirmationChangedParamsRequest) fromAppMessage(message *appmessage.ModifyNotifyTxsConfirmationChangedParamsRequestMessage) error {
	x.ModifyNotifyTxsConfirmationChangedParamsRequest = &ModifyNotifyTxsConfirmationChangedParamsRequestMessage{
		RequiredConfirmations: message.RequiredConfirmations,
		IncludePending: message.IncludePending,
	}
	return nil
}

func (x *ModifyNotifyTxsConfirmationChangedParamsRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "ModifyNotifyTxsConfirmationChangedParamsRequestMessage is nil")
	}
	return &appmessage.ModifyNotifyTxsConfirmationChangedParamsRequestMessage{
		RequiredConfirmations: x.RequiredConfirmations,
		IncludePending: x.IncludePending,
	}, nil
}

func (x *KaspadMessage_ModifyNotifyTxsConfirmationChangedParamsResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "ModifyNotifyTxsConfirmationChangedParamsResponseMessage is nil")
	}
	return x.ModifyNotifyTxsConfirmationChangedParamsResponse.toAppMessage()
}

func (x *KaspadMessage_ModifyNotifyTxsConfirmationChangedParamsResponse) fromAppMessage(message *appmessage.ModifyNotifyTxsConfirmationChangedParamsResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.ModifyNotifyTxsConfirmationChangedParamsResponse = &ModifyNotifytTxsConfirmationChangedParamsResponseMessage{
		Error: err,
	}
	return nil
}
