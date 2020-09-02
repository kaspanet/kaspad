package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_NotifyChainChangedRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.NotifyChainChangedRequestMessage{}, nil
}

func (x *KaspadMessage_NotifyChainChangedRequest) fromAppMessage(_ *appmessage.NotifyChainChangedRequestMessage) error {
	x.NotifyChainChangedRequest = &NotifyChainChangedRequestMessage{}
	return nil
}

func (x *KaspadMessage_NotifyChainChangedResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.NotifyChainChangedResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.NotifyChainChangedResponse.Error.Message}
	}
	return &appmessage.NotifyChainChangedResponseMessage{
		Error: err,
	}, nil
}

func (x *KaspadMessage_NotifyChainChangedResponse) fromAppMessage(message *appmessage.NotifyChainChangedResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.NotifyChainChangedResponse = &NotifyChainChangedResponseMessage{
		Error: err,
	}
	return nil
}

func (x *KaspadMessage_ChainChangedNotification) toAppMessage() (appmessage.Message, error) {
	return &appmessage.ChainChangedNotificationMessage{
		// TODO
	}, nil
}

func (x *KaspadMessage_ChainChangedNotification) fromAppMessage(message *appmessage.ChainChangedNotificationMessage) error {
	x.ChainChangedNotification = &ChainChangedNotificationMessage{
		// TODO
	}
	return nil
}
