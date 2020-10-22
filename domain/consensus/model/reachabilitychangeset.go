package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ReachabilityChangeset holds the set of changes to make to a
// reachability tree to insert a new reachability node
type ReachabilityChangeset struct {
	TreeNodeChanges          map[externalapi.DomainHash]*ReachabilityTreeNode
	FutureCoveringSetChanges map[externalapi.DomainHash]FutureCoveringTreeNodeSet
	NewReindexRoot           *externalapi.DomainHash
}
