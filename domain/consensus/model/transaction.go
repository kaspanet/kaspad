package model

type DomainTransaction struct {
	Version      int32
	Inputs       []*DomainTransactionInput
	Outputs      []*DomainTransactionOutput
	LockTime     uint64
	SubnetworkID *DomainSubnetworkID
	Gas          uint64
	PayloadHash  *DomainHash
	Payload      []byte

	Hash  *DomainHash
	ID    *DomainTransactionID
	Index int
}

type DomainTransactionInput struct {
	PreviousOutpoint *DomainOutpoint
	SignatureScript  []byte
	Sequence         uint64
}

type DomainOutpoint struct {
	ID    *DomainTransactionID
	Index uint32
}

type DomainTransactionOutput struct {
	Value           uint64
	ScriptPublicKey []byte
}

type DomainTransactionID DomainHash
