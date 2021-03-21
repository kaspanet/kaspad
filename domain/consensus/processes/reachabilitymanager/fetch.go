package reachabilitymanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/reachabilitydata"
	"github.com/pkg/errors"
)

func (rt *reachabilityManager) reachabilityDataForInsertion(stagingArea *model.StagingArea,
	blockHash *externalapi.DomainHash) (model.MutableReachabilityData, error) {
	data, err := rt.reachabilityDataStore.ReachabilityData(rt.databaseContext, stagingArea, blockHash)
	if err == nil {
		return data.CloneMutable(), nil
	}

	if errors.Is(err, database.ErrNotFound) {
		return reachabilitydata.EmptyReachabilityData(), nil
	}
	return nil, err
}

func (rt *reachabilityManager) futureCoveringSet(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (model.FutureCoveringTreeNodeSet, error) {
	data, err := rt.reachabilityDataStore.ReachabilityData(rt.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	return data.FutureCoveringSet(), nil
}

func (rt *reachabilityManager) interval(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (*model.ReachabilityInterval, error) {
	data, err := rt.reachabilityDataStore.ReachabilityData(rt.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	return data.Interval(), nil
}

func (rt *reachabilityManager) children(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (
	[]*externalapi.DomainHash, error) {

	data, err := rt.reachabilityDataStore.ReachabilityData(rt.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	return data.Children(), nil
}

func (rt *reachabilityManager) parent(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (
	*externalapi.DomainHash, error) {

	data, err := rt.reachabilityDataStore.ReachabilityData(rt.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	return data.Parent(), nil
}

func (rt *reachabilityManager) reindexRoot(stagingArea *model.StagingArea) (*externalapi.DomainHash, error) {
	return rt.reachabilityDataStore.ReachabilityReindexRoot(rt.databaseContext, stagingArea)
}
