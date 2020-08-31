package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_NotifyBlockAddedRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.NotifyBlockAddedRequestMessage{}, nil
}

func (x *KaspadMessage_NotifyBlockAddedRequest) fromAppMessage(_ *appmessage.NotifyBlockAddedRequestMessage) error {
	x.NotifyBlockAddedRequest = &NotifyBlockAddedRequestMessage{}
	return nil
}

func (x *KaspadMessage_NotifyBlockAddedResponse) toAppMessage() (appmessage.Message, error) {
	return &appmessage.NotifyBlockAddedResponseMessage{}, nil
}

func (x *KaspadMessage_NotifyBlockAddedResponse) fromAppMessage(_ *appmessage.NotifyBlockAddedResponseMessage) error {
	return nil
}

func (x *KaspadMessage_BlockAddedNotification) toAppMessage() (appmessage.Message, error) {
	block, err := x.BlockAddedNotification.Block.toAppMessage()
	if err != nil {
		return nil, err
	}
	return &appmessage.BlockAddedNotificationMessage{
		Block: block.(*appmessage.MsgBlock),
	}, nil
}

func (x *KaspadMessage_BlockAddedNotification) fromAppMessage(message *appmessage.BlockAddedNotificationMessage) error {
	blockMessage := &BlockMessage{}
	err := blockMessage.fromAppMessage(message.Block)
	if err != nil {
		return err
	}
	x.BlockAddedNotification = &BlockAddedNotificationMessage{
		Block: blockMessage,
	}
	return nil
}
