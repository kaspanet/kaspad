package models

import (
	"time"
)

// Block is the gorm model for the 'blocks' table
type Block struct {
	ID                   uint64 `gorm:"primary_key"`
	BlockHash            string
	AcceptingBlockID     *uint64
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
	Mass                 uint64
	ParentBlocks         []Block `gorm:"many2many:parent_blocks;"`
}

// ParentBlock is the gorm model for the 'parent_blocks' table
type ParentBlock struct {
	BlockID       uint64
	Block         Block
	ParentBlockID uint64
	ParentBlock   Block
}

// RawBlock is the gorm model for the 'raw_blocks' table
type RawBlock struct {
	BlockID   uint64
	Block     Block
	BlockData []byte
}

// Subnetwork is the gorm model for the 'subnetworks' table
type Subnetwork struct {
	ID           uint64 `gorm:"primary_key"`
	SubnetworkID string
	GasLimit     *uint64
}

// Transaction is the gorm model for the 'transactions' table
type Transaction struct {
	ID                 uint64 `gorm:"primary_key"`
	AcceptingBlockID   *uint64
	AcceptingBlock     *Block
	TransactionHash    string
	TransactionID      string
	LockTime           uint64
	SubnetworkID       uint64
	Subnetwork         Subnetwork
	Gas                uint64
	PayloadHash        string
	Payload            []byte
	Mass               uint64
	Blocks             []Block `gorm:"many2many:transactions_to_blocks;"`
	TransactionOutputs []TransactionOutput
	TransactionInputs  []TransactionInput
}

// TransactionBlock is the gorm model for the 'transactions_to_blocks' table
type TransactionBlock struct {
	TransactionID uint64
	Transaction   Transaction
	BlockID       uint64
	Block         Block
	Index         uint32
}

// TableName returns the table name associated to the
// TransactionBlock gorm model
func (TransactionBlock) TableName() string {
	return "transactions_to_blocks"
}

// TransactionOutput is the gorm model for the 'transaction_outputs' table
type TransactionOutput struct {
	ID            uint64 `gorm:"primary_key"`
	TransactionID uint64
	Transaction   Transaction
	Index         uint32
	Value         uint64
	ScriptPubKey  []byte
	IsSpent       bool
	AddressID     uint64
	Address       Address
}

// TransactionInput is the gorm model for the 'transaction_inputs' table
type TransactionInput struct {
	ID                  uint64 `gorm:"primary_key"`
	TransactionID       uint64
	Transaction         Transaction
	TransactionOutputID uint64
	TransactionOutput   TransactionOutput
	Index               uint32
	SignatureScript     []byte
	Sequence            uint64
}

// Address is the gorm model for the 'utxos' table
type Address struct {
	ID      uint64 `gorm:"primary_key"`
	Address string
}
