package externalapi

// AcceptanceData stores data about which transactions were accepted by a block.
// It's ordered in the same way as the block merge set blues.
type AcceptanceData []*BlockAcceptanceData

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal and Clone accordingly.
var _ AcceptanceData = []*BlockAcceptanceData{}

// Equal returns whether ad equals to other
func (ad AcceptanceData) Equal(other AcceptanceData) bool {
	if len(ad) != len(other) {
		return false
	}

	for i, blockAcceptanceData := range ad {
		if !blockAcceptanceData.Equal(other[i]) {
			return false
		}
	}

	return true
}

// Clone clones the AcceptanceData
func (ad AcceptanceData) Clone() AcceptanceData {
	clone := make(AcceptanceData, len(ad))
	for i, blockAcceptanceData := range ad {
		clone[i] = blockAcceptanceData.Clone()
	}

	return clone
}

// BlockAcceptanceData stores all transactions in a block with an indication
// if they were accepted or not by some other block
type BlockAcceptanceData struct {
	BlockHash                 *DomainHash
	TransactionAcceptanceData []*TransactionAcceptanceData
}

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal and Clone accordingly.
var _ = &BlockAcceptanceData{&DomainHash{}, []*TransactionAcceptanceData{}}

// Equal returns whether bad equals to other
func (bad *BlockAcceptanceData) Equal(other *BlockAcceptanceData) bool {
	if bad == nil || other == nil {
		return bad == other
	}

	if !bad.BlockHash.Equal(other.BlockHash) {
		return false
	}

	if len(bad.TransactionAcceptanceData) != len(other.TransactionAcceptanceData) {
		return false
	}

	for i, acceptanceData := range bad.TransactionAcceptanceData {
		if !acceptanceData.Equal(other.TransactionAcceptanceData[i]) {
			return false
		}
	}

	return true
}

// Clone returns a clone of BlockAcceptanceData
func (bad *BlockAcceptanceData) Clone() *BlockAcceptanceData {
	if bad == nil {
		return nil
	}

	clone := &BlockAcceptanceData{
		BlockHash:                 bad.BlockHash,
		TransactionAcceptanceData: make([]*TransactionAcceptanceData, len(bad.TransactionAcceptanceData)),
	}
	for i, acceptanceData := range bad.TransactionAcceptanceData {
		clone.TransactionAcceptanceData[i] = acceptanceData.Clone()
	}

	return clone
}

// TransactionAcceptanceData stores a transaction together with an indication
// if it was accepted or not by some block
type TransactionAcceptanceData struct {
	Transaction                 *DomainTransaction
	Fee                         uint64
	IsAccepted                  bool
	TransactionInputUTXOEntries []UTXOEntry
}

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal and Clone accordingly.
var _ = &TransactionAcceptanceData{&DomainTransaction{}, 0, false, []UTXOEntry{}}

// Equal returns whether tad equals to other
func (tad *TransactionAcceptanceData) Equal(other *TransactionAcceptanceData) bool {
	if tad == nil || other == nil {
		return tad == other
	}

	if !tad.Transaction.Equal(other.Transaction) {
		return false
	}

	if tad.Fee != other.Fee {
		return false
	}

	if tad.IsAccepted != other.IsAccepted {
		return false
	}

	if len(tad.TransactionInputUTXOEntries) != len(other.TransactionInputUTXOEntries) {
		return false
	}

	for i, thisUTXOEntry := range tad.TransactionInputUTXOEntries {
		otherUTXOEntry := other.TransactionInputUTXOEntries[i]
		if !thisUTXOEntry.Equal(otherUTXOEntry) {
			return false
		}
	}

	return true
}

// Clone returns a clone of TransactionAcceptanceData
func (tad *TransactionAcceptanceData) Clone() *TransactionAcceptanceData {
	cloneTransactionInputUTXOEntries := make([]UTXOEntry, len(tad.TransactionInputUTXOEntries))
	for i, utxoEntry := range tad.TransactionInputUTXOEntries {
		cloneTransactionInputUTXOEntries[i] = utxoEntry
	}

	return &TransactionAcceptanceData{
		Transaction:                 tad.Transaction.Clone(),
		Fee:                         tad.Fee,
		IsAccepted:                  tad.IsAccepted,
		TransactionInputUTXOEntries: cloneTransactionInputUTXOEntries,
	}
}
