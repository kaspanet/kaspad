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
	return &appmessage.NotifyBlockAddedRequestMessage{}, nil
}

func (x *KaspadMessage_NotifyBlockAddedResponse) fromAppMessage(_ *appmessage.NotifyBlockAddedResponseMessage) error {
	return nil
}

func (x *KaspadMessage_BlockAddedNotification) toAppMessage() (appmessage.Message, error) {
	return &appmessage.BlockAddedNotificationMessage{}, nil
}

func (x *KaspadMessage_BlockAddedNotification) fromAppMessage(_ *appmessage.BlockAddedNotificationMessage) error {
	return nil
}
