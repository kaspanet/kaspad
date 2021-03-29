package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// UTXODiffStore represents a store of UTXODiffs
type UTXODiffStore interface {
	Store
	Stage(stagingArea *StagingArea, blockHash *externalapi.DomainHash, utxoDiff externalapi.UTXODiff, utxoDiffChild *externalapi.DomainHash)
	IsStaged(stagingArea *StagingArea) bool
	UTXODiff(dbContext DBReader, stagingArea *StagingArea, blockHash *externalapi.DomainHash) (externalapi.UTXODiff, error)
	UTXODiffChild(dbContext DBReader, stagingArea *StagingArea, blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error)
	HasUTXODiffChild(dbContext DBReader, stagingArea *StagingArea, blockHash *externalapi.DomainHash) (bool, error)
	Delete(stagingArea *StagingArea, blockHash *externalapi.DomainHash)
}
