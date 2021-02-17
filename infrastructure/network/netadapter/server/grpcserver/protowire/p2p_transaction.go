package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
	"math"
)

func (x *KaspadMessage_Transaction) toAppMessage() (appmessage.Message, error) {
	return x.Transaction.toAppMessage()
}

func (x *KaspadMessage_Transaction) fromAppMessage(msgTx *appmessage.MsgTx) error {
	x.Transaction = new(TransactionMessage)
	x.Transaction.fromAppMessage(msgTx)
	return nil
}

func (x *TransactionMessage) toAppMessage() (appmessage.Message, error) {
	inputs := make([]*appmessage.TxIn, len(x.Inputs))
	for i, protoInput := range x.Inputs {
		prevTxID, err := protoInput.PreviousOutpoint.TransactionId.toDomain()
		if err != nil {
			return nil, err
		}

		outpoint := appmessage.NewOutpoint(prevTxID, protoInput.PreviousOutpoint.Index)
		inputs[i] = appmessage.NewTxIn(outpoint, protoInput.SignatureScript, protoInput.Sequence)
	}

	outputs := make([]*appmessage.TxOut, len(x.Outputs))
	for i, protoOutput := range x.Outputs {
		if protoOutput.ScriptPublicKey.Version > math.MaxUint16 {
			return nil, errors.Errorf("The version on ScriptPublicKey is bigger then uint16.")
		}
		outputs[i] = &appmessage.TxOut{
			Value:        protoOutput.Value,
			ScriptPubKey: &externalapi.ScriptPublicKey{protoOutput.ScriptPublicKey.Script, uint16(protoOutput.ScriptPublicKey.Version)},
		}
	}

	if x.SubnetworkId == nil {
		return nil, errors.New("transaction subnetwork field cannot be nil")
	}

	subnetworkID, err := x.SubnetworkId.toDomain()
	if err != nil {
		return nil, err
	}

	payloadHash := &externalapi.DomainHash{}
	if x.PayloadHash != nil {
		payloadHash, err = x.PayloadHash.toDomain()
		if err != nil {
			return nil, err
		}
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
		PayloadHash:  *payloadHash,
		Payload:      x.Payload,
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
		PayloadHash:  domainHashToProto(&msgTx.PayloadHash),
		Payload:      msgTx.Payload,
	}

}
