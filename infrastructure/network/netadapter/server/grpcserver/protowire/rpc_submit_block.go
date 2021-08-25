package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
	"math"
)

func (x *KaspadMessage_SubmitBlockRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "SubmitBlockRequestMessage is nil")
	}
	return x.SubmitBlockRequest.toAppMessage()
}

func (x *KaspadMessage_SubmitBlockRequest) fromAppMessage(message *appmessage.SubmitBlockRequestMessage) error {
	x.SubmitBlockRequest = &SubmitBlockRequestMessage{Block: &RpcBlock{}}
	return x.SubmitBlockRequest.Block.fromAppMessage(message.Block)
}

func (x *SubmitBlockRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "SubmitBlockRequestMessage is nil")
	}
	blockAppMessage, err := x.Block.toAppMessage()
	if err != nil {
		return nil, err
	}
	return &appmessage.SubmitBlockRequestMessage{
		Block: blockAppMessage,
	}, nil
}

func (x *KaspadMessage_SubmitBlockResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_SubmitBlockResponse is nil")
	}
	return x.SubmitBlockResponse.toAppMessage()
}

func (x *KaspadMessage_SubmitBlockResponse) fromAppMessage(message *appmessage.SubmitBlockResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.SubmitBlockResponse = &SubmitBlockResponseMessage{
		RejectReason: SubmitBlockResponseMessage_RejectReason(message.RejectReason),
		Error:        err,
	}
	return nil
}

func (x *SubmitBlockResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "SubmitBlockResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}
	return &appmessage.SubmitBlockResponseMessage{
		RejectReason: appmessage.RejectReason(x.RejectReason),
		Error:        rpcErr,
	}, nil
}

func (x *RpcBlock) toAppMessage() (*appmessage.RPCBlock, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RpcBlock is nil")
	}
	header, err := x.Header.toAppMessage()
	if err != nil {
		return nil, err
	}
	transactions := make([]*appmessage.RPCTransaction, len(x.Transactions))
	for i, transaction := range x.Transactions {
		appTransaction, err := transaction.toAppMessage()
		if err != nil {
			return nil, err
		}
		transactions[i] = appTransaction
	}
	var verboseData *appmessage.RPCBlockVerboseData
	if x.VerboseData != nil {
		appMessageVerboseData, err := x.VerboseData.toAppMessage()
		if err != nil {
			return nil, err
		}
		verboseData = appMessageVerboseData
	}
	return &appmessage.RPCBlock{
		Header:       header,
		Transactions: transactions,
		VerboseData:  verboseData,
	}, nil
}

func (x *RpcBlock) fromAppMessage(message *appmessage.RPCBlock) error {
	header := &RpcBlockHeader{}
	header.fromAppMessage(message.Header)
	transactions := make([]*RpcTransaction, len(message.Transactions))
	for i, transaction := range message.Transactions {
		rpcTransaction := &RpcTransaction{}
		rpcTransaction.fromAppMessage(transaction)
		transactions[i] = rpcTransaction
	}
	var verboseData *RpcBlockVerboseData
	if message.VerboseData != nil {
		verboseData = &RpcBlockVerboseData{}
		verboseData.fromAppMessage(message.VerboseData)
	}
	*x = RpcBlock{
		Header:       header,
		Transactions: transactions,
		VerboseData:  verboseData,
	}
	return nil
}

func (x *RpcBlockHeader) toAppMessage() (*appmessage.RPCBlockHeader, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RpcBlockHeader is nil")
	}
	if x.Version > math.MaxUint16 {
		return nil, errors.Errorf("Invalid block header version - bigger then uint16")
	}
	return &appmessage.RPCBlockHeader{
		Version:              x.Version,
		ParentHashes:         x.ParentHashes,
		HashMerkleRoot:       x.HashMerkleRoot,
		AcceptedIDMerkleRoot: x.AcceptedIdMerkleRoot,
		UTXOCommitment:       x.UtxoCommitment,
		Timestamp:            x.Timestamp,
		Bits:                 x.Bits,
		Nonce:                x.Nonce,
		DAAScore:             x.DaaScore,
		BlueWork:             x.BlueWork,
		BlueScore:            x.BlueScore,
		PruningPoint:         x.PruningPoint,
	}, nil
}

func (x *RpcBlockHeader) fromAppMessage(message *appmessage.RPCBlockHeader) {
	*x = RpcBlockHeader{
		Version:              message.Version,
		ParentHashes:         message.ParentHashes,
		HashMerkleRoot:       message.HashMerkleRoot,
		AcceptedIdMerkleRoot: message.AcceptedIDMerkleRoot,
		UtxoCommitment:       message.UTXOCommitment,
		Timestamp:            message.Timestamp,
		Bits:                 message.Bits,
		Nonce:                message.Nonce,
		DaaScore:             message.DAAScore,
		BlueWork:             message.BlueWork,
		BlueScore:            message.BlueScore,
		PruningPoint:         message.PruningPoint,
	}
}

func (x *RpcBlockVerboseData) toAppMessage() (*appmessage.RPCBlockVerboseData, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RpcBlockVerboseData is nil")
	}
	return &appmessage.RPCBlockVerboseData{
		Hash:               x.Hash,
		Difficulty:         x.Difficulty,
		SelectedParentHash: x.SelectedParentHash,
		TransactionIDs:     x.TransactionIds,
		IsHeaderOnly:       x.IsHeaderOnly,
		BlueScore:          x.BlueScore,
		ChildrenHashes:     x.ChildrenHashes,
	}, nil
}

func (x *RpcBlockVerboseData) fromAppMessage(message *appmessage.RPCBlockVerboseData) {
	*x = RpcBlockVerboseData{
		Hash:               message.Hash,
		Difficulty:         message.Difficulty,
		SelectedParentHash: message.SelectedParentHash,
		TransactionIds:     message.TransactionIDs,
		IsHeaderOnly:       message.IsHeaderOnly,
		BlueScore:          message.BlueScore,
		ChildrenHashes:     message.ChildrenHashes,
	}
}
