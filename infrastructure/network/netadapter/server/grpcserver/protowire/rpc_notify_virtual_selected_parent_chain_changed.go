package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_NotifyVirtualSelectedParentChainChangedRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_NotifyVirtualSelectedParentChainChangedRequest is nil")
	}
	return &appmessage.NotifyVirtualSelectedParentChainChangedRequestMessage{}, nil
}

func (x *KaspadMessage_NotifyVirtualSelectedParentChainChangedRequest) fromAppMessage(_ *appmessage.NotifyVirtualSelectedParentChainChangedRequestMessage) error {
	x.NotifyVirtualSelectedParentChainChangedRequest = &NotifyVirtualSelectedParentChainChangedRequestMessage{}
	return nil
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
	addedChainBlocks := make([]*ChainBlock, len(message.AddedChainBlocks))
	for i, addedChainBlock := range message.AddedChainBlocks {
		protoAddedChainBlock := &ChainBlock{}
		err := protoAddedChainBlock.fromAppMessage(addedChainBlock)
		if err != nil {
			return err
		}
		addedChainBlocks[i] = protoAddedChainBlock
	}
	x.VirtualSelectedParentChainChangedNotification = &VirtualSelectedParentChainChangedNotificationMessage{
		RemovedChainBlockHashes: message.RemovedChainBlockHashes,
		AddedChainBlocks:        addedChainBlocks,
	}
	return nil
}

func (x *VirtualSelectedParentChainChangedNotificationMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "VirtualSelectedParentChainChangedNotificationMessage is nil")
	}
	addedChainBlocks := make([]*appmessage.ChainBlock, len(x.AddedChainBlocks))
	for i, addedChainBlock := range x.AddedChainBlocks {
		appAddedChainBlock, err := addedChainBlock.toAppMessage()
		if err != nil {
			return nil, err
		}
		addedChainBlocks[i] = appAddedChainBlock
	}
	return &appmessage.VirtualSelectedParentChainChangedNotificationMessage{
		RemovedChainBlockHashes: x.RemovedChainBlockHashes,
		AddedChainBlocks:        addedChainBlocks,
	}, nil
}

func (x *ChainBlock) toAppMessage() (*appmessage.ChainBlock, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "ChainBlock is nil")
	}
	acceptedBlocks := make([]*appmessage.AcceptedBlock, len(x.AcceptedBlocks))
	for j, acceptedBlock := range x.AcceptedBlocks {
		appAcceptedBlock, err := acceptedBlock.toAppMessage()
		if err != nil {
			return nil, err
		}
		acceptedBlocks[j] = appAcceptedBlock
	}
	return &appmessage.ChainBlock{
		Hash:           x.Hash,
		AcceptedBlocks: acceptedBlocks,
	}, nil
}

func (x *ChainBlock) fromAppMessage(message *appmessage.ChainBlock) error {
	acceptedBlocks := make([]*AcceptedBlock, len(message.AcceptedBlocks))
	for j, acceptedBlock := range message.AcceptedBlocks {
		acceptedBlocks[j] = &AcceptedBlock{
			Hash:                   acceptedBlock.Hash,
			AcceptedTransactionIds: acceptedBlock.AcceptedTransactionIDs,
		}
	}
	*x = ChainBlock{
		Hash:           message.Hash,
		AcceptedBlocks: acceptedBlocks,
	}
	return nil
}

func (x *AcceptedBlock) toAppMessage() (*appmessage.AcceptedBlock, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "AcceptedBlock is nil")
	}
	return &appmessage.AcceptedBlock{
		Hash:                   x.Hash,
		AcceptedTransactionIDs: x.AcceptedTransactionIds,
	}, nil
}
