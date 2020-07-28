package protowire

import (
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/subnetworkid"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_Transaction) toWireMessage() (*wire.MsgTx, error) {
	return x.toWireMessage()
}

func (x *KaspadMessage_Transaction) fromWireMessage(msgTx *wire.MsgTx) error {
	x.Transaction = new(TransactionMessage)
	x.Transaction.fromWireMessage(msgTx)
	return nil
}

func (x *TransactionMessage) toWireMessage() (*wire.MsgTx, error) {
	inputs := make([]*wire.TxIn, len(x.Inputs))
	for i, protoInput := range x.Inputs {
		prevTxID, err := protoInput.PreviousOutpoint.TransactionID.toWire()
		if err != nil {
			return nil, err
		}

		outpoint := wire.NewOutpoint(prevTxID, protoInput.PreviousOutpoint.Index)
		inputs[i] = wire.NewTxIn(outpoint, protoInput.SignatureScript)
	}

	if x.SubnetworkID == nil {
		return nil, errors.New("transaction subnetwork field cannot be nil")
	}

	subnetworkID, err := subnetworkid.New(x.SubnetworkID.Bytes)
	if err != nil {
		return nil, err
	}

	payloadHash, err := daghash.NewHash(x.PayloadHash.Bytes)
	if err != nil {
		return nil, err
	}

	return &wire.MsgTx{
		Version:      x.Version,
		TxIn:         inputs,
		TxOut:        nil,
		LockTime:     x.LockTime,
		SubnetworkID: *subnetworkID,
		Gas:          x.Gas,
		PayloadHash:  payloadHash,
		Payload:      x.Payload,
	}, nil
}

func (x *TransactionMessage) fromWireMessage(msgTx *wire.MsgTx) {
	protoInputs := make([]*TransactionInput, len(msgTx.TxIn))
	for i, input := range msgTx.TxIn {
		protoInputs[i] = &TransactionInput{
			PreviousOutpoint: &Outpoint{
				TransactionID: wireTransactionIDToProto(&input.PreviousOutpoint.TxID),
				Index:         input.PreviousOutpoint.Index,
			},
			SignatureScript: input.SignatureScript,
			Sequence:        input.Sequence,
		}
	}

	protoOutputs := make([]*TransactionOutput, len(msgTx.TxOut))
	for i, output := range msgTx.TxOut {
		protoOutputs[i] = &TransactionOutput{
			Value:        output.Value,
			ScriptPubKey: output.ScriptPubKey,
		}
	}

	*x = TransactionMessage{
		Version:      msgTx.Version,
		Inputs:       protoInputs,
		Outputs:      protoOutputs,
		LockTime:     msgTx.LockTime,
		SubnetworkID: wireSubnetworkIDToProto(&msgTx.SubnetworkID),
		Gas:          msgTx.Gas,
		PayloadHash:  wireHashToProto(msgTx.PayloadHash),
		Payload:      msgTx.Payload,
	}
}
