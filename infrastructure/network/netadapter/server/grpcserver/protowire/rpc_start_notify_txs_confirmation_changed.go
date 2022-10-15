package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_StartNotifyTxsConfirmationChangedRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_StartNotifyTxsConfirmationChangedRequest is nil")
	}
	return x.StartNotifyTxsConfirmationChangedRequest.toAppMessage()
}

func (x *KaspadMessage_StartNotifyTxsConfirmationChangedRequest) fromAppMessage(message *appmessage.StartNotifyTxsConfirmationChangedRequestMessage) error {
	x.StartNotifyTxsConfirmationChangedRequest = &StartNotifyTxsConfirmationChangedRequestMessage{
		TxIDs: message.TxIDs,
	}
	return nil
}

func (x *StartNotifyTxsConfirmationChangedRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "StartNotifyTxsConfirmationChangedRequestMessage is nil")
	}
	return &appmessage.StartNotifyTxsConfirmationChangedRequestMessage{
		TxIDs: x.TxIDs,
	}, nil
}

func (x *KaspadMessage_StartNotifyTxsConfirmationChangedResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "StartNotifyTxsConfirmationChangedResponseMessage is nil")
	}
	return x.StartNotifyTxsConfirmationChangedResponse.toAppMessage()
}

func (x *KaspadMessage_StartNotifyTxsConfirmationChangedResponse) fromAppMessage(message *appmessage.StartNotifyTxsConfirmationChangedResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.StartNotifyTxsConfirmationChangedResponse = &StartNotifytTxsConfirmationChangedResponseMessage{
		Error: err,
	}
	return nil
}
