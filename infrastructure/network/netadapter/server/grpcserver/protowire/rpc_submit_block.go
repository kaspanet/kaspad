package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_SubmitBlockRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "SubmitBlockRequestMessage is nil")
	}
	return x.SubmitBlockRequest.toAppMessage()
}

func (x *KaspadMessage_SubmitBlockRequest) fromAppMessage(message *appmessage.SubmitBlockRequestMessage) error {
	x.SubmitBlockRequest = &SubmitBlockRequestMessage{Block: &BlockMessage{}}
	return x.SubmitBlockRequest.Block.fromAppMessage(message.Block)
}

func (x *SubmitBlockRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "SubmitBlockRequestMessage is nil")
	}
	blockAppMessage, err := x.Block.toAppMessage()
	if err != nil {
		return nil, err
	}

	return &appmessage.SubmitBlockRequestMessage{
		Block: blockAppMessage,
	}, nil
}

func (x *KaspadMessage_SubmitBlockResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_SubmitBlockResponse is nil")
	}
	return x.SubmitBlockResponse.toAppMessage()
}

func (x *KaspadMessage_SubmitBlockResponse) fromAppMessage(message *appmessage.SubmitBlockResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.SubmitBlockResponse = &SubmitBlockResponseMessage{
		RejectReason: SubmitBlockResponseMessage_RejectReason(message.RejectReason),
		Error:        err,
	}
	return nil
}

func (x *SubmitBlockResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "SubmitBlockResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}
	return &appmessage.SubmitBlockResponseMessage{
		RejectReason: appmessage.RejectReason(x.RejectReason),
		Error:        rpcErr,
	}, nil
}
