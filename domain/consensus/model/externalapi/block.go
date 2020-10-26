package externalapi

// DomainBlock represents a Kaspa block
type DomainBlock struct {
	Header       *DomainBlockHeader
	Transactions []*DomainTransaction

	Hash *DomainHash
}

// DomainBlockHeader represents the header part of a Kaspa block
type DomainBlockHeader struct {
	Version              int32
	ParentHashes         []*DomainHash
	HashMerkleRoot       DomainHash
	AcceptedIDMerkleRoot DomainHash
	UTXOCommitment       DomainHash
	TimeInMilliseconds   int64
	Bits                 uint32
	Nonce                uint64
}
