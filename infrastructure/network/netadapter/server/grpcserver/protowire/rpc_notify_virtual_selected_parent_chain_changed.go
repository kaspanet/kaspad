package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_NotifyVirtualSelectedParentChainChangedRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.NotifyVirtualSelectedParentChainChangedRequestMessage{}, nil
}

func (x *KaspadMessage_NotifyVirtualSelectedParentChainChangedRequest) fromAppMessage(_ *appmessage.NotifyVirtualSelectedParentChainChangedRequestMessage) error {
	x.NotifyVirtualSelectedParentChainChangedRequest = &NotifyVirtualSelectedParentChainChangedRequestMessage{}
	return nil
}

func (x *KaspadMessage_NotifyVirtualSelectedParentChainChangedResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.NotifyVirtualSelectedParentChainChangedResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.NotifyVirtualSelectedParentChainChangedResponse.Error.Message}
	}
	return &appmessage.NotifyVirtualSelectedParentChainChangedResponseMessage{
		Error: err,
	}, nil
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

func (x *KaspadMessage_VirtualSelectedParentChainChangedNotification) toAppMessage() (appmessage.Message, error) {
	addedChainBlocks := make([]*appmessage.ChainBlock, len(x.VirtualSelectedParentChainChangedNotification.AddedChainBlocks))
	for i, addedChainBlock := range x.VirtualSelectedParentChainChangedNotification.AddedChainBlocks {
		appAddedChainBlock, err := addedChainBlock.toAppMessage()
		if err != nil {
			return nil, err
		}
		addedChainBlocks[i] = appAddedChainBlock
	}
	return &appmessage.VirtualSelectedParentChainChangedNotificationMessage{
		RemovedChainBlockHashes: x.VirtualSelectedParentChainChangedNotification.RemovedChainBlockHashes,
		AddedChainBlocks:        addedChainBlocks,
	}, nil
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

func (x *ChainBlock) toAppMessage() (*appmessage.ChainBlock, error) {
	acceptedBlocks := make([]*appmessage.AcceptedBlock, len(x.AcceptedBlocks))
	for j, acceptedBlock := range x.AcceptedBlocks {
		acceptedBlocks[j] = &appmessage.AcceptedBlock{
			Hash:                   acceptedBlock.Hash,
			AcceptedTransactionIDs: acceptedBlock.AcceptedTransactionIds,
		}
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
