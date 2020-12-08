package reachabilitymanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func (rt *reachabilityManager) data(blockHash *externalapi.DomainHash) (*model.ReachabilityData, error) {
	hasData, err := rt.reachabilityDataStore.HasReachabilityData(rt.databaseContext, blockHash)
	if err != nil {
		return nil, err
	}

	if !hasData {
		return &model.ReachabilityData{
			TreeNode:          nil,
			FutureCoveringSet: nil,
		}, nil
	}

	return rt.reachabilityDataStore.ReachabilityData(rt.databaseContext, blockHash)
}

func (rt *reachabilityManager) futureCoveringSet(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	data, err := rt.data(blockHash)
	if err != nil {
		return nil, err
	}

	return data.FutureCoveringSet, nil
}

func (rt *reachabilityManager) treeNode(blockHash *externalapi.DomainHash) (*model.ReachabilityTreeNode, error) {
	data, err := rt.data(blockHash)
	if err != nil {
		return nil, err
	}

	return data.TreeNode, nil
}

func (rt *reachabilityManager) interval(blockHash *externalapi.DomainHash) (*model.ReachabilityInterval, error) {
	treeNode, err := rt.treeNode(blockHash)
	if err != nil {
		return nil, err
	}

	return treeNode.Interval, nil
}

func (rt *reachabilityManager) children(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	data, err := rt.data(blockHash)
	if err != nil {
		return nil, err
	}

	return data.TreeNode.Children, nil
}

func (rt *reachabilityManager) parent(blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	data, err := rt.data(blockHash)
	if err != nil {
		return nil, err
	}

	return data.TreeNode.Parent, nil
}

func (rt *reachabilityManager) reindexRoot() (*externalapi.DomainHash, error) {
	return rt.reachabilityDataStore.ReachabilityReindexRoot(rt.databaseContext)
}
