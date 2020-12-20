package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_SubmitTransactionRequest) toAppMessage() (appmessage.Message, error) {
	rpcTransaction := x.SubmitTransactionRequest.Transaction.toAppMessage()
	return &appmessage.SubmitTransactionRequestMessage{
		Transaction: rpcTransaction,
	}, nil
}

func (x *KaspadMessage_SubmitTransactionRequest) fromAppMessage(message *appmessage.SubmitTransactionRequestMessage) error {
	x.SubmitTransactionRequest = &SubmitTransactionRequestMessage{
		Transaction: &RpcTransaction{},
	}
	x.SubmitTransactionRequest.Transaction.fromAppMessage(message.Transaction)
	return nil
}

func (x *KaspadMessage_SubmitTransactionResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.SubmitTransactionResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.SubmitTransactionResponse.Error.Message}
	}
	return &appmessage.SubmitTransactionResponseMessage{
		TransactionID: x.SubmitTransactionResponse.TransactionId,
		Error:         err,
	}, nil
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

func (x *RpcTransaction) toAppMessage() *appmessage.RPCTransaction {
	inputs := make([]*appmessage.RPCTransactionInput, len(x.Inputs))
	for i, input := range x.Inputs {
		previousOutpoint := &appmessage.RPCOutpoint{
			TransactionID: input.PreviousOutpoint.TransactionId,
			Index:         input.PreviousOutpoint.Index,
		}
		inputs[i] = &appmessage.RPCTransactionInput{
			PreviousOutpoint: previousOutpoint,
			SignatureScript:  input.SignatureScript,
			Sequence:         input.Sequence,
		}
	}
	outputs := make([]*appmessage.RPCTransactionOutput, len(x.Outputs))
	for i, output := range x.Outputs {
		outputs[i] = &appmessage.RPCTransactionOutput{
			Amount:       output.Amount,
			ScriptPubKey: output.ScriptPubKey,
		}
	}
	return &appmessage.RPCTransaction{
		Version:      x.Version,
		Inputs:       inputs,
		Outputs:      outputs,
		LockTime:     x.LockTime,
		SubnetworkID: x.SubnetworkId,
		Gas:          x.Gas,
		PayloadHash:  x.PayloadHash,
		Payload:      x.Payload,
	}
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
			Amount:       output.Amount,
			ScriptPubKey: output.ScriptPubKey,
		}
	}
	*x = RpcTransaction{
		Version:      transaction.Version,
		Inputs:       inputs,
		Outputs:      outputs,
		LockTime:     transaction.LockTime,
		SubnetworkId: transaction.SubnetworkID,
		Gas:          transaction.Gas,
		PayloadHash:  transaction.PayloadHash,
		Payload:      transaction.Payload,
	}
}
