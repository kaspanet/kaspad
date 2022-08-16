package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_NotifyBlockAddedRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_NotifyBlockAddedRequest is nil")
	}
	return x.NotifyBlockAddedRequest.toAppMessage()
}

func (x *KaspadMessage_NotifyBlockAddedRequest) fromAppMessage(message *appmessage.NotifyBlockAddedRequestMessage) error {
	x.NotifyBlockAddedRequest = &NotifyBlockAddedRequestMessage{Id: message.ID}
	return nil
}

func (x *NotifyBlockAddedRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "NotifyBlockAddedRequest is nil")
	}
	return &appmessage.NotifyBlockAddedRequestMessage{
		ID: x.Id,
	}, nil
}

func (x *KaspadMessage_NotifyBlockAddedResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_NotifyBlockAddedResponse is nil")
	}
	return x.NotifyBlockAddedResponse.toAppMessage()
}

func (x *KaspadMessage_NotifyBlockAddedResponse) fromAppMessage(message *appmessage.NotifyBlockAddedResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.NotifyBlockAddedResponse = &NotifyBlockAddedResponseMessage{
		Id:    message.ID,
		Error: err,
	}
	return nil
}

func (x *NotifyBlockAddedResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "NotifyBlockAddedResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}
	return &appmessage.NotifyBlockAddedResponseMessage{
		ID:    x.Id,
		Error: rpcErr,
	}, nil
}

func (x *KaspadMessage_BlockAddedNotification) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_BlockAddedNotification is nil")
	}
	return x.BlockAddedNotification.toAppMessage()
}

func (x *KaspadMessage_BlockAddedNotification) fromAppMessage(message *appmessage.BlockAddedNotificationMessage) error {
	block := &RpcBlock{}
	err := block.fromAppMessage(message.Block)
	if err != nil {
		return err
	}
	x.BlockAddedNotification = &BlockAddedNotificationMessage{
		Id:    message.ID,
		Block: block,
	}
	return nil
}

func (x *BlockAddedNotificationMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "BlockAddedNotificationMessage is nil")
	}
	block, err := x.Block.toAppMessage()
	if err != nil {
		return nil, err
	}
	return &appmessage.BlockAddedNotificationMessage{
		ID:    x.Id,
		Block: block,
	}, nil
}
