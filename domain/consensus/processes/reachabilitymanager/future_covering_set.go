package reachabilitymanager

import (
	"fmt"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// insertToFutureCoveringSet inserts the given block into this node's FutureCoveringSet
// while keeping it ordered by interval.
// If a block B ∈ node.FutureCoveringSet exists such that its interval
// contains block's interval, block need not be added. If block's
// interval contains B's interval, it replaces it.
//
// Notes:
// * Intervals never intersect unless one contains the other
//   (this follows from the tree structure and the indexing rule).
// * Since node.FutureCoveringSet is kept ordered, a binary search can be
//   used for insertion/queries.
// * Although reindexing may change a block's interval, the
//   is-superset relation will by definition
//   be always preserved.
func (rt *reachabilityManager) insertToFutureCoveringSet(node, futureNode *externalapi.DomainHash) error {
	reachabilityData, err := rt.reachabilityDataForInsertion(node)
	if err != nil {
		return err
	}
	futureCoveringSet := reachabilityData.FutureCoveringSet()

	ancestorIndex, ok, err := rt.findAncestorIndexOfNode(orderedTreeNodeSet(futureCoveringSet), futureNode)
	if err != nil {
		return err
	}

	var newSet []*externalapi.DomainHash
	if !ok {
		newSet = append([]*externalapi.DomainHash{futureNode}, futureCoveringSet...)
		err := rt.stageFutureCoveringSet(node, newSet)
		if err != nil {
			return err
		}
		return nil
	} else {
		candidate := futureCoveringSet[ancestorIndex]
		candidateIsAncestorOfFutureNode, err := rt.IsReachabilityTreeAncestorOf(candidate, futureNode)
		if err != nil {
			return err
		}

		if candidateIsAncestorOfFutureNode {
			// candidate is an ancestor of futureNode, no need to insert
			return nil
		}

		futureNodeIsAncestorOfCandidate, err := rt.IsReachabilityTreeAncestorOf(futureNode, candidate)
		if err != nil {
			return err
		}

		if futureNodeIsAncestorOfCandidate {
			// futureNode is an ancestor of candidate, and can thus replace it
			newSet := make([]*externalapi.DomainHash, len(futureCoveringSet))
			copy(newSet, futureCoveringSet)
			newSet[ancestorIndex] = futureNode

			return rt.stageFutureCoveringSet(node, newSet)
		}

		// Insert futureNode in the correct index to maintain futureCoveringTreeNodeSet as
		// a sorted-by-interval list.
		// Note that ancestorIndex might be equal to len(futureCoveringTreeNodeSet)
		left := futureCoveringSet[:ancestorIndex+1]
		right := append([]*externalapi.DomainHash{futureNode}, futureCoveringSet[ancestorIndex+1:]...)
		newSet = append(left, right...)
	}
	reachabilityData.SetFutureCoveringSet(newSet)
	rt.stageData(node, reachabilityData)

	return nil
}

// futureCoveringSetHasAncestorOf resolves whether the given node `other` is in the subtree of
// any node in this.FutureCoveringSet.
// See insertNode method for the complementary insertion behavior.
//
// Like the insert method, this method also relies on the fact that
// this.FutureCoveringSet is kept ordered by interval to efficiently perform a
// binary search over this.FutureCoveringSet and answer the query in
// O(log(|futureCoveringTreeNodeSet|)).
func (rt *reachabilityManager) futureCoveringSetHasAncestorOf(this, other *externalapi.DomainHash) (bool, error) {
	futureCoveringSet, err := rt.futureCoveringSet(this)
	if err != nil {
		return false, err
	}

	ancestorIndex, ok, err := rt.findAncestorIndexOfNode(orderedTreeNodeSet(futureCoveringSet), other)
	if err != nil {
		return false, err
	}

	if !ok {
		// No candidate to contain other
		return false, nil
	}

	candidate := futureCoveringSet[ancestorIndex]
	if candidate.String() == "975bc80003886bc362033a47175732b8c3404eb2a13ef59f99d921bf3496e4cf" {
		fmt.Printf("~~~~ HERE IT IS!!!")
	}
	return rt.IsReachabilityTreeAncestorOf(candidate, other)
}
