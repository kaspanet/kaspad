package model

type DomainBlock struct {
	Header       *DomainBlockHeader
	Transactions []*DomainTransaction

	Hash *DomainHash
}

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
