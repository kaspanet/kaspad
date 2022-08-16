package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_NotifyFinalityConflictsRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_NotifyFinalityConflictsRequest is nil")
	}
	return x.NotifyFinalityConflictsRequest.toAppMessage()
}

func (x *KaspadMessage_NotifyFinalityConflictsRequest) fromAppMessage(message *appmessage.NotifyFinalityConflictsRequestMessage) error {
	x.NotifyFinalityConflictsRequest = &NotifyFinalityConflictsRequestMessage{Id: message.ID}
	return nil
}

func (x *NotifyFinalityConflictsRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "NotifyFinalityConflictsRequestMessage is nil")
	}
	return &appmessage.NotifyFinalityConflictsRequestMessage{
		ID: x.Id,
	}, nil
}

func (x *KaspadMessage_NotifyFinalityConflictsResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_NotifyFinalityConflictsResponse is nil")
	}
	return x.NotifyFinalityConflictsResponse.toAppMessage()
}

func (x *KaspadMessage_NotifyFinalityConflictsResponse) fromAppMessage(message *appmessage.NotifyFinalityConflictsResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.NotifyFinalityConflictsResponse = &NotifyFinalityConflictsResponseMessage{
		Id:    message.ID,
		Error: err,
	}
	return nil
}

func (x *NotifyFinalityConflictsResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "NotifyFinalityConflictsResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}
	return &appmessage.NotifyFinalityConflictsResponseMessage{
		ID:    x.Id,
		Error: rpcErr,
	}, nil
}

func (x *KaspadMessage_FinalityConflictNotification) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_FinalityConflictNotification is nil")
	}
	return x.FinalityConflictNotification.toAppMessage()
}

func (x *KaspadMessage_FinalityConflictNotification) fromAppMessage(message *appmessage.FinalityConflictNotificationMessage) error {
	x.FinalityConflictNotification = &FinalityConflictNotificationMessage{
		Id:                 message.ID,
		ViolatingBlockHash: message.ViolatingBlockHash,
	}
	return nil
}

func (x *FinalityConflictNotificationMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "FinalityConflictNotificationMessage is nil")
	}
	return &appmessage.FinalityConflictNotificationMessage{
		ID:                 x.Id,
		ViolatingBlockHash: x.ViolatingBlockHash,
	}, nil
}

func (x *KaspadMessage_FinalityConflictResolvedNotification) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_FinalityConflictResolvedNotification is nil")
	}
	return x.FinalityConflictResolvedNotification.toAppMessage()
}

func (x *KaspadMessage_FinalityConflictResolvedNotification) fromAppMessage(message *appmessage.FinalityConflictResolvedNotificationMessage) error {
	x.FinalityConflictResolvedNotification = &FinalityConflictResolvedNotificationMessage{
		Id:                message.ID,
		FinalityBlockHash: message.FinalityBlockHash,
	}
	return nil
}

func (x *FinalityConflictResolvedNotificationMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "FinalityConflictResolvedNotificationMessage is nil")
	}
	return &appmessage.FinalityConflictResolvedNotificationMessage{
		ID:                x.Id,
		FinalityBlockHash: x.FinalityBlockHash,
	}, nil
}
