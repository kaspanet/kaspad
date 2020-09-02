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
	addedChainBlocks := make([]*appmessage.ChainChangedChainBlock, len(x.ChainChangedNotification.AddedChainBlocks))
	for i, addedChainBlock := range x.ChainChangedNotification.AddedChainBlocks {
		acceptedBlocks := make([]*appmessage.ChainChangedAcceptedBlock, len(addedChainBlock.AcceptedBlocks))
		for j, acceptedBlock := range addedChainBlock.AcceptedBlocks {
			acceptedBlocks[j] = &appmessage.ChainChangedAcceptedBlock{
				Hash:          acceptedBlock.Hash,
				AcceptedTxIds: acceptedBlock.AcceptedTxIds,
			}
		}
		addedChainBlocks[i] = &appmessage.ChainChangedChainBlock{
			Hash:           addedChainBlock.Hash,
			AcceptedBlocks: acceptedBlocks,
		}
	}
	return &appmessage.ChainChangedNotificationMessage{
		RemovedChainBlockHashes: x.ChainChangedNotification.RemovedChainBlockHashes,
		AddedChainBlocks:        addedChainBlocks,
	}, nil
}

func (x *KaspadMessage_ChainChangedNotification) fromAppMessage(message *appmessage.ChainChangedNotificationMessage) error {
	addedChainBlocks := make([]*ChainChangedChainBlock, len(message.AddedChainBlocks))
	for i, addedChainBlock := range message.AddedChainBlocks {
		acceptedBlocks := make([]*ChainChangedAcceptedBlock, len(addedChainBlock.AcceptedBlocks))
		for j, acceptedBlock := range addedChainBlock.AcceptedBlocks {
			acceptedBlocks[j] = &ChainChangedAcceptedBlock{
				Hash:          acceptedBlock.Hash,
				AcceptedTxIds: acceptedBlock.AcceptedTxIds,
			}
		}
		addedChainBlocks[i] = &ChainChangedChainBlock{
			Hash:           addedChainBlock.Hash,
			AcceptedBlocks: acceptedBlocks,
		}
	}
	x.ChainChangedNotification = &ChainChangedNotificationMessage{
		RemovedChainBlockHashes: message.RemovedChainBlockHashes,
		AddedChainBlocks:        addedChainBlocks,
	}
	return nil
}
