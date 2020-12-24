package externalapi

// DomainBlock represents a Kaspa block
type DomainBlock struct {
	Header       *DomainBlockHeader
	Transactions []*DomainTransaction
}

// Clone returns a clone of DomainBlock
func (block *DomainBlock) Clone() *DomainBlock {
	transactionClone := make([]*DomainTransaction, len(block.Transactions))
	for i, tx := range block.Transactions {
		transactionClone[i] = tx.Clone()
	}

	return &DomainBlock{
		Header:       block.Header.Clone(),
		Transactions: transactionClone,
	}
}

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal and Clone accordingly.
var _ = DomainBlock{&DomainBlockHeader{}, []*DomainTransaction{}}

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

// Clone returns a clone of DomainBlockHeader
func (header *DomainBlockHeader) Clone() *DomainBlockHeader {
	return &DomainBlockHeader{
		Version:              header.Version,
		ParentHashes:         CloneHashes(header.ParentHashes),
		HashMerkleRoot:       header.HashMerkleRoot,
		AcceptedIDMerkleRoot: header.AcceptedIDMerkleRoot,
		UTXOCommitment:       header.UTXOCommitment,
		TimeInMilliseconds:   header.TimeInMilliseconds,
		Bits:                 header.Bits,
		Nonce:                header.Nonce,
	}
}

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal and Clone accordingly.
var _ = &DomainBlockHeader{0, []*DomainHash{}, DomainHash{},
	DomainHash{}, DomainHash{}, 0, 0, 0}

// Equal returns whether header equals to other
func (header *DomainBlockHeader) Equal(other *DomainBlockHeader) bool {
	if header == nil || other == nil {
		return header == other
	}

	if header.Version != other.Version {
		return false
	}

	if !HashesEqual(header.ParentHashes, other.ParentHashes) {
		return false
	}

	if !header.HashMerkleRoot.Equal(&other.HashMerkleRoot) {
		return false
	}

	if !header.AcceptedIDMerkleRoot.Equal(&other.AcceptedIDMerkleRoot) {
		return false
	}

	if !header.UTXOCommitment.Equal(&other.UTXOCommitment) {
		return false
	}

	if header.TimeInMilliseconds != other.TimeInMilliseconds {
		return false
	}

	if header.Bits != other.Bits {
		return false
	}

	if header.Nonce != other.Nonce {
		return false
	}

	return true
}
