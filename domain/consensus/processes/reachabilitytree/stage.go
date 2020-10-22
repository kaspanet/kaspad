package reachabilitytree

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func (rt *reachabilityTreeManager) stageData(blockHash *externalapi.DomainHash, data *model.ReachabilityData) {
	rt.reachabilityDataStore.StageReachabilityData(blockHash, data)
}

func (rt *reachabilityTreeManager) stageFutureCoveringSet(blockHash *externalapi.DomainHash, set model.FutureCoveringTreeNodeSet) error {
	data, err := rt.data(blockHash)
	if err != nil {
		return err
	}

	data.FutureCoveringSet = set
	rt.reachabilityDataStore.StageReachabilityData(blockHash, data)
	return nil
}

func (rt *reachabilityTreeManager) stageTreeNode(blockHash *externalapi.DomainHash, node *model.ReachabilityTreeNode) error {
	data, err := rt.data(blockHash)
	if err != nil {
		return err
	}

	data.TreeNode = node
	rt.reachabilityDataStore.StageReachabilityData(blockHash, data)
	return nil
}

func (rt *reachabilityTreeManager) stageReindexRoot(blockHash *externalapi.DomainHash) {
	rt.reachabilityDataStore.StageReachabilityReindexRoot(blockHash)
}

func (rt *reachabilityTreeManager) addChildAndStage(node, child *externalapi.DomainHash) error {
	nodeData, err := rt.data(node)
	if err != nil {
		return err
	}

	nodeData.TreeNode.Children = append(nodeData.TreeNode.Children, child)
	return rt.stageTreeNode(node, nodeData.TreeNode)
}

func (rt *reachabilityTreeManager) stageParent(node, parent *externalapi.DomainHash) error {
	treeNode, err := rt.treeNode(node)
	if err != nil {
		return err
	}

	treeNode.Parent = parent
	return rt.stageTreeNode(node, treeNode)
}

func (rt *reachabilityTreeManager) stageInterval(node *externalapi.DomainHash, interval *model.ReachabilityInterval) error {
	treeNode, err := rt.treeNode(node)
	if err != nil {
		return err
	}

	treeNode.Interval = interval
	return rt.stageTreeNode(node, treeNode)
}
