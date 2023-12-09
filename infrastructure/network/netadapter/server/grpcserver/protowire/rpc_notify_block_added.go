package protowire

import (
	"github.com/pkg/errors"
	"github.com/zoomy-network/zoomyd/app/appmessage"
)

func (x *KaspadMessage_NotifyBlockAddedRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.NotifyBlockAddedRequestMessage{}, nil
}

func (x *KaspadMessage_NotifyBlockAddedRequest) fromAppMessage(_ *appmessage.NotifyBlockAddedRequestMessage) error {
	x.NotifyBlockAddedRequest = &NotifyBlockAddedRequestMessage{}
	return nil
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
		Block: block,
	}, nil
}
