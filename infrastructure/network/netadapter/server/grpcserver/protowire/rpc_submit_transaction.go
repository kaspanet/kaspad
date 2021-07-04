package protowire

import (
	"math"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_SubmitTransactionRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_SubmitTransactionRequest is nil")
	}
	return x.SubmitTransactionRequest.toAppMessage()
}

func (x *KaspadMessage_SubmitTransactionRequest) fromAppMessage(message *appmessage.SubmitTransactionRequestMessage) error {
	x.SubmitTransactionRequest = &SubmitTransactionRequestMessage{
		Transaction: &RpcTransaction{},
		AllowOrphan: message.AllowOrphan,
	}
	x.SubmitTransactionRequest.Transaction.fromAppMessage(message.Transaction)
	return nil
}

func (x *SubmitTransactionRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "SubmitBlockRequestMessage is nil")
	}
	rpcTransaction, err := x.Transaction.toAppMessage()
	if err != nil {
		return nil, err
	}
	return &appmessage.SubmitTransactionRequestMessage{
		Transaction: rpcTransaction,
		AllowOrphan: x.AllowOrphan,
	}, nil
}

func (x *KaspadMessage_SubmitTransactionResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_SubmitTransactionResponse is nil")
	}
	return x.SubmitTransactionResponse.toAppMessage()
}

func (x *KaspadMessage_SubmitTransactionResponse) fromAppMessage(message *appmessage.SubmitTransactionResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.SubmitTransactionResponse = &SubmitTransactionResponseMessage{
		TransactionId: message.TransactionID,
		Error:         err,
	}
	return nil
}

func (x *SubmitTransactionResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "SubmitTransactionResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}
	return &appmessage.SubmitTransactionResponseMessage{
		TransactionID: x.TransactionId,
		Error:         rpcErr,
	}, nil
}

func (x *RpcTransaction) toAppMessage() (*appmessage.RPCTransaction, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RpcTransaction is nil")
	}
	inputs := make([]*appmessage.RPCTransactionInput, len(x.Inputs))
	for i, input := range x.Inputs {
		appInput, err := input.toAppMessage()
		if err != nil {
			return nil, err
		}
		inputs[i] = appInput
	}
	outputs := make([]*appmessage.RPCTransactionOutput, len(x.Outputs))
	for i, output := range x.Outputs {
		appOutput, err := output.toAppMessage()
		if err != nil {
			return nil, err
		}
		outputs[i] = appOutput
	}
	if x.Version > math.MaxUint16 {
		return nil, errors.Errorf("Invalid RPC transaction version - bigger then uint16")
	}
	var verboseData *appmessage.RPCTransactionVerboseData
	if x.VerboseData != nil {
		appMessageVerboseData, err := x.VerboseData.toAppMessage()
		if err != nil {
			return nil, err
		}
		verboseData = appMessageVerboseData
	}
	return &appmessage.RPCTransaction{
		Version:      uint16(x.Version),
		Inputs:       inputs,
		Outputs:      outputs,
		LockTime:     x.LockTime,
		SubnetworkID: x.SubnetworkId,
		Gas:          x.Gas,
		Payload:      x.Payload,
		VerboseData:  verboseData,
	}, nil
}

func (x *RpcTransaction) fromAppMessage(transaction *appmessage.RPCTransaction) {
	inputs := make([]*RpcTransactionInput, len(transaction.Inputs))
	for i, input := range transaction.Inputs {
		inputs[i] = &RpcTransactionInput{}
		inputs[i].fromAppMessage(input)
	}
	outputs := make([]*RpcTransactionOutput, len(transaction.Outputs))
	for i, output := range transaction.Outputs {
		outputs[i] = &RpcTransactionOutput{}
		outputs[i].fromAppMessage(output)
	}
	var verboseData *RpcTransactionVerboseData
	if transaction.VerboseData != nil {
		verboseData = &RpcTransactionVerboseData{}
		verboseData.fromAppMessage(transaction.VerboseData)
	}
	*x = RpcTransaction{
		Version:      uint32(transaction.Version),
		Inputs:       inputs,
		Outputs:      outputs,
		LockTime:     transaction.LockTime,
		SubnetworkId: transaction.SubnetworkID,
		Gas:          transaction.Gas,
		Payload:      transaction.Payload,
		VerboseData:  verboseData,
	}
}

func (x *RpcTransactionInput) toAppMessage() (*appmessage.RPCTransactionInput, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RpcTransactionInput is nil")
	}
	if x.SigOpCount > math.MaxUint8 {
		return nil, errors.New("TransactionInput SigOpCount > math.MaxUint8")
	}
	outpoint, err := x.PreviousOutpoint.toAppMessage()
	if err != nil {
		return nil, err
	}
	var verboseData *appmessage.RPCTransactionInputVerboseData
	for x.VerboseData != nil {
		appMessageVerboseData, err := x.VerboseData.toAppMessage()
		if err != nil {
			return nil, err
		}
		verboseData = appMessageVerboseData
	}
	return &appmessage.RPCTransactionInput{
		PreviousOutpoint: outpoint,
		SignatureScript:  x.SignatureScript,
		Sequence:         x.Sequence,
		VerboseData:      verboseData,
		SigOpCount:       byte(x.SigOpCount),
	}, nil
}

func (x *RpcTransactionInput) fromAppMessage(message *appmessage.RPCTransactionInput) {
	previousOutpoint := &RpcOutpoint{}
	previousOutpoint.fromAppMessage(message.PreviousOutpoint)
	var verboseData *RpcTransactionInputVerboseData
	if message.VerboseData != nil {
		verboseData := &RpcTransactionInputVerboseData{}
		verboseData.fromAppData(message.VerboseData)
	}
	*x = RpcTransactionInput{
		PreviousOutpoint: previousOutpoint,
		SignatureScript:  message.SignatureScript,
		Sequence:         message.Sequence,
		VerboseData:      verboseData,
		SigOpCount:       uint32(message.SigOpCount),
	}
}

func (x *RpcTransactionOutput) toAppMessage() (*appmessage.RPCTransactionOutput, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RpcTransactionOutput is nil")
	}
	scriptPublicKey, err := x.ScriptPublicKey.toAppMessage()
	if err != nil {
		return nil, err
	}
	var verboseData *appmessage.RPCTransactionOutputVerboseData
	if x.VerboseData != nil {
		appMessageVerboseData, err := x.VerboseData.toAppMessage()
		if err != nil {
			return nil, err
		}
		verboseData = appMessageVerboseData
	}
	return &appmessage.RPCTransactionOutput{
		Amount:          x.Amount,
		ScriptPublicKey: scriptPublicKey,
		VerboseData:     verboseData,
	}, nil
}

func (x *RpcTransactionOutput) fromAppMessage(message *appmessage.RPCTransactionOutput) {
	scriptPublicKey := &RpcScriptPublicKey{}
	scriptPublicKey.fromAppMessage(message.ScriptPublicKey)
	var verboseData *RpcTransactionOutputVerboseData
	if message.VerboseData != nil {
		verboseData = &RpcTransactionOutputVerboseData{}
		verboseData.fromAppMessage(message.VerboseData)
	}
	*x = RpcTransactionOutput{
		Amount:          message.Amount,
		ScriptPublicKey: scriptPublicKey,
		VerboseData:     verboseData,
	}
}

func (x *RpcOutpoint) toAppMessage() (*appmessage.RPCOutpoint, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RpcOutpoint is nil")
	}
	return &appmessage.RPCOutpoint{
		TransactionID: x.TransactionId,
		Index:         x.Index,
	}, nil
}

func (x *RpcOutpoint) fromAppMessage(message *appmessage.RPCOutpoint) {
	*x = RpcOutpoint{
		TransactionId: message.TransactionID,
		Index:         message.Index,
	}
}

func (x *RpcUtxoEntry) toAppMessage() (*appmessage.RPCUTXOEntry, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RpcUtxoEntry is nil")
	}
	scriptPubKey, err := x.ScriptPublicKey.toAppMessage()
	if err != nil {
		return nil, err
	}
	return &appmessage.RPCUTXOEntry{
		Amount:          x.Amount,
		ScriptPublicKey: scriptPubKey,
		BlockDAAScore:   x.BlockDaaScore,
		IsCoinbase:      x.IsCoinbase,
	}, nil
}

func (x *RpcUtxoEntry) fromAppMessage(message *appmessage.RPCUTXOEntry) {
	scriptPublicKey := &RpcScriptPublicKey{}
	scriptPublicKey.fromAppMessage(message.ScriptPublicKey)
	*x = RpcUtxoEntry{
		Amount:          message.Amount,
		ScriptPublicKey: scriptPublicKey,
		BlockDaaScore:   message.BlockDAAScore,
		IsCoinbase:      message.IsCoinbase,
	}
}

func (x *RpcScriptPublicKey) toAppMessage() (*appmessage.RPCScriptPublicKey, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RpcScriptPublicKey is nil")
	}
	if x.Version > math.MaxUint16 {
		return nil, errors.Errorf("Invalid header version - bigger then uint16")
	}
	return &appmessage.RPCScriptPublicKey{
		Version: uint16(x.Version),
		Script:  x.ScriptPublicKey,
	}, nil
}

func (x *RpcScriptPublicKey) fromAppMessage(message *appmessage.RPCScriptPublicKey) {
	*x = RpcScriptPublicKey{
		Version:         uint32(message.Version),
		ScriptPublicKey: message.Script,
	}
}

func (x *RpcTransactionVerboseData) toAppMessage() (*appmessage.RPCTransactionVerboseData, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RpcTransactionVerboseData is nil")
	}
	return &appmessage.RPCTransactionVerboseData{
		TransactionID: x.TransactionId,
		Hash:          x.Hash,
		Size:          x.Size,
		BlockHash:     x.BlockHash,
		BlockTime:     x.BlockTime,
	}, nil
}

func (x *RpcTransactionVerboseData) fromAppMessage(message *appmessage.RPCTransactionVerboseData) {
	*x = RpcTransactionVerboseData{
		TransactionId: message.TransactionID,
		Hash:          message.Hash,
		Size:          message.Size,
		BlockHash:     message.BlockHash,
		BlockTime:     message.BlockTime,
	}
}

func (x *RpcTransactionInputVerboseData) toAppMessage() (*appmessage.RPCTransactionInputVerboseData, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RpcTransactionInputVerboseData is nil")
	}
	return &appmessage.RPCTransactionInputVerboseData{}, nil
}

func (x *RpcTransactionInputVerboseData) fromAppData(message *appmessage.RPCTransactionInputVerboseData) {
	*x = RpcTransactionInputVerboseData{}
}

func (x *RpcTransactionOutputVerboseData) toAppMessage() (*appmessage.RPCTransactionOutputVerboseData, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RpcTransactionOutputVerboseData is nil")
	}
	return &appmessage.RPCTransactionOutputVerboseData{
		ScriptPublicKeyType:    x.ScriptPublicKeyType,
		ScriptPublicKeyAddress: x.ScriptPublicKeyAddress,
	}, nil
}

func (x *RpcTransactionOutputVerboseData) fromAppMessage(message *appmessage.RPCTransactionOutputVerboseData) {
	*x = RpcTransactionOutputVerboseData{
		ScriptPublicKeyType:    message.ScriptPublicKeyType,
		ScriptPublicKeyAddress: message.ScriptPublicKeyAddress,
	}
}
