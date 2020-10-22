package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// AcceptanceManager manages transaction acceptance
// and related data
type AcceptanceManager interface {
	CalculateAcceptanceDataAndUTXOMultiset(blockHash *externalapi.DomainHash) error
}
