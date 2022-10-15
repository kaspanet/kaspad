package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_NotifyTxsConfirmationChangedRequst) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_NotifyTxsConfirmationChangedRequest is nil")
	}
	return x.NotifyTxsConfirmationChangedRequst.toAppMessage()
}

func (x *KaspadMessage_NotifyTxsConfirmationChangedRequst) fromAppMessage(message *appmessage.NotifyTxsConfirmationChangedRequstMessage) error {
	x.NotifyTxsConfirmationChangedRequst = &NotifyTxsConfirmationChangedRequstMessage{
		TxIDs: message.TxIDs,
		RequiredConfirmations: message.RequiredConfirmations,
		IncludePending: message.IncludePending,
	}
	return nil
}

func (x *NotifyTxsConfirmationChangedRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "NotifyTxsConfirmationChangedRequestMessage is nil")
	}
	return &appmessage.NotifyTxsConfirmationChangedRequestMessage{
		TxIDs: x.TxIDs,
		RequiredConfirmations: x.RequiredConfirmations,
		IncludePending: x.IncludePending,
	}, nil
}

func (x *KaspadMessage_NotifyTxsConfirmationChangedResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "NotifyTxsConfirmationChangedResponseMessage is nil")
	}
	return x.NotifyTxsConfirmationChangedResponse.toAppMessage()
}

func (x *KaspadMessage_NotifyTxsConfirmationChangedResponse) fromAppMessage(message *appmessage.NotifyTxsConfirmationChangedResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.NotifyTxsConfirmationChangedResponse = &NotifyTxsConfirmationChangedResponseMessage{
		Error: err,
	}
	return nil
}

func (x *NotifyTxsConfirmationChangedResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "NotifyTxsConfirmationChangedResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}
	return &appmessage.NotifyTxsConfirmationChangedResponseMessage{
		Error: rpcErr,
	}, nil
}

func (x *KaspadMessage_TxsConfirmationChangedNotification) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_TxsConfirmationChangedNotification is nil")
	}
	return x.TxsConfirmationChangedNotification.toAppMessage()
}

func (x *KaspadMessage_TxsConfirmationChangedNotification) fromAppMessage(message *appmessage.TxsConfirmationChangedNotificationMessage) error {
	pending := make([]*RpcTxIDConfirmationsPair, len(message.Pending))
	for i, entry := range message.Pending {
		pending[i] = &RpcTxIDConfirmationsPair{}
		pending[i].fromAppMessage(entry)
	}

	confirmed := make([]*RpcTxIDConfirmationsPair, len(message.Confirmed))
	for i, entry := range message.Confirmed {
		confirmed[i] = &RpcTxIDConfirmationsPair{}
		confirmed[i].fromAppMessage(entry)
	}

	x.TxsConfirmationChangedNotification = &TxsConfirmationChangedNotificationMessage{
		RequiredConfirmations: message.RequiredConfirmations,
		Pending: pending,
		Confirmed: confirmed,
		UnconfirmedTxIds: message.UnconfirmedTxIds,
	}
	return nil
}

func (x *TxsConfirmationChangedNotificationMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "TxsConfirmationChangedNotificationMessage is nil")
	}
	pending := make([]*appmessage.TxIDConfirmationsPair, len(x.Pending))
	for i, confirmationPair := range x.Pending {
		appConfirmationPair, err := confirmationPair.toAppMessage()
		if err != nil {
			return nil, err
		}
		pending[i] = appConfirmationPair
	}

	confirmed := make([]*appmessage.TxIDConfirmationsPair, len(x.Confirmed))
	for i, appConfirmationPair := range x.Confirmed {
		appConfirmationPair, err := appConfirmationPair.toAppMessage()
		if err != nil {
			return nil, err
		}
		confirmed[i] = appConfirmationPair
	}

	return &appmessage.TxsConfirmationChangedNotificationMessage{
		RequiredConfirmations: x.RequiredConfirmations,
		Pending: pending,
		Confirmed: confirmed,
		UnconfirmedTxIds: x.UnconfirmedTxIds,
	}, nil
}
