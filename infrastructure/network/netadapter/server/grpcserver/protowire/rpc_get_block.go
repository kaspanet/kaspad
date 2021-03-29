package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetBlockRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetBlockRequest is nil")
	}
	return x.GetBlockRequest.toAppMessage()
}

func (x *GetBlockRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetBlockRequestMessage is nil")
	}
	return &appmessage.GetBlockRequestMessage{
		Hash:                          x.Hash,
		IncludeTransactionVerboseData: x.IncludeTransactionVerboseData,
	}, nil
}

func (x *KaspadMessage_GetBlockRequest) fromAppMessage(message *appmessage.GetBlockRequestMessage) error {
	x.GetBlockRequest = &GetBlockRequestMessage{
		Hash:                          message.Hash,
		IncludeTransactionVerboseData: message.IncludeTransactionVerboseData,
	}
	return nil
}

func (x *KaspadMessage_GetBlockResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetBlockResponse is nil")
	}
	return x.GetBlockResponse.toAppMessage()
}

func (x *GetBlockResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetBlockResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}
	var blockVerboseData *appmessage.RPCBlockVerboseData
	// Return verbose data only if there's no error
	if rpcErr != nil && x.BlockVerboseData != nil {
		return nil, errors.New("GetBlockResponseMessage contains both an error and a response")
	}
	if rpcErr == nil {
		blockVerboseData, err = x.BlockVerboseData.toAppMessage()
		if err != nil {
			return nil, err
		}
	}
	return &appmessage.GetBlockResponseMessage{
		BlockVerboseData: blockVerboseData,
		Error:            rpcErr,
	}, nil
}

func (x *KaspadMessage_GetBlockResponse) fromAppMessage(message *appmessage.GetBlockResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	var blockVerboseData *BlockVerboseData
	if message.BlockVerboseData != nil {
		protoBlockVerboseData := &BlockVerboseData{}
		err := protoBlockVerboseData.fromAppMessage(message.BlockVerboseData)
		if err != nil {
			return err
		}
		blockVerboseData = protoBlockVerboseData
	}
	x.GetBlockResponse = &GetBlockResponseMessage{
		BlockVerboseData: blockVerboseData,
		Error:            err,
	}
	return nil
}

func (x *BlockVerboseData) toAppMessage() (*appmessage.RPCBlockVerboseData, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RPCBlockVerboseData is nil")
	}
	block, err := x.Block.toAppMessage()
	if err != nil {
		return nil, err
	}
	transactionVerboseData := make([]*appmessage.RPCTransactionVerboseData, len(x.TransactionVerboseData))
	for i, transactionVerboseDatum := range x.TransactionVerboseData {
		appTransactionVerboseDatum, err := transactionVerboseDatum.toAppMessage()
		if err != nil {
			return nil, err
		}
		transactionVerboseData[i] = appTransactionVerboseDatum
	}
	return &appmessage.RPCBlockVerboseData{
		Hash:                   x.Hash,
		Block:                  block,
		TransactionIDs:         x.TransactionIDs,
		TransactionVerboseData: transactionVerboseData,
		Difficulty:             x.Difficulty,
		ChildrenHashes:         x.ChildrenHashes,
		SelectedParentHash:     x.SelectedParentHash,
		IsHeaderOnly:           x.IsHeaderOnly,
		BlueScore:              x.BlueScore,
	}, nil
}

func (x *BlockVerboseData) fromAppMessage(message *appmessage.RPCBlockVerboseData) error {
	block := &RpcBlock{}
	err := block.fromAppMessage(message.Block)
	if err != nil {
		return err
	}
	transactionVerboseData := make([]*TransactionVerboseData, len(message.TransactionVerboseData))
	for i, transactionVerboseDatum := range message.TransactionVerboseData {
		protoTransactionVerboseDatum := &TransactionVerboseData{}
		err := protoTransactionVerboseDatum.fromAppMessage(transactionVerboseDatum)
		if err != nil {
			return err
		}
		transactionVerboseData[i] = protoTransactionVerboseDatum
	}
	*x = BlockVerboseData{
		Hash:                   message.Hash,
		Block:                  block,
		TransactionIDs:         message.TransactionIDs,
		TransactionVerboseData: transactionVerboseData,
		Difficulty:             message.Difficulty,
		ChildrenHashes:         message.ChildrenHashes,
		SelectedParentHash:     message.SelectedParentHash,
		IsHeaderOnly:           message.IsHeaderOnly,
		BlueScore:              message.BlueScore,
	}
	return nil
}

func (x *TransactionVerboseData) toAppMessage() (*appmessage.RPCTransactionVerboseData, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RPCTransactionVerboseData is nil")
	}
	transaction, err := x.Transaction.toAppMessage()
	if err != nil {
		return nil, err
	}
	inputs := make([]*appmessage.RPCTransactionInputVerboseData, len(x.TransactionVerboseInputs))
	for j := range x.TransactionVerboseInputs {
		inputs[j] = &appmessage.RPCTransactionInputVerboseData{}
	}
	outputs := make([]*appmessage.RPCTransactionOutputVerboseData, len(x.TransactionVerboseOutputs))
	for j, item := range x.TransactionVerboseOutputs {
		outputs[j] = &appmessage.RPCTransactionOutputVerboseData{
			ScriptPublicKeyType:    item.ScriptPublicKeyType,
			ScriptPublicKeyAddress: item.ScriptPublicKeyAddress,
		}
	}
	return &appmessage.RPCTransactionVerboseData{
		TransactionID:             x.TxId,
		Hash:                      x.Hash,
		Size:                      x.Size,
		TransactionVerboseInputs:  inputs,
		TransactionVerboseOutputs: outputs,
		BlockHash:                 x.BlockHash,
		BlockTime:                 x.BlockTime,
		Transaction:               transaction,
	}, nil
}

func (x *TransactionVerboseData) fromAppMessage(message *appmessage.RPCTransactionVerboseData) error {
	transaction := &RpcTransaction{}
	transaction.fromAppMessage(message.Transaction)
	inputs := make([]*TransactionVerboseInput, len(message.TransactionVerboseInputs))
	for j := range message.TransactionVerboseInputs {
		inputs[j] = &TransactionVerboseInput{}
	}
	outputs := make([]*TransactionVerboseOutput, len(message.TransactionVerboseOutputs))
	for j, item := range message.TransactionVerboseOutputs {
		outputs[j] = &TransactionVerboseOutput{
			ScriptPublicKeyType:    item.ScriptPublicKeyType,
			ScriptPublicKeyAddress: item.ScriptPublicKeyAddress,
		}
	}
	*x = TransactionVerboseData{
		TxId:                      message.TransactionID,
		Hash:                      message.Hash,
		Size:                      message.Size,
		TransactionVerboseInputs:  inputs,
		TransactionVerboseOutputs: outputs,
		BlockHash:                 message.BlockHash,
		BlockTime:                 message.BlockTime,
		Transaction:               transaction,
	}
	return nil
}
