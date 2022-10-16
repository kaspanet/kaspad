package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_ModifyNotifyingTxsConfirmationChangedRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_ModifyNotifyingTxsConfirmationChangedRequest is nil")
	}
	return x.ModifyNotifyingTxsConfirmationChangedRequest.toAppMessage()
}

func (x *KaspadMessage_ModifyNotifyingTxsConfirmationChangedRequest) fromAppMessage(message *appmessage.ModifyNotifyingTxsConfirmationChangedRequestMessage) error {
	x.ModifyNotifyingTxsConfirmationChangedRequest = &ModifyNotifyingTxsConfirmationChangedRequestMessage{
		AddTxIDs: message.AddTxIDs,
		RemoveTxIDs: message.RemoveTxIDs,
		RequiredConfirmations: message.RequiredConfirmations,
		IncludePending: message.IncludePending,
	}
	return nil
}

func (x *ModifyNotifyingTxsConfirmationChangedRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "ModifyNotifyingTxsConfirmationChangedRequestMessage is nil")
	}
	return &appmessage.ModifyNotifyingTxsConfirmationChangedRequestMessage{
		AddTxIDs: x.AddTxIDs,
		RemoveTxIDs: x.RemoveTxIDs,
		RequiredConfirmations: x.RequiredConfirmations,
		IncludePending: x.IncludePending,
	}, nil
}

func (x *KaspadMessage_ModifyNotifyingTxsConfirmationChangedResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "ModifyNotifyingTxsConfirmationChangedResponseMessage is nil")
	}
	return x.ModifyNotifyingTxsConfirmationChangedResponse.toAppMessage()
}

func (x *ModifyNotifyingTxsConfirmationChangedResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "ModifyNotifyingTxsConfirmationChangedResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}
	return &appmessage.ModifyNotifyingTxsConfirmationChangedResponseMessage{
		Error: rpcErr,
	}, nil
}


func (x *KaspadMessage_ModifyNotifyingTxsConfirmationChangedResponse) fromAppMessage(message *appmessage.ModifyNotifyingTxsConfirmationChangedResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.ModifyNotifyingTxsConfirmationChangedResponse = &ModifyNotifyingTxsConfirmationChangedResponseMessage{
		Error: err,
	}
	return nil
}
