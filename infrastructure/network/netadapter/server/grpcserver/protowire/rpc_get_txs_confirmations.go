package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetTxsConfirmationsRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetTxsConfirmationsRequest")
	}
	return x.GetTxsConfirmationsRequest.toAppMessage()
}

func (x *KaspadMessage_GetTxsConfirmationsRequest) fromAppMessage(message *appmessage.GetTxsConfirmationsRequestMessage) error {
	x.GetTxsConfirmationsRequest = &GetTxsConfirmationsRequestMessage{
		TxIDs: message.TxIDs,
	}
	return nil
}

func (x *GetTxsConfirmationsRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetTxsConfirmationsRequestMessage is nil")
	}
	return &appmessage.GetTxsConfirmationsRequestMessage{
		TxIDs: x.TxIDs,
	}, nil
}

func (x *KaspadMessage_GetTxsConfirmationsResponse) toAppMessage() (appmessage.Message, error) {

	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetTxsConfirmationsResponse is nil")
	}
	return x.GetTxsConfirmationsResponse.toAppMessage()
}

func (x *KaspadMessage_GetTxsConfirmationsResponse) fromAppMessage(message *appmessage.GetTxsConfirmationsResponseMessage) error {
	var rpcErr *RPCError
	if message.Error != nil {
		rpcErr = &RPCError{Message: message.Error.Message}
	}

	rpcTxIDConfirmationsPairs := make([]*TxIDConfirmationsPair, len(message.TxIDConfirmationsPairs))
	for i := range rpcTxIDConfirmationsPairs {
		rpcTxIDConfirmationsPairs[i].fromAppMessage(message.TxIDConfirmationsPairs[i])
	}
	
	x.GetTxsConfirmationsResponse = &GetTxsConfirmationsResponseMessage{
		TxIDConfirmationsPairs: rpcTxIDConfirmationsPairs,

		Error: rpcErr,
	}
	return nil
}

func (x *GetTxsConfirmationsResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetTxsConfirmationsResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	if rpcErr != nil && x.TxIDConfirmationsPairs != nil {
		return nil, errors.New("GetTxsConfirmationsResponseMessage contains both an error and a response")
	}

	appTxIDConfirmationsPairs := make([]*appmessage.TxIDConfirmationsPair, len(x.TxIDConfirmationsPairs))
	for i := range appTxIDConfirmationsPairs {
		appTxIDConfirmationsPairs[i], err = x.TxIDConfirmationsPairs[i].toAppMessage()
		if err != nil {
			return nil, err
		}
	}

	return &appmessage.GetTxsConfirmationsResponseMessage{
		TxIDConfirmationsPairs:  appTxIDConfirmationsPairs,
		Error: rpcErr,
	}, nil
}

func (x *TxIDConfirmationsPair) toAppMessage() (*appmessage.TxIDConfirmationsPair, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "TxIDConfirmationsPair is nil")
	}

	return &appmessage.TxIDConfirmationsPair{
		TxID:  x.TxID,
		Confirmations:  x.Confirmations,
	}, nil
}

func (x *TxIDConfirmationsPair) fromAppMessage(message *appmessage.TxIDConfirmationsPair) {
		
	*x = TxIDConfirmationsPair{
		TxID: message.TxID,
		Confirmations: message.Confirmations,
	}
}
