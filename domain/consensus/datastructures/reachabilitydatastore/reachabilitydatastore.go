package reachabilitydatastore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// reachabilityDataStore represents a store of ReachabilityData
type reachabilityDataStore struct {
}

// New instantiates a new ReachabilityDataStore
func New() model.ReachabilityDataStore {
	return &reachabilityDataStore{}
}

// StageReachabilityData stages the given reachabilityData for the given blockHash
func (rds *reachabilityDataStore) StageReachabilityData(blockHash *externalapi.DomainHash, reachabilityData *model.ReachabilityData) {
	panic("implement me")
}

// StageReachabilityReindexRoot stages the given reachabilityReindexRoot
func (rds *reachabilityDataStore) StageReachabilityReindexRoot(reachabilityReindexRoot *externalapi.DomainHash) {
	panic("implement me")
}

func (rds *reachabilityDataStore) IsAnythingStaged() bool {
	panic("implement me")
}

func (rds *reachabilityDataStore) Discard() {
	panic("implement me")
}

func (rds *reachabilityDataStore) Commit(dbTx model.DBTransaction) error {
	panic("implement me")
}

// ReachabilityData returns the reachabilityData associated with the given blockHash
func (rds *reachabilityDataStore) ReachabilityData(dbContext model.DBReader,
	blockHash *externalapi.DomainHash) (*model.ReachabilityData, error) {

	panic("implement me")
}

// ReachabilityReindexRoot returns the current reachability reindex root
func (rds *reachabilityDataStore) ReachabilityReindexRoot(dbContext model.DBReader) (*externalapi.DomainHash, error) {
	panic("implement me")
}
