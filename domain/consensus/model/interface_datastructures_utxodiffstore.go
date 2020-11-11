package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// UTXODiffStore represents a store of UTXODiffs
type UTXODiffStore interface {
	Store
	Stage(blockHash *externalapi.DomainHash, utxoDiff *UTXODiff, utxoDiffChild *externalapi.DomainHash) error
	IsStaged() bool
	UTXODiff(dbContext DBReader, blockHash *externalapi.DomainHash) (*UTXODiff, error)
	UTXODiffChild(dbContext DBReader, blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error)
	HasUTXODiffChild(dbContext DBReader, blockHash *externalapi.DomainHash) (bool, error)
	Delete(blockHash *externalapi.DomainHash)
}
