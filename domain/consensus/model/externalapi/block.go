package externalapi

import "math/big"

// DomainBlock represents a Kaspa block
type DomainBlock struct {
	Header       BlockHeader
	Transactions []*DomainTransaction
}

// Clone returns a clone of DomainBlock
func (block *DomainBlock) Clone() *DomainBlock {
	transactionClone := make([]*DomainTransaction, len(block.Transactions))
	for i, tx := range block.Transactions {
		transactionClone[i] = tx.Clone()
	}

	return &DomainBlock{
		Header:       block.Header,
		Transactions: transactionClone,
	}
}

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal and Clone accordingly.
var _ = DomainBlock{nil, []*DomainTransaction{}}

// Equal returns whether block equals to other
func (block *DomainBlock) Equal(other *DomainBlock) bool {
	if block == nil || other == nil {
		return block == other
	}

	if len(block.Transactions) != len(other.Transactions) {
		return false
	}

	if !block.Header.Equal(other.Header) {
		return false
	}

	for i, tx := range block.Transactions {
		if !tx.Equal(other.Transactions[i]) {
			return false
		}
	}

	return true
}

// BlockHeader represents an immutable block header.
type BlockHeader interface {
	BaseBlockHeader
	ToMutable() MutableBlockHeader
}

// BaseBlockHeader represents the header part of a Kaspa block
type BaseBlockHeader interface {
	Version() uint16
	Parents() []BlockLevelParents
	ParentsAtLevel(level int) BlockLevelParents
	DirectParents() BlockLevelParents
	HashMerkleRoot() *DomainHash
	AcceptedIDMerkleRoot() *DomainHash
	UTXOCommitment() *DomainHash
	TimeInMilliseconds() int64
	Bits() uint32
	Nonce() uint64
	DAAScore() uint64
	BlueScore() uint64
	BlueWork() *big.Int
	PruningPoint() *DomainHash
	Equal(other BaseBlockHeader) bool
}

// MutableBlockHeader represents a block header that can be mutated, but only
// the fields that are relevant to mining (Nonce and TimeInMilliseconds).
type MutableBlockHeader interface {
	BaseBlockHeader
	ToImmutable() BlockHeader
	SetNonce(nonce uint64)
	SetTimeInMilliseconds(timeInMilliseconds int64)
}
