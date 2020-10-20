package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// UTXODiffStore represents a store of UTXODiffs
type UTXODiffStore interface {
	Insert(dbTx DBTxProxy, blockHash *externalapi.DomainHash, utxoDiff *UTXODiff, utxoDiffChild *externalapi.DomainHash) error
	UTXODiff(dbContext DBContextProxy, blockHash *externalapi.DomainHash) (*UTXODiff, error)
	UTXODiffChild(dbContext DBContextProxy, blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error)
	Delete(dbTx DBTxProxy, blockHash *externalapi.DomainHash) error
}
