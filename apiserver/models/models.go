package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Block struct{
	gorm.Model
	ID uint
	BlockHash string
	Version int32
	HashMerkleRoot string
	AcceptedIDMerkleRoot	string
	UTXOCommitment	string
	Timestamp	time.Time
	Bits	uint32
	Nonce	uint64
	BlueScore	uint64
	IsChainBlock	bool
}

type ParentBlock struct{
	gorm.Model
	BlockID uint64
	ParentBlockID uint64
}

type AcceptingBlock struct{
	gorm.Model
	BlockID uint64
	ParentBlockID uint64
}

type RawBlock struct{
	gorm.Model
	BlockID uint64
	BlockData []byte
}

type Subnetwork struct{
	gorm.Model
	ID uint64
	SubnetworkID []byte
}

type Transactions struct {
	gorm.Model
	ID              uint64
	BlockID         uint64
	TransactionHash string
	TransactionID   string
	LockTime   uint64
	SubnetworkID   uint64
	Gas   uint64
	PayloadHash   string
	Payload   []byte
}

type TransactionsToBlocks struct{
	gorm.Model
	TransactionID uint64
	BlockID uint64
	LocationInBlock uint32
}

type TransactionOutputs struct{
	gorm.Model
	TransactionID uint64
	Index uint32
	Value uint64
	PkScript []byte
}

type TransactionInputs struct{
	gorm.Model
	TransactionID uint64
	TransactionOutputID uint64
	Index uint32
	SignatureScript []byte
	Sequence uint64
}

type UTXO struct{
	gorm.Model
	TransactionOutputID uint64
	AcceptingBlockID uint64
}