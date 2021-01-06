package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_NotifyFinalityConflictsRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.NotifyFinalityConflictsRequestMessage{}, nil
}

func (x *KaspadMessage_NotifyFinalityConflictsRequest) fromAppMessage(_ *appmessage.NotifyFinalityConflictsRequestMessage) error {
	x.NotifyFinalityConflictsRequest = &NotifyFinalityConflictsRequestMessage{}
	return nil
}

func (x *KaspadMessage_NotifyFinalityConflictsResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.NotifyFinalityConflictsResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.NotifyFinalityConflictsResponse.Error.Message}
	}
	return &appmessage.NotifyFinalityConflictsResponseMessage{
		Error: err,
	}, nil
}

func (x *KaspadMessage_NotifyFinalityConflictsResponse) fromAppMessage(message *appmessage.NotifyFinalityConflictsResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.NotifyFinalityConflictsResponse = &NotifyFinalityConflictsResponseMessage{
		Error: err,
	}
	return nil
}

func (x *KaspadMessage_FinalityConflictNotification) toAppMessage() (appmessage.Message, error) {
	return &appmessage.FinalityConflictNotificationMessage{
		ViolatingBlockHash: x.FinalityConflictNotification.ViolatingBlockHash,
	}, nil
}

func (x *KaspadMessage_FinalityConflictNotification) fromAppMessage(message *appmessage.FinalityConflictNotificationMessage) error {
	x.FinalityConflictNotification = &FinalityConflictNotificationMessage{
		ViolatingBlockHash: message.ViolatingBlockHash,
	}
	return nil
}

func (x *KaspadMessage_FinalityConflictResolvedNotification) toAppMessage() (appmessage.Message, error) {
	return &appmessage.FinalityConflictResolvedNotificationMessage{
		FinalityBlockHash: x.FinalityConflictResolvedNotification.FinalityBlockHash,
	}, nil
}

func (x *KaspadMessage_FinalityConflictResolvedNotification) fromAppMessage(message *appmessage.FinalityConflictResolvedNotificationMessage) error {
	x.FinalityConflictResolvedNotification = &FinalityConflictResolvedNotificationMessage{
		FinalityBlockHash: message.FinalityBlockHash,
	}
	return nil
}
