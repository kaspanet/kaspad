package externalapi

// AcceptanceData stores data about which transactions were accepted by a block.
// It's ordered in the same way as the block merge set blues.
type AcceptanceData []*BlockAcceptanceData

// Clone clones the AcceptanceData
func (ad AcceptanceData) Clone() AcceptanceData {
	if ad == nil {
		return nil
	}
	clone := make(AcceptanceData, len(ad))
	for i, blockAcceptanceData := range ad {
		clone[i] = blockAcceptanceData.Clone()
	}

	return clone
}

// BlockAcceptanceData stores all transactions in a block with an indication
// if they were accepted or not by some other block
type BlockAcceptanceData struct {
	TransactionAcceptanceData []*TransactionAcceptanceData
}

// Clone returns a clone of BlockAcceptanceData
func (bad *BlockAcceptanceData) Clone() *BlockAcceptanceData {
	if bad == nil {
		return nil
	}

	clone := &BlockAcceptanceData{TransactionAcceptanceData: make([]*TransactionAcceptanceData, len(bad.TransactionAcceptanceData))}
	for i, acceptanceData := range bad.TransactionAcceptanceData {
		clone.TransactionAcceptanceData[i] = acceptanceData.Clone()
	}

	return clone
}

// TransactionAcceptanceData stores a transaction together with an indication
// if it was accepted or not by some block
type TransactionAcceptanceData struct {
	Transaction *DomainTransaction
	Fee         uint64
	IsAccepted  bool
}

// Clone returns a clone of TransactionAcceptanceData
func (tad *TransactionAcceptanceData) Clone() *TransactionAcceptanceData {
	if tad == nil {
		return nil
	}

	return &TransactionAcceptanceData{
		Transaction: tad.Transaction.Clone(),
		Fee:         tad.Fee,
		IsAccepted:  tad.IsAccepted,
	}
}
