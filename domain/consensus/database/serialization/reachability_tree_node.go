package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func reachablityTreeNodeToDBReachablityTreeNode(reachabilityTreeNode *model.ReachabilityTreeNode) *DbReachabilityTreeNode {
	var parent *DbHash
	if reachabilityTreeNode.Parent != nil {
		parent = DomainHashToDbHash(reachabilityTreeNode.Parent)
	}

	return &DbReachabilityTreeNode{
		Children: DomainHashesToDbHashes(reachabilityTreeNode.Children),
		Parent:   parent,
		Interval: reachablityIntervalToDBReachablityInterval(reachabilityTreeNode.Interval),
	}
}

func dbReachablityTreeNodeToReachablityTreeNode(dbReachabilityTreeNode *DbReachabilityTreeNode) (*model.ReachabilityTreeNode, error) {
	children, err := DbHashesToDomainHashes(dbReachabilityTreeNode.Children)
	if err != nil {
		return nil, err
	}

	var parent *externalapi.DomainHash
	if dbReachabilityTreeNode.Parent != nil {
		var err error
		parent, err = DbHashToDomainHash(dbReachabilityTreeNode.Parent)
		if err != nil {
			return nil, err
		}
	}

	return &model.ReachabilityTreeNode{
		Children: children,
		Parent:   parent,
		Interval: dbReachablityIntervalToReachablityInterval(dbReachabilityTreeNode.Interval),
	}, nil
}
