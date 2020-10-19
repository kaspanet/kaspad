package externalapi

// DomainTransaction represents a Kaspa transaction
type DomainTransaction struct {
	Version      int32
	Inputs       []*DomainTransactionInput
	Outputs      []*DomainTransactionOutput
	LockTime     uint64
	SubnetworkID *DomainSubnetworkID
	Gas          uint64
	PayloadHash  *DomainHash
	Payload      []byte

	Fee  uint64
	Mass uint64
}

// DomainTransactionInput represents a Kaspa transaction input
type DomainTransactionInput struct {
	PreviousOutpoint *DomainOutpoint
	SignatureScript  []byte
	Sequence         uint64

	UTXOEntry *UTXOEntry
}

// DomainOutpoint represents a Kaspa transaction outpoint
type DomainOutpoint struct {
	ID    *DomainTransactionID
	Index uint32
}

// DomainTransactionOutput represents a Kaspad transaction output
type DomainTransactionOutput struct {
	Value           uint64
	ScriptPublicKey []byte
}

// DomainTransactionID represents the ID of a Kaspa transaction
type DomainTransactionID DomainHash
