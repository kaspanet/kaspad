package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// ReachablityDataToDBReachablityData converts ReachabilityData to DbReachabilityData
func ReachablityDataToDBReachablityData(reachabilityData *model.ReachabilityData) *DbReachabilityData {
	return &DbReachabilityData{
		TreeNode:          reachablityTreeNodeToDBReachablityTreeNode(reachabilityData.TreeNode),
		FutureCoveringSet: DomainHashesToDbHashes(reachabilityData.FutureCoveringSet),
	}
}

// DBReachablityDataToReachablityData converts DbReachabilityData to ReachabilityData
func DBReachablityDataToReachablityData(dbReachabilityData *DbReachabilityData) *model.ReachabilityData {
	treeNode, err := dbReachablityTreeNodeToReachablityTreeNode(dbReachabilityData.TreeNode)
	if err != nil {
		return nil
	}

	futureCoveringSet, err := DbHashesToDomainHashes(dbReachabilityData.FutureCoveringSet)
	if err != nil {
		return nil
	}

	return &model.ReachabilityData{
		TreeNode:          treeNode,
		FutureCoveringSet: futureCoveringSet,
	}
}
