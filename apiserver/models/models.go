package models

import (
	"time"
)

type Block struct {
	ID                   uint64 `gorm:"primary_key"`
	BlockHash            string
	AcceptingBlockID     uint64
	AcceptingBlock       *Block
	Version              int32
	HashMerkleRoot       string
	AcceptedIDMerkleRoot string
	UTXOCommitment       string
	Timestamp            time.Time
	Bits                 uint32
	Nonce                uint64
	BlueScore            uint64
	IsChainBlock         bool
	ParentBlocks         []Block `gorm:"many2many:parent_blocks;"`
}

type ParentBlock struct {
	BlockID       uint64
	Block         Block
	ParentBlockID uint64
	ParentBlock   Block
}

type RawBlock struct {
	BlockID   uint64
	Block     Block
	BlockData []byte
}

type Subnetwork struct {
	ID           uint64 `gorm:"primary_key"`
	SubnetworkID []byte
}

type Transaction struct {
	ID               uint64 `gorm:"primary_key"`
	AcceptingBlockID uint64
	AcceptingBlock   Block
	TransactionHash  string
	TransactionID    string
	LockTime         uint64
	SubnetworkID     uint64
	Gas              uint64
	PayloadHash      string
	Payload          []byte
	Blocks           []Block `gorm:"many2many:transactions_to_blocks;"`
}

type TransactionBlock struct {
	TransactionID uint64
	Transaction   Transaction
	BlockID       uint64
	Block         Block
	Index         uint32
}

func (TransactionBlock) TableName() string {
	return "transactions_to_blocks"
}

type TransactionOutput struct {
	TransactionID uint64
	Transaction   Transaction
	Index         uint32
	Value         uint64
	PkScript      []byte
}

type TransactionInput struct {
	TransactionID       uint64
	Transaction         Transaction
	TransactionOutputID uint64
	TransactionOutput   TransactionOutput
	Index               uint32
	SignatureScript     []byte
	Sequence            uint64
}

type UTXO struct {
	TransactionOutputID uint64
	TransactionOutput   TransactionOutput
	AcceptingBlockID    uint64
	AcceptingBlock      Block
}
