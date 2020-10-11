package model

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
	HashMerkleRoot       *DomainHash
	AcceptedIDMerkleRoot *DomainHash
	UTXOCommitment       *DomainHash
	Timestamp            *DomainTime
	Bits                 uint32
	Nonce                uint64
}
