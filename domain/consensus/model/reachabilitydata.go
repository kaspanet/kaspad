package model

// ReachabilityData holds the set of data required to answer
// reachability queries
type ReachabilityData struct {
	treeNode          *reachabilityTreeNode
	futureCoveringSet futureCoveringTreeNodeSet
}

// reachabilityTreeNode represents a node in the reachability tree
// of some DAG block. It mainly provides the ability to query *tree*
// reachability with O(1) query time. It does so by managing an
// index interval for each node and making sure all nodes in its
// subtree are indexed within the interval, so the query
// B ∈ subtree(A) simply becomes B.interval ⊂ A.interval.
//
// The main challenge of maintaining such intervals is that our tree
// is an ever-growing tree and as such pre-allocated intervals may
// not suffice as per future events. This is where the reindexing
// algorithm below comes into place.
// We use the reasonable assumption that the initial root interval
// (e.g., [0, 2^64-1]) should always suffice for any practical use-
// case, and so reindexing should always succeed unless more than
// 2^64 blocks are added to the DAG/tree.
type reachabilityTreeNode struct {
	children []*reachabilityTreeNode
	parent   *reachabilityTreeNode

	// interval is the index interval containing all intervals of
	// blocks in this node's subtree
	interval *reachabilityInterval
}

// reachabilityInterval represents an interval to be used within the
// tree reachability algorithm. See reachabilityTreeNode for further
// details.
type reachabilityInterval struct {
	start uint64
	end   uint64
}

// orderedTreeNodeSet is an ordered set of reachabilityTreeNodes
// Note that this type does not validate order validity. It's the
// responsibility of the caller to construct instances of this
// type properly.
type orderedTreeNodeSet []*reachabilityTreeNode

// futureCoveringTreeNodeSet represents a collection of blocks in the future of
// a certain block. Once a block B is added to the DAG, every block A_i in
// B's selected parent anticone must register B in its futureCoveringTreeNodeSet. This allows
// to relatively quickly (O(log(|futureCoveringTreeNodeSet|))) query whether B
// is a descendent (is in the "future") of any block that previously
// registered it.
//
// Note that futureCoveringTreeNodeSet is meant to be queried only if B is not
// a reachability tree descendant of the block in question, as reachability
// tree queries are always O(1).
//
// See insertNode, hasAncestorOf, and reachabilityTree.isInPast for further
// details.
type futureCoveringTreeNodeSet orderedTreeNodeSet
