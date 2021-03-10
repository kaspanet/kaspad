package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// FinalityManager provides method to validate that a block does not violate finality
type FinalityManager interface {
	VirtualFinalityPoint() (*externalapi.DomainHash, error)
	FinalityPoint(blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error)
}
