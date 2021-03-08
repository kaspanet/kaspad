package protowire

import (
	"math"

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
	var blockVerboseData *appmessage.BlockVerboseData
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

func (x *BlockVerboseData) toAppMessage() (*appmessage.BlockVerboseData, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "BlockVerboseData is nil")
	}
	transactionVerboseData := make([]*appmessage.TransactionVerboseData, len(x.TransactionVerboseData))
	for i, transactionVerboseDatum := range x.TransactionVerboseData {
		appTransactionVerboseDatum, err := transactionVerboseDatum.toAppMessage()
		if err != nil {
			return nil, err
		}
		transactionVerboseData[i] = appTransactionVerboseDatum
	}

	if x.Version > math.MaxUint16 {
		return nil, errors.Errorf("Invalid block header version - bigger then uint16")
	}

	return &appmessage.BlockVerboseData{
		Hash:                   x.Hash,
		Version:                uint16(x.Version),
		VersionHex:             x.VersionHex,
		HashMerkleRoot:         x.HashMerkleRoot,
		AcceptedIDMerkleRoot:   x.AcceptedIDMerkleRoot,
		UTXOCommitment:         x.UtxoCommitment,
		TxIDs:                  x.TransactionIDs,
		TransactionVerboseData: transactionVerboseData,
		Time:                   x.Time,
		Nonce:                  x.Nonce,
		Bits:                   x.Bits,
		Difficulty:             x.Difficulty,
		ParentHashes:           x.ParentHashes,
		ChildrenHashes:         x.ChildrenHashes,
		SelectedParentHash:     x.SelectedParentHash,
		IsHeaderOnly:           x.IsHeaderOnly,
		BlueScore:              x.BlueScore,
	}, nil
}

func (x *BlockVerboseData) fromAppMessage(message *appmessage.BlockVerboseData) error {
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
		Version:                uint32(message.Version),
		VersionHex:             message.VersionHex,
		HashMerkleRoot:         message.HashMerkleRoot,
		AcceptedIDMerkleRoot:   message.AcceptedIDMerkleRoot,
		UtxoCommitment:         message.UTXOCommitment,
		TransactionIDs:         message.TxIDs,
		TransactionVerboseData: transactionVerboseData,
		Time:                   message.Time,
		Nonce:                  message.Nonce,
		Bits:                   message.Bits,
		Difficulty:             message.Difficulty,
		ParentHashes:           message.ParentHashes,
		ChildrenHashes:         message.ChildrenHashes,
		SelectedParentHash:     message.SelectedParentHash,
		IsHeaderOnly:           message.IsHeaderOnly,
		BlueScore:              message.BlueScore,
	}
	return nil
}

func (x *TransactionVerboseData) toAppMessage() (*appmessage.TransactionVerboseData, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "TransactionVerboseData is nil")
	}
	inputs := make([]*appmessage.TransactionVerboseInput, len(x.TransactionVerboseInputs))
	for j, item := range x.TransactionVerboseInputs {
		input, err := item.toAppMessage()
		if err != nil {
			return nil, err
		}
		inputs[j] = input
	}
	outputs := make([]*appmessage.TransactionVerboseOutput, len(x.TransactionVerboseOutputs))
	for j, item := range x.TransactionVerboseOutputs {
		output, err := item.toAppMessage()
		if err != nil {
			return nil, err
		}
		outputs[j] = output
	}
	if x.Version > math.MaxUint16 {
		return nil, errors.Errorf("Invalid transaction version - bigger then uint16")
	}
	return &appmessage.TransactionVerboseData{
		TxID:                      x.TxId,
		Hash:                      x.Hash,
		Size:                      x.Size,
		Version:                   uint16(x.Version),
		LockTime:                  x.LockTime,
		SubnetworkID:              x.SubnetworkId,
		Gas:                       x.Gas,
		Payload:                   x.Payload,
		TransactionVerboseInputs:  inputs,
		TransactionVerboseOutputs: outputs,
		BlockHash:                 x.BlockHash,
		Time:                      x.Time,
		BlockTime:                 x.BlockTime,
	}, nil
}

func (x *TransactionVerboseData) fromAppMessage(message *appmessage.TransactionVerboseData) error {
	inputs := make([]*TransactionVerboseInput, len(message.TransactionVerboseInputs))
	for j, item := range message.TransactionVerboseInputs {
		scriptSig := &ScriptSig{
			Asm: item.ScriptSig.Asm,
			Hex: item.ScriptSig.Hex,
		}
		inputs[j] = &TransactionVerboseInput{
			TxId:        item.TxID,
			OutputIndex: item.OutputIndex,
			ScriptSig:   scriptSig,
			Sequence:    item.Sequence,
		}
	}
	outputs := make([]*TransactionVerboseOutput, len(message.TransactionVerboseOutputs))
	for j, item := range message.TransactionVerboseOutputs {
		scriptPubKey := &ScriptPublicKeyResult{
			Hex:     item.ScriptPubKey.Hex,
			Type:    item.ScriptPubKey.Type,
			Address: item.ScriptPubKey.Address,
			Version: uint32(item.ScriptPubKey.Version),
		}
		outputs[j] = &TransactionVerboseOutput{
			Value:           item.Value,
			Index:           item.Index,
			ScriptPublicKey: scriptPubKey,
		}
	}
	*x = TransactionVerboseData{
		TxId:                      message.TxID,
		Hash:                      message.Hash,
		Size:                      message.Size,
		Version:                   uint32(message.Version),
		LockTime:                  message.LockTime,
		SubnetworkId:              message.SubnetworkID,
		Gas:                       message.Gas,
		Payload:                   message.Payload,
		TransactionVerboseInputs:  inputs,
		TransactionVerboseOutputs: outputs,
		BlockHash:                 message.BlockHash,
		Time:                      message.Time,
		BlockTime:                 message.BlockTime,
	}
	return nil
}

func (x *TransactionVerboseInput) toAppMessage() (*appmessage.TransactionVerboseInput, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "TransactionVerboseInput is nil")
	}
	scriptSig, err := x.ScriptSig.toAppMessage()
	if err != nil {
		return nil, err
	}
	return &appmessage.TransactionVerboseInput{
		TxID:        x.TxId,
		OutputIndex: x.OutputIndex,
		ScriptSig:   scriptSig,
		Sequence:    x.Sequence,
	}, nil
}

func (x *TransactionVerboseOutput) toAppMessage() (*appmessage.TransactionVerboseOutput, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "TransactionVerboseOutput is nil")
	}
	scriptPubKey, err := x.ScriptPublicKey.toAppMessage()
	if err != nil {
		return nil, err
	}
	return &appmessage.TransactionVerboseOutput{
		Value:        x.Value,
		Index:        x.Index,
		ScriptPubKey: scriptPubKey,
	}, nil
}

func (x *ScriptSig) toAppMessage() (*appmessage.ScriptSig, error) {
	if x == nil {
		return nil, errors.Wrap(errorNil, "ScriptSig is nil")
	}
	return &appmessage.ScriptSig{
		Asm: x.Asm,
		Hex: x.Hex,
	}, nil
}

func (x *ScriptPublicKeyResult) toAppMessage() (*appmessage.ScriptPubKeyResult, error) {
	if x == nil {
		return nil, errors.Wrap(errorNil, "ScriptPublicKeyResult is nil")
	}
	return &appmessage.ScriptPubKeyResult{
		Hex:     x.Hex,
		Type:    x.Type,
		Address: x.Address,
		Version: uint16(x.Version),
	}, nil

}
