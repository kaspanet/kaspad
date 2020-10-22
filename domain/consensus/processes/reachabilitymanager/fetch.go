package reachabilitymanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func (rt *reachabilityTreeManager) data(blockHash *externalapi.DomainHash) (*model.ReachabilityData, error) {
	return rt.reachabilityDataStore.ReachabilityData(rt.databaseContext, blockHash)
}

func (rt *reachabilityTreeManager) futureCoveringSet(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	data, err := rt.data(blockHash)
	if err != nil {
		return nil, err
	}

	return data.FutureCoveringSet, nil
}

func (rt *reachabilityTreeManager) treeNode(blockHash *externalapi.DomainHash) (*model.ReachabilityTreeNode, error) {
	data, err := rt.data(blockHash)
	if err != nil {
		return nil, err
	}

	return data.TreeNode, nil
}

func (rt *reachabilityTreeManager) interval(blockHash *externalapi.DomainHash) (*model.ReachabilityInterval, error) {
	treeNode, err := rt.treeNode(blockHash)
	if err != nil {
		return nil, err
	}

	return treeNode.Interval, nil
}

func (rt *reachabilityTreeManager) children(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	data, err := rt.data(blockHash)
	if err != nil {
		return nil, err
	}

	return data.TreeNode.Children, nil
}

func (rt *reachabilityTreeManager) parent(blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	data, err := rt.data(blockHash)
	if err != nil {
		return nil, err
	}

	return data.TreeNode.Parent, nil
}

func (rt *reachabilityTreeManager) reindexRoot() (*externalapi.DomainHash, error) {
	return rt.reachabilityDataStore.ReachabilityReindexRoot(rt.databaseContext)
}
