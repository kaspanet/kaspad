package reachabilitymanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

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
// See insertNode, hasAncestorOf, and isInPast for further
// details.
type futureCoveringTreeNodeSet orderedTreeNodeSet

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
	futureCoveringSet, err := rt.futureCoveringSet(node)
	if err != nil {
		return err
	}

	ancestorIndex, ok, err := rt.findAncestorIndexOfNode(futureCoveringSet, futureNode)
	if err != nil {
		return err
	}

	if !ok {
		newSet := append([]*externalapi.DomainHash{futureNode}, futureCoveringSet...)
		err := rt.stageFutureCoveringSet(node, newSet)
		if err != nil {
			return err
		}

		return nil
	}

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
	newSet := append(left, right...)
	return rt.stageFutureCoveringSet(node, newSet)

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

	ancestorIndex, ok, err := rt.findAncestorIndexOfNode(futureCoveringSet, other)
	if err != nil {
		return false, err
	}

	if !ok {
		// No candidate to contain other
		return false, nil
	}

	candidate := futureCoveringSet[ancestorIndex]
	return rt.IsReachabilityTreeAncestorOf(candidate, other)
}

// futureCoveringSetString returns a string representation of the intervals in this futureCoveringSet.
func (rt *reachabilityManager) futureCoveringSetString(futureCoveringSet []*externalapi.DomainHash) (string, error) {
	intervalsString := ""
	for _, node := range futureCoveringSet {
		nodeInterval, err := rt.interval(node)
		if err != nil {
			return "", err
		}

		intervalsString += intervalString(nodeInterval)
	}
	return intervalsString, nil
}
