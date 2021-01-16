package model

import (
	"fmt"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// MutableReachabilityData represents a node in the reachability tree
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
//
// In addition, we keep a future covering set for every node.
// This set allows to query reachability over the entirety of the DAG.
// See documentation of FutureCoveringTreeNodeSet for additional details.

// ReachabilityData is a read-only version of a block's MutableReachabilityData
// Use CloneWritable to edit the MutableReachabilityData.
type ReachabilityData interface {
	Children() []*externalapi.DomainHash
	Parent() *externalapi.DomainHash
	Interval() *ReachabilityInterval
	FutureCoveringSet() FutureCoveringTreeNodeSet
	CloneMutable() MutableReachabilityData
	Equal(other ReachabilityData) bool
}

// MutableReachabilityData represents a block's MutableReachabilityData, with ability to edit it
type MutableReachabilityData interface {
	ReachabilityData

	AddChild(child *externalapi.DomainHash)
	SetParent(parent *externalapi.DomainHash)
	SetInterval(interval *ReachabilityInterval)
	SetFutureCoveringSet(futureCoveringSet FutureCoveringTreeNodeSet)
}

// ReachabilityInterval represents an interval to be used within the
// tree reachability algorithm. See ReachabilityTreeNode for further
// details.
type ReachabilityInterval struct {
	Start uint64
	End   uint64
}

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal and Clone accordingly.
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

// Increase returns a ReachabilityInterval with offset added to start and end
func (ri *ReachabilityInterval) Increase(offset uint64) *ReachabilityInterval {
	return &ReachabilityInterval{
		Start: ri.Start + offset,
		End:   ri.End + offset,
	}
}

// Decrease returns a ReachabilityInterval with offset subtracted from start and end
func (ri *ReachabilityInterval) Decrease(offset uint64) *ReachabilityInterval {
	return &ReachabilityInterval{
		Start: ri.Start - offset,
		End:   ri.End - offset,
	}
}

// IncreaseStart returns a ReachabilityInterval with offset added to start
func (ri *ReachabilityInterval) IncreaseStart(offset uint64) *ReachabilityInterval {
	return &ReachabilityInterval{
		Start: ri.Start + offset,
		End:   ri.End,
	}
}

// DecreaseStart returns a ReachabilityInterval with offset reduced from start
func (ri *ReachabilityInterval) DecreaseStart(offset uint64) *ReachabilityInterval {
	return &ReachabilityInterval{
		Start: ri.Start - offset,
		End:   ri.End,
	}
}

// IncreaseEnd returns a ReachabilityInterval with offset added to end
func (ri *ReachabilityInterval) IncreaseEnd(offset uint64) *ReachabilityInterval {
	return &ReachabilityInterval{
		Start: ri.Start,
		End:   ri.End + offset,
	}
}

// DecreaseEnd returns a ReachabilityInterval with offset subtracted from end
func (ri *ReachabilityInterval) DecreaseEnd(offset uint64) *ReachabilityInterval {
	return &ReachabilityInterval{
		Start: ri.Start,
		End:   ri.End - offset,
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
// an indication to update Equal and Clone accordingly.
var _ FutureCoveringTreeNodeSet = []*externalapi.DomainHash{}

// Equal returns whether fctns equals to other
func (fctns FutureCoveringTreeNodeSet) Equal(other FutureCoveringTreeNodeSet) bool {
	return externalapi.HashesEqual(fctns, other)
}
