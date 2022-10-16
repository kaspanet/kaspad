package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_StopNotifyTxsConfirmationChangedRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_StopNotifyTxsConfirmationChangedRequest is nil")
	}
	return x.StopNotifyTxsConfirmationChangedRequest.toAppMessage()
}

func (x *KaspadMessage_StopNotifyTxsConfirmationChangedRequest) fromAppMessage(message *appmessage.StopNotifyTxsConfirmationChangedRequestMessage) error {
	x.StopNotifyTxsConfirmationChangedRequest = &StopNotifyTxsConfirmationChangedRequestMessage{
		TxIDs: message.TxIDs,
	}
	return nil
}

func (x *StopNotifyTxsConfirmationChangedRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "StopNotifyTxsConfirmationChangedRequestMessage is nil")
	}
	return &appmessage.StopNotifyTxsConfirmationChangedRequestMessage{
		TxIDs: x.TxIDs,
	}, nil
}

func (x *KaspadMessage_StopNotifyTxsConfirmationChangedResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "StopNotifyTxsConfirmationChangedResponseMessage is nil")
	}
	return x.StopNotifyTxsConfirmationChangedResponse.toAppMessage()
}

func (x *KaspadMessage_StopNotifyTxsConfirmationChangedResponse) fromAppMessage(message *appmessage.StopNotifyTxsConfirmationChangedResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.StopNotifyTxsConfirmationChangedResponse = &StopNotifytTxsConfirmationChangedResponseMessage{
		Error: err,
	}
	return nil
}