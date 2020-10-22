package reachabilitytree

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func (rt *reachabilityTreeManager) stageFutureCoveringSet(blockHash *externalapi.DomainHash, node model.FutureCoveringTreeNodeSet) error {
	panic("unimplemented")
}

func (rt *reachabilityTreeManager) stageTreeNode(blockHash *externalapi.DomainHash, node *model.ReachabilityTreeNode) error {
	panic("unimplemented")
}

func (rt *reachabilityTreeManager) stageReindexRoot(blockHash *externalapi.DomainHash) {
	panic("unimplemented")
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
