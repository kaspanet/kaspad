package model

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type FinalityStore interface {
	Store
	IsStaged() bool
	StageFinalityPoint(blockHash *externalapi.DomainHash, finalityPointHash *externalapi.DomainHash)
	FinalityPoint(dbContext DBReader, blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error)
}
