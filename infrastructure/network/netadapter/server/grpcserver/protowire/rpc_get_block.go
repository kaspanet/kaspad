package protowire

import (
	"math"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetBlockRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetBlockRequestMessage{
		Hash:                          x.GetBlockRequest.Hash,
		IncludeTransactionVerboseData: x.GetBlockRequest.IncludeTransactionVerboseData,
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
	var err *appmessage.RPCError
	if x.GetBlockResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.GetBlockResponse.Error.Message}
	}
	var blockVerboseData *appmessage.BlockVerboseData
	if x.GetBlockResponse.BlockVerboseData != nil {
		appBlockVerboseData, err := x.GetBlockResponse.BlockVerboseData.toAppMessage()
		if err != nil {
			return nil, err
		}
		blockVerboseData = appBlockVerboseData
	}
	return &appmessage.GetBlockResponseMessage{
		BlockVerboseData: blockVerboseData,
		Error:            err,
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
	inputs := make([]*appmessage.TransactionVerboseInput, len(x.TransactionVerboseInputs))
	for j, item := range x.TransactionVerboseInputs {
		scriptSig := &appmessage.ScriptSig{
			Asm: item.ScriptSig.Asm,
			Hex: item.ScriptSig.Hex,
		}
		inputs[j] = &appmessage.TransactionVerboseInput{
			TxID:        item.TxId,
			OutputIndex: item.OutputIndex,
			ScriptSig:   scriptSig,
			Sequence:    item.Sequence,
		}
	}
	outputs := make([]*appmessage.TransactionVerboseOutput, len(x.TransactionVerboseOutputs))
	for j, item := range x.TransactionVerboseOutputs {
		scriptPubKey := &appmessage.ScriptPubKeyResult{
			Hex:     item.ScriptPublicKey.Hex,
			Type:    item.ScriptPublicKey.Type,
			Address: item.ScriptPublicKey.Address,
			Version: uint16(item.ScriptPublicKey.Version),
		}
		outputs[j] = &appmessage.TransactionVerboseOutput{
			Value:        item.Value,
			Index:        item.Index,
			ScriptPubKey: scriptPubKey,
		}
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
		PayloadHash:               x.PayloadHash,
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
		PayloadHash:               message.PayloadHash,
		Payload:                   message.Payload,
		TransactionVerboseInputs:  inputs,
		TransactionVerboseOutputs: outputs,
		BlockHash:                 message.BlockHash,
		Time:                      message.Time,
		BlockTime:                 message.BlockTime,
	}
	return nil
}
