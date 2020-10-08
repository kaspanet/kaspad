package model

// ReachabilityChangeset holds the set of changes to make to a
// reachability tree to insert a new reachability node
type ReachabilityChangeset struct {
	TreeNodeChanges          map[*DomainHash]*ReachabilityTreeNode
	FutureCoveringSetChanges map[*DomainHash]FutureCoveringTreeNodeSet
}
