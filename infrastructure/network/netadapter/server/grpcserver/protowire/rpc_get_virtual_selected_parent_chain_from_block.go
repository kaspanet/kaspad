package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetVirtualSelectedParentChainFromBlockRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetVirtualSelectedParentChainFromBlockRequest is nil")
	}
	return x.GetVirtualSelectedParentChainFromBlockRequest.toAppMessage()
}

func (x *KaspadMessage_GetVirtualSelectedParentChainFromBlockRequest) fromAppMessage(message *appmessage.GetVirtualSelectedParentChainFromBlockRequestMessage) error {
	x.GetVirtualSelectedParentChainFromBlockRequest = &GetVirtualSelectedParentChainFromBlockRequestMessage{
		StartHash: message.StartHash,
	}
	return nil
}

func (x *GetVirtualSelectedParentChainFromBlockRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetVirtualSelectedParentChainFromBlockRequestMessage is nil")
	}
	return &appmessage.GetVirtualSelectedParentChainFromBlockRequestMessage{
		StartHash: x.StartHash,
	}, nil
}

func (x *KaspadMessage_GetVirtualSelectedParentChainFromBlockResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetVirtualSelectedParentChainFromBlockResponse is nil")
	}
	return x.GetVirtualSelectedParentChainFromBlockResponse.toAppMessage()
}

func (x *KaspadMessage_GetVirtualSelectedParentChainFromBlockResponse) fromAppMessage(message *appmessage.GetVirtualSelectedParentChainFromBlockResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.GetVirtualSelectedParentChainFromBlockResponse = &GetVirtualSelectedParentChainFromBlockResponseMessage{
		RemovedChainBlockHashes: message.RemovedChainBlockHashes,
		AddedChainBlockHashes:   message.AddedChainBlockHashes,
		Error:                   err,
	}
	return nil
}

func (x *GetVirtualSelectedParentChainFromBlockResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetVirtualSelectedParentChainFromBlockResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}

	if rpcErr != nil && (len(x.AddedChainBlockHashes) != 0 || len(x.RemovedChainBlockHashes) != 0) {
		return nil, errors.New("GetVirtualSelectedParentChainFromBlockResponseMessage contains both an error and a response")
	}

	return &appmessage.GetVirtualSelectedParentChainFromBlockResponseMessage{
		RemovedChainBlockHashes: x.RemovedChainBlockHashes,
		AddedChainBlockHashes:   x.AddedChainBlockHashes,
		Error:                   rpcErr,
	}, nil
}
