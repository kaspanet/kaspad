package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetCoinSupplyRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetCoinSupplyRequestMessage{}, nil
}

func (x *KaspadMessage_GetCoinSupplyRequest) fromAppMessage(_ *appmessage.GetCoinSupplyRequestMessage) error {
	x.GetCoinSupplyRequest = &GetCoinSupplyRequestMessage{}
	return nil
}

func (x *KaspadMessage_GetCoinSupplyResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetCoinSupplyResponse is nil")
	}
	return x.GetCoinSupplyResponse.toAppMessage()
}

func (x *KaspadMessage_GetCoinSupplyResponse) fromAppMessage(message *appmessage.GetCoinSupplyResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.GetCoinSupplyResponse = &GetCoinSupplyResponseMessage{
		TotalSompi:       message.TotalSompi,
		CirculatingSompi: message.CirculatingSompi,

		Error: err,
	}
	return nil
}

func (x *GetCoinSupplyResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetCoinSupplyResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	return &appmessage.GetCoinSupplyResponseMessage{
		TotalSompi:       x.TotalSompi,
		CirculatingSompi: x.CirculatingSompi,

		Error: rpcErr,
	}, nil
}
