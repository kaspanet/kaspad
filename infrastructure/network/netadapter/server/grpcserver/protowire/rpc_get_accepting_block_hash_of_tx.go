package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetAcceptingBlockHashOfTxRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetAcceptingBlockHashOfTxRequest")
	}
	return x.GetAcceptingBlockHashOfTxRequest.toAppMessage()
}

func (x *KaspadMessage_GetAcceptingBlockHashOfTxRequest) fromAppMessage(message *appmessage.GetAcceptingBlockHashOfTxRequestMessage) error {
	x.GetAcceptingBlockHashOfTxRequest = &GetAcceptingBlockHashOfTxRequestMessage{
		TxID: message.TxID,
	}
	return nil
}

func (x *GetAcceptingBlockHashOfTxRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetAcceptingBlockHashOfTxRequestMessage is nil")
	}
	return &appmessage.GetAcceptingBlockHashOfTxRequestMessage{
		TxID: x.TxID,
	}, nil
}

func (x *KaspadMessage_GetAcceptingBlockHashOfTxResponse) toAppMessage() (appmessage.Message, error) {

	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetAcceptingBlockHashOfTxResponse is nil")
	}
	return x.GetAcceptingBlockHashOfTxResponse.toAppMessage()
}

func (x *KaspadMessage_GetAcceptingBlockHashOfTxResponse) fromAppMessage(message *appmessage.GetAcceptingBlockHashOfTxResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.GetAcceptingBlockHashOfTxResponse = &GetAcceptingBlockHashOfTxResponseMessage{
		Hash: message.Hash,

		Error: err,
	}
	return nil
}

func (x *GetAcceptingBlockHashOfTxResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetAcceptingBlockHashOfTxResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	if rpcErr != nil && x.Hash != "" {
		return nil, errors.New("GetAcceptingBlockHashOfTxResponseMessage contains both an error and a response")
	}

	return &appmessage.GetAcceptingBlockHashOfTxResponseMessage{
		Hash:  x.Hash,
		Error: rpcErr,
	}, nil
}
