package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetCurrentNetworkRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetCurrentNetworkRequestMessage{}, nil
}

func (x *KaspadMessage_GetCurrentNetworkRequest) fromAppMessage(_ *appmessage.GetCurrentNetworkRequestMessage) error {
	return nil
}

func (x *KaspadMessage_GetCurrentNetworkResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetCurrentNetworkResponse is nil")
	}
	return x.toAppMessage()
}

func (x *KaspadMessage_GetCurrentNetworkResponse) fromAppMessage(message *appmessage.GetCurrentNetworkResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.GetCurrentNetworkResponse = &GetCurrentNetworkResponseMessage{
		CurrentNetwork: message.CurrentNetwork,
		Error:          err,
	}
	return nil
}

func (x *GetCurrentNetworkResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetCurrentNetworkResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	if rpcErr != nil && len(x.CurrentNetwork) != 0 {
		return nil, errors.New("GetCurrentNetworkResponseMessage contains both an error and a response")
	}

	return &appmessage.GetCurrentNetworkResponseMessage{
		CurrentNetwork: x.CurrentNetwork,
		Error:          rpcErr,
	}, nil
}
