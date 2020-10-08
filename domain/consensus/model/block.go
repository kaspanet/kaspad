package model

import (
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/mstime"
)

type DomainBlock struct {
	Header       *DomainBlockHeader
	Transactions []*DomainTransaction

	Hash *daghash.Hash
}

type DomainBlockHeader struct {
	Version              int32
	ParentHashes         []*daghash.Hash
	HashMerkleRoot       *daghash.Hash
	AcceptedIDMerkleRoot *daghash.Hash
	UTXOCommitment       *daghash.Hash
	Timestamp            mstime.Time
	Bits                 uint32
	Nonce                uint64
}
