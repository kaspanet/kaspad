package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_NotifyVirtualSelectedParentChainChangedRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_NotifyVirtualSelectedParentChainChangedRequest is nil")
	}
	return x.NotifyVirtualSelectedParentChainChangedRequest.toAppMessage()
}

func (x *KaspadMessage_NotifyVirtualSelectedParentChainChangedRequest) fromAppMessage(message *appmessage.NotifyVirtualSelectedParentChainChangedRequestMessage) error {
	x.NotifyVirtualSelectedParentChainChangedRequest = &NotifyVirtualSelectedParentChainChangedRequestMessage{
		Id:                            message.ID,
		IncludeAcceptedTransactionIds: message.IncludeAcceptedTransactionIDs,
	}
	return nil
}

func (x *NotifyVirtualSelectedParentChainChangedRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "NotifyVirtualSelectedParentChainChangedRequestMessage is nil")
	}
	return &appmessage.NotifyVirtualSelectedParentChainChangedRequestMessage{ID: x.Id}, nil
}

func (x *KaspadMessage_NotifyVirtualSelectedParentChainChangedResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_NotifyVirtualSelectedParentChainChangedResponse is nil")
	}
	return x.NotifyVirtualSelectedParentChainChangedResponse.toAppMessage()
}

func (x *KaspadMessage_NotifyVirtualSelectedParentChainChangedResponse) fromAppMessage(message *appmessage.NotifyVirtualSelectedParentChainChangedResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.NotifyVirtualSelectedParentChainChangedResponse = &NotifyVirtualSelectedParentChainChangedResponseMessage{
		Id:    message.ID,
		Error: err,
	}
	return nil
}

func (x *NotifyVirtualSelectedParentChainChangedResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "NotifyVirtualSelectedParentChainChangedResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}
	return &appmessage.NotifyVirtualSelectedParentChainChangedResponseMessage{
		ID:    x.Id,
		Error: rpcErr,
	}, nil
}

func (x *KaspadMessage_VirtualSelectedParentChainChangedNotification) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_VirtualSelectedParentChainChangedNotification is nil")
	}
	return x.VirtualSelectedParentChainChangedNotification.toAppMessage()
}

func (x *KaspadMessage_VirtualSelectedParentChainChangedNotification) fromAppMessage(message *appmessage.VirtualSelectedParentChainChangedNotificationMessage) error {
	x.VirtualSelectedParentChainChangedNotification = &VirtualSelectedParentChainChangedNotificationMessage{
		Id:                      message.ID,
		RemovedChainBlockHashes: message.RemovedChainBlockHashes,
		AddedChainBlockHashes:   message.AddedChainBlockHashes,
		AcceptedTransactionIds:  make([]*AcceptedTransactionIds, len(message.AcceptedTransactionIDs)),
	}

	for i, acceptedTransactionIDs := range message.AcceptedTransactionIDs {
		x.VirtualSelectedParentChainChangedNotification.AcceptedTransactionIds[i] = &AcceptedTransactionIds{}
		x.VirtualSelectedParentChainChangedNotification.AcceptedTransactionIds[i].fromAppMessage(acceptedTransactionIDs)
	}
	return nil
}

func (x *VirtualSelectedParentChainChangedNotificationMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "VirtualSelectedParentChainChangedNotificationMessage is nil")
	}
	message := &appmessage.VirtualSelectedParentChainChangedNotificationMessage{
		ID:                      x.Id,
		RemovedChainBlockHashes: x.RemovedChainBlockHashes,
		AddedChainBlockHashes:   x.AddedChainBlockHashes,
		AcceptedTransactionIDs:  make([]*appmessage.AcceptedTransactionIDs, len(x.AcceptedTransactionIds)),
	}

	for i, acceptedTransactionIds := range x.AcceptedTransactionIds {
		message.AcceptedTransactionIDs[i] = acceptedTransactionIds.toAppMessage()
	}
	return message, nil
}
