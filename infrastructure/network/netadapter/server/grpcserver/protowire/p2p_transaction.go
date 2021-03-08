package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
	"math"
)

func (x *KaspadMessage_Transaction) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_Transaction is nil")
	}
	return x.Transaction.toAppMessage()
}

func (x *KaspadMessage_Transaction) fromAppMessage(msgTx *appmessage.MsgTx) error {
	x.Transaction = new(TransactionMessage)
	x.Transaction.fromAppMessage(msgTx)
	return nil
}

func (x *TransactionMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "TransactionMessage is nil")
	}
	inputs := make([]*appmessage.TxIn, len(x.Inputs))
	for i, protoInput := range x.Inputs {
		input, err := protoInput.toAppMessage()
		if err != nil {
			return nil, err
		}
		inputs[i] = input
	}

	outputs := make([]*appmessage.TxOut, len(x.Outputs))
	for i, protoOutput := range x.Outputs {
		output, err := protoOutput.toAppMessage()
		if err != nil {
			return nil, err
		}
		outputs[i] = output
	}

	subnetworkID, err := x.SubnetworkId.toDomain()
	if err != nil {
		return nil, err
	}

	if x.Version > math.MaxUint16 {
		return nil, errors.Errorf("Invalid transaction version - bigger then uint16")
	}
	return &appmessage.MsgTx{
		Version:      uint16(x.Version),
		TxIn:         inputs,
		TxOut:        outputs,
		LockTime:     x.LockTime,
		SubnetworkID: *subnetworkID,
		Gas:          x.Gas,
		Payload:      x.Payload,
	}, nil
}

func (x *TransactionInput) toAppMessage() (*appmessage.TxIn, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "TransactionInput is nil")
	}
	outpoint, err := x.PreviousOutpoint.toAppMessage()
	if err != nil {
		return nil, err
	}
	return appmessage.NewTxIn(outpoint, x.SignatureScript, x.Sequence), nil
}

func (x *TransactionOutput) toAppMessage() (*appmessage.TxOut, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "TransactionOutput is nil")
	}
	scriptPublicKey, err := x.ScriptPublicKey.toAppMessage()
	if err != nil {
		return nil, err
	}
	return &appmessage.TxOut{
		Value:        x.Value,
		ScriptPubKey: scriptPublicKey,
	}, nil
}

func (x *TransactionMessage) fromAppMessage(msgTx *appmessage.MsgTx) {
	protoInputs := make([]*TransactionInput, len(msgTx.TxIn))
	for i, input := range msgTx.TxIn {
		protoInputs[i] = &TransactionInput{
			PreviousOutpoint: &Outpoint{
				TransactionId: domainTransactionIDToProto(&input.PreviousOutpoint.TxID),
				Index:         input.PreviousOutpoint.Index,
			},
			SignatureScript: input.SignatureScript,
			Sequence:        input.Sequence,
		}
	}

	protoOutputs := make([]*TransactionOutput, len(msgTx.TxOut))
	for i, output := range msgTx.TxOut {
		protoOutputs[i] = &TransactionOutput{
			Value: output.Value,
			ScriptPublicKey: &ScriptPublicKey{
				Script:  output.ScriptPubKey.Script,
				Version: uint32(output.ScriptPubKey.Version),
			},
		}
	}

	*x = TransactionMessage{
		Version:      uint32(msgTx.Version),
		Inputs:       protoInputs,
		Outputs:      protoOutputs,
		LockTime:     msgTx.LockTime,
		SubnetworkId: domainSubnetworkIDToProto(&msgTx.SubnetworkID),
		Gas:          msgTx.Gas,
		Payload:      msgTx.Payload,
	}

}
