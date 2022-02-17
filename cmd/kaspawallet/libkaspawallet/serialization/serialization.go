package serialization

import (
	"math"

	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet/serialization/protoserialization"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/pkg/errors"
)

// PartiallySignedTransaction is a type that is intended
// to be transferred between multiple parties so each
// party will be able to sign the transaction before
// it's fully signed.
type PartiallySignedTransaction struct {
	Tx                    *externalapi.DomainTransaction
	PartiallySignedInputs []*PartiallySignedInput
}

// PartiallySignedInput represents an input signed
// only by some of the relevant parties.
type PartiallySignedInput struct {
	PrevOutput           *externalapi.DomainTransactionOutput
	MinimumSignatures    uint32
	PubKeySignaturePairs []*PubKeySignaturePair
	DerivationPath       string
}

// PubKeySignaturePair is a pair of public key and (potentially) its associated signature
type PubKeySignaturePair struct {
	ExtendedPublicKey string
	Signature         []byte
}

func (pst *PartiallySignedTransaction) Clone() *PartiallySignedTransaction {
	clone := &PartiallySignedTransaction{
		Tx:                    pst.Tx.Clone(),
		PartiallySignedInputs: make([]*PartiallySignedInput, len(pst.PartiallySignedInputs)),
	}
	for i, partiallySignedInput := range pst.PartiallySignedInputs {
		clone.PartiallySignedInputs[i] = partiallySignedInput.Clone()
	}
	return clone
}

func (psi PartiallySignedInput) Clone() *PartiallySignedInput {
	clone := &PartiallySignedInput{
		PrevOutput:           psi.PrevOutput.Clone(),
		MinimumSignatures:    psi.MinimumSignatures,
		PubKeySignaturePairs: make([]*PubKeySignaturePair, len(psi.PubKeySignaturePairs)),
		DerivationPath:       psi.DerivationPath,
	}
	for i, pubKeySignaturePair := range psi.PubKeySignaturePairs {
		clone.PubKeySignaturePairs[i] = pubKeySignaturePair.Clone()
	}
	return clone
}

func (psp PubKeySignaturePair) Clone() *PubKeySignaturePair {
	clone := &PubKeySignaturePair{
		ExtendedPublicKey: psp.ExtendedPublicKey,
		Signature:         make([]byte, len(psp.Signature)),
	}
	copy(clone.Signature, psp.Signature)
	return clone
}

// DeserializePartiallySignedTransaction deserializes a byte slice into PartiallySignedTransaction.
func DeserializePartiallySignedTransaction(serializedPartiallySignedTransaction []byte) (*PartiallySignedTransaction, error) {
	protoPartiallySignedTransaction := &protoserialization.PartiallySignedTransaction{}
	err := proto.Unmarshal(serializedPartiallySignedTransaction, protoPartiallySignedTransaction)
	if err != nil {
		return nil, err
	}

	return partiallySignedTransactionFromProto(protoPartiallySignedTransaction)
}

// SerializePartiallySignedTransaction serializes a PartiallySignedTransaction.
func SerializePartiallySignedTransaction(partiallySignedTransaction *PartiallySignedTransaction) ([]byte, error) {
	return proto.Marshal(partiallySignedTransactionToProto(partiallySignedTransaction))
}

func partiallySignedTransactionFromProto(protoPartiallySignedTransaction *protoserialization.PartiallySignedTransaction) (*PartiallySignedTransaction, error) {
	tx, err := transactionFromProto(protoPartiallySignedTransaction.Tx)
	if err != nil {
		return nil, err
	}

	inputs := make([]*PartiallySignedInput, len(protoPartiallySignedTransaction.PartiallySignedInputs))
	for i, protoInput := range protoPartiallySignedTransaction.PartiallySignedInputs {
		inputs[i], err = partiallySignedInputFromProto(protoInput)
		if err != nil {
			return nil, err
		}
	}

	return &PartiallySignedTransaction{
		Tx:                    tx,
		PartiallySignedInputs: inputs,
	}, nil
}

func partiallySignedTransactionToProto(partiallySignedTransaction *PartiallySignedTransaction) *protoserialization.PartiallySignedTransaction {
	protoInputs := make([]*protoserialization.PartiallySignedInput, len(partiallySignedTransaction.PartiallySignedInputs))
	for i, input := range partiallySignedTransaction.PartiallySignedInputs {
		protoInputs[i] = partiallySignedInputToProto(input)
	}

	return &protoserialization.PartiallySignedTransaction{
		Tx:                    transactionToProto(partiallySignedTransaction.Tx),
		PartiallySignedInputs: protoInputs,
	}
}

func partiallySignedInputFromProto(protoPartiallySignedInput *protoserialization.PartiallySignedInput) (*PartiallySignedInput, error) {
	output, err := transactionOutputFromProto(protoPartiallySignedInput.PrevOutput)
	if err != nil {
		return nil, err
	}

	pubKeySignaturePairs := make([]*PubKeySignaturePair, len(protoPartiallySignedInput.PubKeySignaturePairs))
	for i, protoPair := range protoPartiallySignedInput.PubKeySignaturePairs {
		pubKeySignaturePairs[i] = pubKeySignaturePairFromProto(protoPair)
	}

	return &PartiallySignedInput{
		PrevOutput:           output,
		MinimumSignatures:    protoPartiallySignedInput.MinimumSignatures,
		PubKeySignaturePairs: pubKeySignaturePairs,
		DerivationPath:       protoPartiallySignedInput.DerivationPath,
	}, nil
}

func partiallySignedInputToProto(partiallySignedInput *PartiallySignedInput) *protoserialization.PartiallySignedInput {
	protoPairs := make([]*protoserialization.PubKeySignaturePair, len(partiallySignedInput.PubKeySignaturePairs))
	for i, pair := range partiallySignedInput.PubKeySignaturePairs {
		protoPairs[i] = pubKeySignaturePairToProto(pair)
	}

	return &protoserialization.PartiallySignedInput{
		PrevOutput:           transactionOutputToProto(partiallySignedInput.PrevOutput),
		MinimumSignatures:    partiallySignedInput.MinimumSignatures,
		PubKeySignaturePairs: protoPairs,
		DerivationPath:       partiallySignedInput.DerivationPath,
	}
}

func pubKeySignaturePairFromProto(protoPubKeySignaturePair *protoserialization.PubKeySignaturePair) *PubKeySignaturePair {
	return &PubKeySignaturePair{
		ExtendedPublicKey: protoPubKeySignaturePair.ExtendedPubKey,
		Signature:         protoPubKeySignaturePair.Signature,
	}
}

func pubKeySignaturePairToProto(pubKeySignaturePair *PubKeySignaturePair) *protoserialization.PubKeySignaturePair {
	return &protoserialization.PubKeySignaturePair{
		ExtendedPubKey: pubKeySignaturePair.ExtendedPublicKey,
		Signature:      pubKeySignaturePair.Signature,
	}
}

func transactionFromProto(protoTransaction *protoserialization.TransactionMessage) (*externalapi.DomainTransaction, error) {
	if protoTransaction.Version > math.MaxUint16 {
		return nil, errors.Errorf("protoTransaction.Version is %d and is too big to be a uint16", protoTransaction.Version)
	}

	inputs := make([]*externalapi.DomainTransactionInput, len(protoTransaction.Inputs))
	for i, protoInput := range protoTransaction.Inputs {
		var err error
		inputs[i], err = transactionInputFromProto(protoInput)
		if err != nil {
			return nil, err
		}
	}

	outputs := make([]*externalapi.DomainTransactionOutput, len(protoTransaction.Outputs))
	for i, protoOutput := range protoTransaction.Outputs {
		var err error
		outputs[i], err = transactionOutputFromProto(protoOutput)
		if err != nil {
			return nil, err
		}
	}

	subnetworkID, err := subnetworks.FromBytes(protoTransaction.SubnetworkId.Bytes)
	if err != nil {
		return nil, err
	}

	return &externalapi.DomainTransaction{
		Version:      uint16(protoTransaction.Version),
		Inputs:       inputs,
		Outputs:      outputs,
		LockTime:     protoTransaction.LockTime,
		SubnetworkID: *subnetworkID,
		Gas:          protoTransaction.Gas,
		Payload:      protoTransaction.Payload,
	}, nil
}

func transactionToProto(tx *externalapi.DomainTransaction) *protoserialization.TransactionMessage {
	protoInputs := make([]*protoserialization.TransactionInput, len(tx.Inputs))
	for i, input := range tx.Inputs {
		protoInputs[i] = transactionInputToProto(input)
	}

	protoOutputs := make([]*protoserialization.TransactionOutput, len(tx.Outputs))
	for i, output := range tx.Outputs {
		protoOutputs[i] = transactionOutputToProto(output)
	}

	return &protoserialization.TransactionMessage{
		Version:      uint32(tx.Version),
		Inputs:       protoInputs,
		Outputs:      protoOutputs,
		LockTime:     tx.LockTime,
		SubnetworkId: &protoserialization.SubnetworkId{Bytes: tx.SubnetworkID[:]},
		Gas:          tx.Gas,
		Payload:      tx.Payload,
	}
}

func transactionInputFromProto(protoInput *protoserialization.TransactionInput) (*externalapi.DomainTransactionInput, error) {
	if protoInput.SigOpCount > math.MaxUint8 {
		return nil, errors.New("TransactionInput SigOpCount > math.MaxUint8")
	}

	outpoint, err := outpointFromProto(protoInput.PreviousOutpoint)
	if err != nil {
		return nil, err
	}

	return &externalapi.DomainTransactionInput{
		PreviousOutpoint: *outpoint,
		SignatureScript:  protoInput.SignatureScript,
		Sequence:         protoInput.Sequence,
		SigOpCount:       byte(protoInput.SigOpCount),
	}, nil
}

func transactionInputToProto(input *externalapi.DomainTransactionInput) *protoserialization.TransactionInput {
	return &protoserialization.TransactionInput{
		PreviousOutpoint: outpointToProto(&input.PreviousOutpoint),
		SignatureScript:  input.SignatureScript,
		Sequence:         input.Sequence,
		SigOpCount:       uint32(input.SigOpCount),
	}
}

func outpointFromProto(protoOutpoint *protoserialization.Outpoint) (*externalapi.DomainOutpoint, error) {
	txID, err := transactionIDFromProto(protoOutpoint.TransactionId)
	if err != nil {
		return nil, err
	}
	return &externalapi.DomainOutpoint{
		TransactionID: *txID,
		Index:         protoOutpoint.Index,
	}, nil
}

func outpointToProto(outpoint *externalapi.DomainOutpoint) *protoserialization.Outpoint {
	return &protoserialization.Outpoint{
		TransactionId: &protoserialization.TransactionId{Bytes: outpoint.TransactionID.ByteSlice()},
		Index:         outpoint.Index,
	}
}

func transactionIDFromProto(protoTxID *protoserialization.TransactionId) (*externalapi.DomainTransactionID, error) {
	if protoTxID == nil {
		return nil, errors.Errorf("protoTxID is nil")
	}

	return externalapi.NewDomainTransactionIDFromByteSlice(protoTxID.Bytes)
}

func transactionOutputFromProto(protoOutput *protoserialization.TransactionOutput) (*externalapi.DomainTransactionOutput, error) {
	scriptPublicKey, err := scriptPublicKeyFromProto(protoOutput.ScriptPublicKey)
	if err != nil {
		return nil, err
	}

	return &externalapi.DomainTransactionOutput{
		Value:           protoOutput.Value,
		ScriptPublicKey: scriptPublicKey,
	}, nil
}

func transactionOutputToProto(output *externalapi.DomainTransactionOutput) *protoserialization.TransactionOutput {
	return &protoserialization.TransactionOutput{
		Value:           output.Value,
		ScriptPublicKey: scriptPublicKeyToProto(output.ScriptPublicKey),
	}
}

func scriptPublicKeyFromProto(protoScriptPublicKey *protoserialization.ScriptPublicKey) (*externalapi.ScriptPublicKey, error) {
	if protoScriptPublicKey.Version > math.MaxUint16 {
		return nil, errors.Errorf("protoOutput.ScriptPublicKey.Version is %d and is too big to be a uint16", protoScriptPublicKey.Version)
	}
	return &externalapi.ScriptPublicKey{
		Script:  protoScriptPublicKey.Script,
		Version: uint16(protoScriptPublicKey.Version),
	}, nil
}

func scriptPublicKeyToProto(scriptPublicKey *externalapi.ScriptPublicKey) *protoserialization.ScriptPublicKey {
	return &protoserialization.ScriptPublicKey{
		Script:  scriptPublicKey.Script,
		Version: uint32(scriptPublicKey.Version),
	}
}
