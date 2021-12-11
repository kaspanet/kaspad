package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetBalanceByAddressRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetBalanceByAddressRequest is nil")
	}
	return x.GetBalanceByAddressRequestMessage.toAppMessage()
}

func (x *KaspadMessage_GetBalanceByAddressRequestMessage) fromAppMessage(message *appmessage.GetBalanceByAddressRequest) error {
	x.GetBalanceByAddressRequestMessage = &GetBalanceByAddressRequestMessage{
		Address: message.Address,
	}
	return nil
}

func (x *GetBalanceByAddressRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetBalanceByAddressRequest is nil")
	}
	return &appmessage.GetBalanceByAddressRequest{
		Address: x.Address,
	}, nil
}

func (x *KaspadMessage_GetBalanceByAddressResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetBalanceByAddressResponse is nil")
	}
	return x.GetBalanceByAddressResponseMessage.toAppMessage()
}

func (x *KaspadMessage_GetBalanceByAddressResponseMessage) fromAppMessage(message *appmessage.GetBalanceByAddressResponse) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.GetBalanceByAddressResponseMessage = &GetBalanceByAddressResponseMessage{
		Balance: message.Balance,

		Error: err,
	}
	return nil
}

func (x *GetBalanceByAddressResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetBalanceByAddressResponse is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	if rpcErr != nil && x.Balance != 1 {
		return nil, errors.New("GetBalanceByAddressResponse contains both an error and a response")
	}

	return &appmessage.GetBalanceByAddressResponse{
		Balance: x.Balance,
		Error:   rpcErr,
	}, nil
}
