package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_ModifyNotifyingAddressesTxsRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_ModifyNotifyingAddressesTxsRequest is nil")
	}
	return x.ModifyNotifyingAddressesTxsRequest.toAppMessage()
}

func (x *KaspadMessage_ModifyNotifyingAddressesTxsRequest) fromAppMessage(message *appmessage.ModifyNotifyingAddressesTxsRequestMessage) error {
	x.ModifyNotifyingAddressesTxsRequest = &ModifyNotifyingAddressesTxsRequestMessage{
		AddAddresses: message.AddAddresses,
		RemoveAddresses: message.RemoveAddresses,
		RequiredConfirmations: message.RequiredConfirmations,
		IncludePending: message.IncludePending,
		IncludeSending: message.IncludeSending,
		IncludeReceiving: message.IncludeReceiving,
	}
	return nil
}

func (x *ModifyNotifyingAddressesTxsRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "ModifyNotifyingAddressesTxsRequestMessage is nil")
	}
	return &appmessage.ModifyNotifyingAddressesTxsRequestMessage{
		AddAddresses: x.AddAddresses,
		RemoveAddresses: x.RemoveAddresses,
		RequiredConfirmations: x.RequiredConfirmations,
		IncludePending: x.IncludePending,
		IncludeSending: x.IncludeSending,
		IncludeReceiving: x.IncludeReceiving,
	}, nil
}

func (x *KaspadMessage_ModifyNotifyingAddressesTxsResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "ModifyNotifyingAddressesTxsResponseMessage is nil")
	}
	return x.ModifyNotifyingAddressesTxsResponse.toAppMessage()
}

func (x *ModifyNotifyingAddressesTxsResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "ModifyNotifyingAddressesTxsResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}
	return &appmessage.ModifyNotifyingAddressesTxsResponseMessage{
		Error: rpcErr,
	}, nil
}

func (x *KaspadMessage_ModifyNotifyingAddressesTxsResponse) fromAppMessage(message *appmessage.ModifyNotifyingAddressesTxsResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.ModifyNotifyingAddressesTxsResponse = &ModifyNotifyingAddressesTxsResponseMessage{
		Error: err,
	}
	return nil
}
