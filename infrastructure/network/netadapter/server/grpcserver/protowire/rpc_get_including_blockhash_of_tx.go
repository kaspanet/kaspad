package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetIncludingBlockHashOfTxRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetIncludingBlockHashOfTxRequest")
	}
	return x.GetIncludingBlockHashOfTxRequest.toAppMessage()
}

func (x *KaspadMessage_GetIncludingBlockHashOfTxRequest) fromAppMessage(message *appmessage.GetIncludingBlockHashOfTxRequestMessage) error {
	x.GetIncludingBlockHashOfTxRequest = &GetIncludingBlockHashOfTxRequestMessage{
		TxID: message.TxID,
	}
	return nil
}

func (x *GetIncludingBlockHashOfTxRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetIncludingBlockHashOfTxRequestMessage is nil")
	}
	return &appmessage.GetIncludingBlockHashOfTxRequestMessage{
		TxID: x.TxID,
	}, nil
}

func (x *KaspadMessage_GetIncludingBlockHashOfTxResponse) toAppMessage() (appmessage.Message, error) {

	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetIncludingBlockHashOfTxResponse is nil")
	}
	return x.GetIncludingBlockHashOfTxResponse.toAppMessage()
}

func (x *KaspadMessage_GetIncludingBlockHashOfTxResponse) fromAppMessage(message *appmessage.GetIncludingBlockHashOfTxResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.GetIncludingBlockHashOfTxResponse = &GetIncludingBlockHashOfTxResponseMessage{
		Hash: message.Hash,

		Error: err,
	}
	return nil
}

func (x *GetIncludingBlockHashOfTxResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetIncludingBlockHashOfTxResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	if rpcErr != nil && x.Hash != "" {
		return nil, errors.New("GetIncludingBlockHashOfTxResponseMessage contains both an error and a response")
	}

	return &appmessage.GetIncludingBlockHashOfTxResponseMessage{
		Hash:  x.Hash,
		Error: rpcErr,
	}, nil
}
