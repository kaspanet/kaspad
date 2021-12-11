package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetBalanceByAddressRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetBalanceByAddressRequest is nil")
	}
	return x.GetBalanceByAddressRequest.toAppMessage()
}

func (x *KaspadMessage_GetBalanceByAddressRequest) fromAppMessage(message *appmessage.GetBalanceByAddressRequestMessage) error {
	x.GetBalanceByAddressRequest = &GetBalanceByAddressRequestMessage{
		Address: message.Addresses,
	}
	return nil
}

func (x *GetBalanceByAddressRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetBalanceByAddressRequestMessage is nil")
	}
	return &appmessage.GetBalanceByAddressRequestMessage{
		Address: x.Address,
	}, nil
}

func (x *KaspadMessage_GetBalanceByAddressResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetBalanceByAddressResponseMessage is nil")
	}
	return x.GetBalanceByAddressResponse.toAppMessage()
}

func (x *KaspadMessage_GetBalanceByAddressResponse) fromAppMessage(message *appmessage.GetBalanceByAddressResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.GetBalanceByAddressResponse = &GetBalanceByAddressResponseMessage{
		balance: message.balance,
	}
	return nil
}

func (x *GetBalanceByAddressResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetBalanceByAddressResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	if rpcErr != nil && len(x.Entries) != 0 {
		return nil, errors.New("GetBalanceByAddressResponseMessage contains both an error and a response")
	}

	return &appmessage.GetBalanceByAddressResponseMessage{
		balance: message.balance,
		Error:   rpcErr,
	}, nil
}
