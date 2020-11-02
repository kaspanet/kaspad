package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ReachabilityData holds the set of data required to answer
// reachability queries
type ReachabilityData struct {
	TreeNode          *ReachabilityTreeNode
	FutureCoveringSet FutureCoveringTreeNodeSet
}

// ReachabilityTreeNode represents a node in the reachability tree
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
type ReachabilityTreeNode struct {
	Children []*externalapi.DomainHash
	Parent   *externalapi.DomainHash

	// interval is the index interval containing all intervals of
	// blocks in this node's subtree
	Interval *ReachabilityInterval
}

// ReachabilityInterval represents an interval to be used within the
// tree reachability algorithm. See ReachabilityTreeNode for further
// details.
type ReachabilityInterval struct {
	Start uint64
	End   uint64
}

// FutureCoveringTreeNodeSet represents a collection of blocks in the future of
// a certain block. Once a block B is added to the DAG, every block A_i in
// B's selected parent anticone must register B in its FutureCoveringTreeNodeSet. This allows
// to relatively quickly (O(log(|FutureCoveringTreeNodeSet|))) query whether B
// is a descendent (is in the "future") of any block that previously
// registered it.
//
// Note that FutureCoveringTreeNodeSet is meant to be queried only if B is not
// a reachability tree descendant of the block in question, as reachability
// tree queries are always O(1).
//
// See insertNode, hasAncestorOf, and isInPast for further details.
type FutureCoveringTreeNodeSet []*externalapi.DomainHash
