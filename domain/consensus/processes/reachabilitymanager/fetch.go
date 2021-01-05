package reachabilitymanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/reachabilitydata"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/pkg/errors"
)

func (rt *reachabilityManager) reachabilityDataForInsertion(
	blockHash *externalapi.DomainHash) (model.MutableReachabilityData, error) {
	data, err := rt.reachabilityDataStore.ReachabilityData(rt.databaseContext, blockHash)
	if err == nil {
		return data.CloneMutable(), nil
	}

	if errors.Is(err, database.ErrNotFound) {
		return reachabilitydata.EmptyReachabilityData(), nil
	}
	return nil, err
}

func (rt *reachabilityManager) futureCoveringSet(blockHash *externalapi.DomainHash) (model.FutureCoveringTreeNodeSet, error) {
	data, err := rt.reachabilityDataStore.ReachabilityData(rt.databaseContext, blockHash)
	if err != nil {
		return nil, err
	}

	return data.FutureCoveringSet(), nil
}

func (rt *reachabilityManager) interval(blockHash *externalapi.DomainHash) (*model.ReachabilityInterval, error) {
	interval, err := rt.reachabilityDataStore.Interval(rt.databaseContext, blockHash)
	if err != nil {
		return nil, err
	}

	return interval, nil
}

func (rt *reachabilityManager) children(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	data, err := rt.reachabilityDataStore.ReachabilityData(rt.databaseContext, blockHash)
	if err != nil {
		return nil, err
	}

	return data.Children(), nil
}

func (rt *reachabilityManager) parent(blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	data, err := rt.reachabilityDataStore.ReachabilityData(rt.databaseContext, blockHash)
	if err != nil {
		return nil, err
	}

	return data.Parent(), nil
}

func (rt *reachabilityManager) reindexRoot() (*externalapi.DomainHash, error) {
	return rt.reachabilityDataStore.ReachabilityReindexRoot(rt.databaseContext)
}
