package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
	"math"
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
		return nil, errors.Errorf("Invalid RPC txn version - bigger then uint16")
	}

	return &appmessage.RPCTransaction{
		Version:      uint16(x.Version),
		Inputs:       inputs,
		Outputs:      outputs,
		LockTime:     x.LockTime,
		SubnetworkID: x.SubnetworkId,
		Gas:          x.Gas,
		PayloadHash:  x.PayloadHash,
		Payload:      x.Payload,
	}, nil
}

func (x *RpcTransaction) fromAppMessage(transaction *appmessage.RPCTransaction) {
	inputs := make([]*RpcTransactionInput, len(transaction.Inputs))
	for i, input := range transaction.Inputs {
		previousOutpoint := &RpcOutpoint{
			TransactionId: input.PreviousOutpoint.TransactionID,
			Index:         input.PreviousOutpoint.Index,
		}
		inputs[i] = &RpcTransactionInput{
			PreviousOutpoint: previousOutpoint,
			SignatureScript:  input.SignatureScript,
			Sequence:         input.Sequence,
		}
	}
	outputs := make([]*RpcTransactionOutput, len(transaction.Outputs))
	for i, output := range transaction.Outputs {
		outputs[i] = &RpcTransactionOutput{
			Amount:          output.Amount,
			ScriptPublicKey: ConvertFromRPCScriptPubKeyToAppMsgRPCScriptPubKey(output.ScriptPublicKey),
		}
	}
	*x = RpcTransaction{
		Version:      uint32(transaction.Version),
		Inputs:       inputs,
		Outputs:      outputs,
		LockTime:     transaction.LockTime,
		SubnetworkId: transaction.SubnetworkID,
		Gas:          transaction.Gas,
		PayloadHash:  transaction.PayloadHash,
		Payload:      transaction.Payload,
	}
}

func (x *RpcTransactionInput) toAppMessage() (*appmessage.RPCTransactionInput, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RpcTransactionInput is nil")
	}
	outpoint, err := x.PreviousOutpoint.toAppMessage()
	if err != nil {
		return nil, err
	}
	return &appmessage.RPCTransactionInput{
		PreviousOutpoint: outpoint,
		SignatureScript:  x.SignatureScript,
		Sequence:         x.Sequence,
	}, nil
}

func (x *RpcTransactionOutput) toAppMessage() (*appmessage.RPCTransactionOutput, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RpcTransactionOutput is nil")
	}
	scriptPubKey, err := x.ScriptPublicKey.toAppMessage()
	if err != nil {
		return nil, err
	}
	return &appmessage.RPCTransactionOutput{
		Amount:          x.Amount,
		ScriptPublicKey: scriptPubKey,
	}, nil
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

// ConvertFromRPCScriptPubKeyToAppMsgRPCScriptPubKey converts from RPCScriptPublicKey to RpcScriptPubKey.
func ConvertFromRPCScriptPubKeyToAppMsgRPCScriptPubKey(toConvert *appmessage.RPCScriptPublicKey) *RpcScriptPublicKey {
	return &RpcScriptPublicKey{Version: uint32(toConvert.Version), ScriptPublicKey: toConvert.Script}
}
