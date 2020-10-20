package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// GHOSTDAGDataStore represents a store of BlockGHOSTDAGData
type GHOSTDAGDataStore interface {
	Insert(dbTx DBTxProxy, blockHash *externalapi.DomainHash, blockGHOSTDAGData *BlockGHOSTDAGData) error
	Get(dbContext DBContextProxy, blockHash *externalapi.DomainHash) (*BlockGHOSTDAGData, error)
}
