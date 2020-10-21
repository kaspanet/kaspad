package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// GHOSTDAGDataStore represents a store of BlockGHOSTDAGData
type GHOSTDAGDataStore interface {
	Stage(blockHash *externalapi.DomainHash, blockGHOSTDAGData *BlockGHOSTDAGData)
	IsStaged() bool
	Discard()
	Commit(dbTx DBTxProxy) error
	Get(dbContext DBContextProxy, blockHash *externalapi.DomainHash) (*BlockGHOSTDAGData, error)
}
