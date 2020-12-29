package reachabilitymanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/reachabilitydata"
)

func (rt *reachabilityManager) reachabilityDataForInsertion(
	blockHash *externalapi.DomainHash) (model.ReachabilityData, error) {

	hasData, err := rt.reachabilityDataStore.HasReachabilityData(rt.databaseContext, blockHash)
	if err != nil {
		return nil, err
	}

	if !hasData {
		return reachabilitydata.EmptyReachabilityData(), nil
	}

	data, err := rt.reachabilityDataStore.ReachabilityData(rt.databaseContext, blockHash)
	if err != nil {
		return nil, err
	}
	return data.CloneWritable(), nil
}

func (rt *reachabilityManager) futureCoveringSet(blockHash *externalapi.DomainHash) (model.FutureCoveringTreeNodeSet, error) {
	data, err := rt.reachabilityDataStore.ReachabilityData(rt.databaseContext, blockHash)
	if err != nil {
		return nil, err
	}

	return data.FutureCoveringSet(), nil
}

func (rt *reachabilityManager) interval(blockHash *externalapi.DomainHash) (*model.ReachabilityInterval, error) {
	data, err := rt.reachabilityDataStore.ReachabilityData(rt.databaseContext, blockHash)
	if err != nil {
		return nil, err
	}

	return data.Interval(), nil
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
