package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_NotifyVirtualSelectedParentBlueScoreChangedRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.NotifyVirtualSelectedParentBlueScoreChangedRequestMessage{}, nil
}

func (x *KaspadMessage_NotifyVirtualSelectedParentBlueScoreChangedRequest) fromAppMessage(_ *appmessage.NotifyVirtualSelectedParentBlueScoreChangedRequestMessage) error {
	x.NotifyVirtualSelectedParentBlueScoreChangedRequest = &NotifyVirtualSelectedParentBlueScoreChangedRequestMessage{}
	return nil
}

func (x *KaspadMessage_NotifyVirtualSelectedParentBlueScoreChangedResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.NotifyVirtualSelectedParentBlueScoreChangedResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.NotifyVirtualSelectedParentBlueScoreChangedResponse.Error.Message}
	}
	return &appmessage.NotifyVirtualSelectedParentBlueScoreChangedResponseMessage{
		Error: err,
	}, nil
}

func (x *KaspadMessage_NotifyVirtualSelectedParentBlueScoreChangedResponse) fromAppMessage(message *appmessage.NotifyVirtualSelectedParentBlueScoreChangedResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.NotifyVirtualSelectedParentBlueScoreChangedResponse = &NotifyVirtualSelectedParentBlueScoreChangedResponseMessage{
		Error: err,
	}
	return nil
}

func (x *KaspadMessage_VirtualSelectedParentBlueScoreChangedNotification) toAppMessage() (appmessage.Message, error) {
	return &appmessage.VirtualSelectedParentBlueScoreChangedNotificationMessage{
		VirtualSelectedParentBlueScore: x.VirtualSelectedParentBlueScoreChangedNotification.VirtualSelectedParentBlueScore,
	}, nil
}

func (x *KaspadMessage_VirtualSelectedParentBlueScoreChangedNotification) fromAppMessage(message *appmessage.VirtualSelectedParentBlueScoreChangedNotificationMessage) error {
	x.VirtualSelectedParentBlueScoreChangedNotification = &VirtualSelectedParentBlueScoreChangedNotificationMessage{
		VirtualSelectedParentBlueScore: message.VirtualSelectedParentBlueScore,
	}
	return nil
}
