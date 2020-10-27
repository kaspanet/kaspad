package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockAcceptanceData stores all transactions in a block with an indication
// if they were accepted or not by some other block
type BlockAcceptanceData struct {
	TransactionAcceptanceData []TransactionAcceptanceData
}

// TransactionAcceptanceData stores a transaction together with an indication
// if it was accepted or not by some block
type TransactionAcceptanceData struct {
	Transaction *externalapi.DomainTransaction
	Fee         uint64
	IsAccepted  bool
}
