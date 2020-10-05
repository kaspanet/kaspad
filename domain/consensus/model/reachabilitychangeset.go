package model

import "github.com/kaspanet/kaspad/util/daghash"

// ReachabilityChangeset holds the set of changes to make to a
// reachability tree to insert a new reachability node
type ReachabilityChangeset struct {
	treeNodeChanges          map[*daghash.Hash]*reachabilityTreeNode
	futureCoveringSetChanges map[*daghash.Hash]futureCoveringTreeNodeSet
}
