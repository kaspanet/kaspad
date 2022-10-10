package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetTxConfirmationsRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetTxConfirmationsRequest")
	}
	return x.GetTxConfirmationsRequest.toAppMessage()
}

func (x *KaspadMessage_GetTxConfirmationsRequest) fromAppMessage(message *appmessage.GetTxConfirmationsRequestMessage) error {
	x.GetTxConfirmationsRequest = &GetTxConfirmationsRequestMessage{
		TxID: message.TxID,
	}
	return nil
}

func (x *GetTxConfirmationsRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetTxConfirmationsRequestMessage is nil")
	}
	return &appmessage.GetTxConfirmationsRequestMessage{
		TxID: x.TxID,
	}, nil
}

func (x *KaspadMessage_GetTxConfirmationsResponse) toAppMessage() (appmessage.Message, error) {

	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetTxConfirmationsResponse is nil")
	}
	return x.GetTxConfirmationsResponse.toAppMessage()
}

func (x *KaspadMessage_GetTxConfirmationsResponse) fromAppMessage(message *appmessage.GetTxConfirmationsResponseMessage) error {
	var rpcErr *RPCError
	if message.Error != nil {
		rpcErr = &RPCError{Message: message.Error.Message}
	}
	
	x.GetTxConfirmationsResponse = &GetTxConfirmationsResponseMessage{
		Confirmations: message.Confirmations,

		Error: rpcErr,
	}
	return nil
}

func (x *GetTxConfirmationsResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetTxConfirmationsResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return &appmessage.GetTxConfirmationsResponseMessage{
		Confirmations:  x.Confirmations,
		Error: rpcErr,
	}, nil
}
