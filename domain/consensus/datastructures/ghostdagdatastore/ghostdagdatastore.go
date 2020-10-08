package ghostdagdatastore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// GHOSTDAGDataStore represents a store of BlockGHOSTDAGData
type GHOSTDAGDataStore struct {
}

// New instantiates a new GHOSTDAGDataStore
func New() *GHOSTDAGDataStore {
	return &GHOSTDAGDataStore{}
}

// Insert inserts the given blockGHOSTDAGData for the given blockHash
func (gds *GHOSTDAGDataStore) Insert(dbTx model.TxContextProxy, blockHash *model.DomainHash, blockGHOSTDAGData *model.BlockGHOSTDAGData) {

}

// Get gets the blockGHOSTDAGData associated with the given blockHash
func (gds *GHOSTDAGDataStore) Get(dbContext model.ContextProxy, blockHash *model.DomainHash) *model.BlockGHOSTDAGData {
	return nil
}
