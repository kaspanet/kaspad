package externalapi

import (
	"fmt"
)

// DomainTransaction represents a Kaspa transaction
type DomainTransaction struct {
	Version      int32
	Inputs       []*DomainTransactionInput
	Outputs      []*DomainTransactionOutput
	LockTime     uint64
	SubnetworkID DomainSubnetworkID
	Gas          uint64
	PayloadHash  DomainHash
	Payload      []byte

	Fee  uint64
	Mass uint64
}

// Clone returns a clone of DomainTransaction
func (tx *DomainTransaction) Clone() *DomainTransaction {
	if tx == nil {
		return nil
	}

	payloadClone := make([]byte, len(tx.Payload))
	copy(payloadClone, tx.Payload)

	inputsClone := make([]*DomainTransactionInput, len(tx.Inputs))
	for i, input := range tx.Inputs {
		inputsClone[i] = input.Clone()
	}

	outputsClone := make([]*DomainTransactionOutput, len(tx.Outputs))
	for i, output := range tx.Outputs {
		outputsClone[i] = output.Clone()
	}

	return &DomainTransaction{
		Version:      tx.Version,
		Inputs:       inputsClone,
		Outputs:      outputsClone,
		LockTime:     tx.LockTime,
		SubnetworkID: *tx.SubnetworkID.Clone(),
		Gas:          tx.Gas,
		PayloadHash:  *tx.PayloadHash.Clone(),
		Payload:      payloadClone,
		Fee:          tx.Fee,
		Mass:         tx.Mass,
	}
}

// DomainTransactionInput represents a Kaspa transaction input
type DomainTransactionInput struct {
	PreviousOutpoint DomainOutpoint
	SignatureScript  []byte
	Sequence         uint64

	UTXOEntry *UTXOEntry
}

// Clone returns a clone of DomainTransactionInput
func (input *DomainTransactionInput) Clone() *DomainTransactionInput {
	if input == nil {
		return nil
	}

	signatureScriptClone := make([]byte, len(input.SignatureScript))
	copy(signatureScriptClone, input.SignatureScript)

	return &DomainTransactionInput{
		PreviousOutpoint: *input.PreviousOutpoint.Clone(),
		SignatureScript:  signatureScriptClone,
		Sequence:         input.Sequence,
		UTXOEntry:        input.UTXOEntry.Clone(),
	}
}

// DomainOutpoint represents a Kaspa transaction outpoint
type DomainOutpoint struct {
	TransactionID DomainTransactionID
	Index         uint32
}

// Clone returns a clone of DomainOutpoint
func (op *DomainOutpoint) Clone() *DomainOutpoint {
	if op == nil {
		return nil
	}

	return &DomainOutpoint{
		TransactionID: *op.TransactionID.Clone(),
		Index:         op.Index,
	}
}

// String stringifies an outpoint.
func (op DomainOutpoint) String() string {
	return fmt.Sprintf("(%s: %d)", op.TransactionID, op.Index)
}

// NewDomainOutpoint instantiates a new DomainOutpoint with the given id and index
func NewDomainOutpoint(id *DomainTransactionID, index uint32) *DomainOutpoint {
	return &DomainOutpoint{
		TransactionID: *id,
		Index:         index,
	}
}

// DomainTransactionOutput represents a Kaspad transaction output
type DomainTransactionOutput struct {
	Value           uint64
	ScriptPublicKey []byte
}

// Clone returns a clone of DomainTransactionOutput
func (output *DomainTransactionOutput) Clone() *DomainTransactionOutput {
	if output == nil {
		return nil
	}

	scriptPublicKeyClone := make([]byte, len(output.ScriptPublicKey))
	copy(scriptPublicKeyClone, output.ScriptPublicKey)

	return &DomainTransactionOutput{
		Value:           output.Value,
		ScriptPublicKey: scriptPublicKeyClone,
	}
}

// DomainTransactionID represents the ID of a Kaspa transaction
type DomainTransactionID DomainHash

// String stringifies a transaction ID.
func (id DomainTransactionID) String() string {
	return DomainHash(id).String()
}

// Clone returns a clone of DomainTransactionID
func (id *DomainTransactionID) Clone() *DomainTransactionID {
	if id == nil {
		return nil
	}

	idClone := *id
	return &idClone
}
