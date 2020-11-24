package externalapi

// DomainBlock represents a Kaspa block
type DomainBlock struct {
	Header       *DomainBlockHeader
	Transactions []*DomainTransaction
}

// Clone returns a clone of DomainBlock
func (block *DomainBlock) Clone() *DomainBlock {
	if block == nil {
		return nil
	}

	transactionClone := make([]*DomainTransaction, len(block.Transactions))
	for i, tx := range block.Transactions {
		transactionClone[i] = tx.Clone()
	}

	return &DomainBlock{
		Header:       block.Header.Clone(),
		Transactions: transactionClone,
	}
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
	if header == nil {
		return nil
	}

	return &DomainBlockHeader{
		Version:              header.Version,
		ParentHashes:         CloneHashes(header.ParentHashes),
		HashMerkleRoot:       *header.HashMerkleRoot.Clone(),
		AcceptedIDMerkleRoot: *header.AcceptedIDMerkleRoot.Clone(),
		UTXOCommitment:       *header.UTXOCommitment.Clone(),
		TimeInMilliseconds:   header.TimeInMilliseconds,
		Bits:                 header.Bits,
		Nonce:                header.Nonce,
	}
}
