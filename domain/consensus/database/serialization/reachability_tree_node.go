package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

func reachablityTreeNodeToDBReachablityTreeNode(reachabilityTreeNode *model.ReachabilityTreeNode) *DbReachabilityTreeNode {
	return &DbReachabilityTreeNode{
		Children: DomainHashesToDbHashes(reachabilityTreeNode.Children),
		Parent:   DomainHashToDbHash(reachabilityTreeNode.Parent),
		Interval: reachablityIntervalToDBReachablityInterval(reachabilityTreeNode.Interval),
	}
}

func dbReachablityTreeNodeToReachablityTreeNode(dbReachabilityTreeNode *DbReachabilityTreeNode) (*model.ReachabilityTreeNode, error) {
	children, err := DbHashesToDomainHashes(dbReachabilityTreeNode.Children)
	if err != nil {
		return nil, err
	}

	parent, err := DbHashToDomainHash(dbReachabilityTreeNode.Parent)
	if err != nil {
		return nil, err
	}

	return &model.ReachabilityTreeNode{
		Children: children,
		Parent:   parent,
		Interval: dbReachablityIntervalToReachablityInterval(dbReachabilityTreeNode.Interval),
	}, nil
}
