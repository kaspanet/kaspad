package model

import (
	"github.com/kaspanet/kaspad/util/daghash"
)

type DomainTransaction struct {
	Version      int32
	Inputs       []*DomainTransactionInput
	Outputs      []*DomainTransactionOutput
	LockTime     uint64
	SubnetworkID *DomainSubnetworkID
	Gas          uint64
	PayloadHash  *daghash.Hash
	Payload      []byte

	Hash  *daghash.Hash
	ID    *daghash.TxID
	Index int
}

type DomainTransactionInput struct {
	PreviousOutpoint *DomainOutpoint
	SignatureScript  []byte
	Sequence         uint64
}

type DomainOutpoint struct {
	ID    *daghash.TxID
	Index uint32
}

type DomainTransactionOutput struct {
	Value           uint64
	ScriptPublicKey []byte
}

type DomainTransactionID DomainHash
