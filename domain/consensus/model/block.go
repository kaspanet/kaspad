package model

import (
	"github.com/kaspanet/kaspad/util/mstime"
)

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
	Timestamp            mstime.Time
	Bits                 uint32
	Nonce                uint64
}
