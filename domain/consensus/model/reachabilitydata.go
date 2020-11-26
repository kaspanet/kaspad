package model

import (
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// ReachabilityData holds the set of data required to answer
// reachability queries
type ReachabilityData struct {
	TreeNode          *ReachabilityTreeNode
	FutureCoveringSet FutureCoveringTreeNodeSet
}

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal accordingly.
var _ = &ReachabilityData{&ReachabilityTreeNode{}, FutureCoveringTreeNodeSet{}}

// Equal returns whether rd equals to other
func (rd *ReachabilityData) Equal(other *ReachabilityData) bool {
	if rd == nil || other == nil {
		return rd == other
	}

	if !rd.TreeNode.Equal(other.TreeNode) {
		return false
	}

	if !rd.FutureCoveringSet.Equal(other.FutureCoveringSet) {
		return false
	}

	return true
}

// Clone returns a clone of ReachabilityData
func (rd *ReachabilityData) Clone() *ReachabilityData {
	return &ReachabilityData{
		TreeNode:          rd.TreeNode.Clone(),
		FutureCoveringSet: rd.FutureCoveringSet.Clone(),
	}
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

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal accordingly.
var _ = &ReachabilityTreeNode{[]*externalapi.DomainHash{}, &externalapi.DomainHash{},
	&ReachabilityInterval{}}

// Equal returns whether rtn equals to other
func (rtn *ReachabilityTreeNode) Equal(other *ReachabilityTreeNode) bool {
	if rtn == nil || other == nil {
		return rtn == other
	}

	if externalapi.HashesEqual(rtn.Children, other.Children) {
		return false
	}

	if !rtn.Parent.Equal(other.Parent) {
		return false
	}

	if !rtn.Interval.Equal(other.Interval) {
		return false
	}

	return true
}

// Clone returns a clone of ReachabilityTreeNode
func (rtn *ReachabilityTreeNode) Clone() *ReachabilityTreeNode {
	return &ReachabilityTreeNode{
		Children: externalapi.CloneHashes(rtn.Children),
		Parent:   rtn.Parent.Clone(),
		Interval: rtn.Interval.Clone(),
	}
}

// ReachabilityInterval represents an interval to be used within the
// tree reachability algorithm. See ReachabilityTreeNode for further
// details.
type ReachabilityInterval struct {
	Start uint64
	End   uint64
}

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal accordingly.
var _ = &ReachabilityInterval{0, 0}

// Equal returns whether ri equals to other
func (ri *ReachabilityInterval) Equal(other *ReachabilityInterval) bool {
	if ri == nil || other == nil {
		return ri == other
	}

	if ri.Start != other.Start {
		return false
	}

	if ri.End != other.End {
		return false
	}

	return true
}

// Clone returns a clone of ReachabilityInterval
func (ri *ReachabilityInterval) Clone() *ReachabilityInterval {
	return &ReachabilityInterval{
		Start: ri.Start,
		End:   ri.End,
	}
}

func (ri *ReachabilityInterval) String() string {
	return fmt.Sprintf("[%d,%d]", ri.Start, ri.End)
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

// Clone returns a clone of FutureCoveringTreeNodeSet
func (fctns FutureCoveringTreeNodeSet) Clone() FutureCoveringTreeNodeSet {
	return externalapi.CloneHashes(fctns)
}

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal accordingly.
var _ FutureCoveringTreeNodeSet = []*externalapi.DomainHash{}

// Equal returns whether fctns equals to other
func (fctns FutureCoveringTreeNodeSet) Equal(other FutureCoveringTreeNodeSet) bool {
	return externalapi.HashesEqual(fctns, other)
}
