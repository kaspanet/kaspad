package reachabilitymanager

import (
	"math"
	"strings"
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/utils/reachabilitydata"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/pkg/errors"
)

func newReachabilityTreeData() model.ReachabilityData {
	// Please see the comment above model.ReachabilityTreeNode to understand why
	// we use these initial values.
	interval := newReachabilityInterval(1, math.MaxUint64-1)
	data := reachabilitydata.EmptyReachabilityData()
	data.SetInterval(interval)

	return data
}

/*

Interval helper functions

*/

func (rt *reachabilityManager) intervalRangeForChildAllocation(node *externalapi.DomainHash) (*model.ReachabilityInterval, error) {
	interval, err := rt.interval(node)
	if err != nil {
		return nil, err
	}

	// We subtract 1 from the end of the range to prevent the node from allocating
	// the entire interval to its child, so its interval would *strictly* contain the interval of its child.
	return newReachabilityInterval(interval.Start, interval.End-1), nil
}

func (rt *reachabilityManager) remainingIntervalBefore(node *externalapi.DomainHash) (*model.ReachabilityInterval, error) {
	childrenRange, err := rt.intervalRangeForChildAllocation(node)
	if err != nil {
		return nil, err
	}

	children, err := rt.children(node)
	if err != nil {
		return nil, err
	}

	if len(children) == 0 {
		return childrenRange, nil
	}

	firstChildInterval, err := rt.interval(children[0])
	if err != nil {
		return nil, err
	}

	return newReachabilityInterval(childrenRange.Start, firstChildInterval.Start-1), nil
}

func (rt *reachabilityManager) remainingIntervalAfter(node *externalapi.DomainHash) (*model.ReachabilityInterval, error) {
	childrenRange, err := rt.intervalRangeForChildAllocation(node)
	if err != nil {
		return nil, err
	}

	children, err := rt.children(node)
	if err != nil {
		return nil, err
	}

	if len(children) == 0 {
		return childrenRange, nil
	}

	lastChildInterval, err := rt.interval(children[len(children)-1])
	if err != nil {
		return nil, err
	}

	return newReachabilityInterval(lastChildInterval.End+1, childrenRange.End), nil
}

func (rt *reachabilityManager) remainingSlackBefore(node *externalapi.DomainHash) (uint64, error) {
	interval, err := rt.remainingIntervalBefore(node)
	if err != nil {
		return 0, err
	}

	return intervalSize(interval), nil
}

func (rt *reachabilityManager) remainingSlackAfter(node *externalapi.DomainHash) (uint64, error) {
	interval, err := rt.remainingIntervalAfter(node)
	if err != nil {
		return 0, err
	}

	return intervalSize(interval), nil
}

func (rt *reachabilityManager) hasSlackIntervalBefore(node *externalapi.DomainHash) (bool, error) {
	interval, err := rt.remainingIntervalBefore(node)
	if err != nil {
		return false, err
	}

	return intervalSize(interval) > 0, nil
}

func (rt *reachabilityManager) hasSlackIntervalAfter(node *externalapi.DomainHash) (bool, error) {
	interval, err := rt.remainingIntervalAfter(node)
	if err != nil {
		return false, err
	}

	return intervalSize(interval) > 0, nil
}

/*

ReachabilityManager API functions

*/

// IsReachabilityTreeAncestorOf checks if this node is a reachability tree ancestor
// of the other node. Note that we use the graph theory convention
// here which defines that node is also an ancestor of itself.
func (rt *reachabilityManager) IsReachabilityTreeAncestorOf(node, other *externalapi.DomainHash) (bool, error) {
	nodeInterval, err := rt.interval(node)
	if err != nil {
		return false, err
	}

	otherInterval, err := rt.interval(other)
	if err != nil {
		return false, err
	}

	return intervalContains(nodeInterval, otherInterval), nil
}

// FindNextAncestor finds the reachability tree child
// of 'ancestor' which is also an ancestor of 'descendant'.
func (rt *reachabilityManager) FindNextAncestor(descendant, ancestor *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	childrenOfAncestor, err := rt.children(ancestor)
	if err != nil {
		return nil, err
	}

	nextAncestor, ok := rt.findAncestorOfNode(childrenOfAncestor, descendant)
	if !ok {
		return nil, errors.Errorf("ancestor is not an ancestor of descendant")
	}

	return nextAncestor, nil
}

// String returns a string representation of a reachability tree node
// and its children.
func (rt *reachabilityManager) String(node *externalapi.DomainHash) (string, error) {
	queue := []*externalapi.DomainHash{node}
	nodeInterval, err := rt.interval(node)
	if err != nil {
		return "", err
	}

	lines := []string{nodeInterval.String()}
	for len(queue) > 0 {
		var current *externalapi.DomainHash
		current, queue = queue[0], queue[1:]
		children, err := rt.children(current)
		if err != nil {
			return "", err
		}

		if len(children) == 0 {
			continue
		}

		line := ""
		for _, child := range children {
			childInterval, err := rt.interval(child)
			if err != nil {
				return "", err
			}

			line += childInterval.String()
			queue = append(queue, child)
		}
		lines = append([]string{line}, lines...)
	}
	return strings.Join(lines, "\n"), nil
}

/*

Tree helper functions

*/

func (rt *reachabilityManager) isStrictAncestorOf(node, other *externalapi.DomainHash) (bool, error) {
	if node.Equal(other) {
		return false, nil
	}
	return rt.IsReachabilityTreeAncestorOf(node, other)
}

// findCommonAncestor finds the most recent reachability
// tree ancestor common to both node and the given reindex root. Note
// that we assume that almost always the chain between the reindex root
// and the common ancestor is longer than the chain between node and the
// common ancestor.
func (rt *reachabilityManager) findCommonAncestor(node, root *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	current := node
	for {
		isAncestorOf, err := rt.IsReachabilityTreeAncestorOf(current, root)
		if err != nil {
			return nil, err
		}

		if isAncestorOf {
			return current, nil
		}

		current, err = rt.parent(current)
		if err != nil {
			return nil, err
		}
	}
}

// splitChildren splits `node` children into two slices: the nodes that are before
// `pivot` and the nodes that are after.
func (rt *reachabilityManager) splitChildren(node, pivot *externalapi.DomainHash) (
	nodesBeforePivot, nodesAfterPivot []*externalapi.DomainHash, err error) {

	children, err := rt.children(node)
	if err != nil {
		return nil, nil, err
	}

	for i, child := range children {
		if child.Equal(pivot) {
			return children[:i], children[i+1:], nil
		}
	}
	return nil, nil, errors.Errorf("pivot not a pivot of node")
}

/*

Internal reachabilityManager API

*/

// addChild adds child to this tree node. If this node has no
// remaining interval to allocate, a reindexing is triggered. When a reindexing
// is triggered, the reindex root point is used within the
// reindex algorithm's logic
func (rt *reachabilityManager) addChild(node, child, reindexRoot *externalapi.DomainHash) error {
	remaining, err := rt.remainingIntervalAfter(node)
	if err != nil {
		return err
	}

	// Set the parent-child relationship
	err = rt.stageAddChild(node, child)
	if err != nil {
		return err
	}

	err = rt.stageParent(child, node)
	if err != nil {
		return err
	}

	// No allocation space left at parent -- reindex
	if intervalSize(remaining) == 0 {

		// Initially set the child's interval to the empty remaining interval.
		// This is done since in some cases, the underlying algorithm will
		// allocate space around this point and call intervalIncreaseEnd or
		// intervalDecreaseStart making for intervalSize > 0
		err = rt.stageInterval(child, remaining)
		if err != nil {
			return err
		}

		rc := newReindexContext(rt)

		reindexStartTime := time.Now()
		err := rc.reindexIntervals(child, reindexRoot)
		if err != nil {
			return err
		}

		reindexTimeElapsed := time.Since(reindexStartTime)
		log.Debugf("Reachability reindex triggered for "+
			"block %s. Took %dms.",
			node, reindexTimeElapsed.Milliseconds())

		return nil
	}

	// Allocate from the remaining space
	allocated, _, err := intervalSplitInHalf(remaining)
	if err != nil {
		return err
	}

	return rt.stageInterval(child, allocated)
}

func (rt *reachabilityManager) updateReindexRoot(selectedTip *externalapi.DomainHash) error {

	currentReindexRoot, err := rt.reindexRoot()
	if err != nil {
		return err
	}

	// First, find the new root
	reindexRootAncestor, newReindexRoot, err := rt.findNextReindexRoot(currentReindexRoot, selectedTip)
	if err != nil {
		return err
	}

	// No update to root, return
	if currentReindexRoot.Equal(newReindexRoot) {
		return nil
	}

	rc := newReindexContext(rt)

	if !newReindexRoot.Equal(reindexRootAncestor) {
		log.Debugf("Concentrating the intervals towards the new reindex root")
		// Iterate from reindexRootAncestor towards newReindexRoot
		for {
			chosenChild, err := rt.FindNextAncestor(selectedTip, reindexRootAncestor)
			if err != nil {
				return err
			}

			isFinalReindexRoot := chosenChild.Equal(newReindexRoot)

			// Concentrate interval from current ancestor to its chosen child
			err = rc.concentrateInterval(reindexRootAncestor, chosenChild, isFinalReindexRoot)
			if err != nil {
				return err
			}

			if isFinalReindexRoot {
				break
			}

			reindexRootAncestor = chosenChild
		}
	} else {
		log.Debugf("newReindexRoot is the same as reindexRootAncestor. Skipping concentration...")
	}

	// Update reindex root data store
	rt.stageReindexRoot(newReindexRoot)
	log.Debugf("Updated the reindex root to %s", newReindexRoot)
	return nil
}

// findNextReindexRoot finds the new reindex root based on the current one and the new selected tip.
// The function also returns the common ancestor between the current and new reindex roots (possibly current root itself).
// This ancestor should be used as a starting point for concentrating the interval towards the new root.
func (rt *reachabilityManager) findNextReindexRoot(currentReindexRoot, selectedTip *externalapi.DomainHash) (
	reindexRootAncestor, newReindexRoot *externalapi.DomainHash, err error) {

	reindexRootAncestor = currentReindexRoot
	newReindexRoot = currentReindexRoot

	selectedTipGHOSTDAGData, err := rt.ghostdagDataStore.Get(rt.databaseContext, nil, selectedTip)
	if err != nil {
		return nil, nil, err
	}

	isCurrentAncestorOfTip, err := rt.IsReachabilityTreeAncestorOf(currentReindexRoot, selectedTip)
	if err != nil {
		return nil, nil, err
	}

	// Test if current root is ancestor of selected tip - if not, this is a reorg case
	if !isCurrentAncestorOfTip {
		currentRootGHOSTDAGData, err := rt.ghostdagDataStore.Get(rt.databaseContext, nil, currentReindexRoot)
		if err != nil {
			return nil, nil, err
		}

		// We have reindex root out of selected tip chain, however we switch chains only after a sufficient
		// threshold of reindexSlack score in order to address possible alternating reorg attacks.
		// The reindexSlack constant is used as an heuristic for a large enough constant on the one hand, but
		// one which will not harm performance on the other hand - given the available slack at the chain split point.
		//
		// Note: In some cases the blue score selected tip can be lower than the current reindex root blue score.
		// If that's the case we keep the reindex root unchanged.
		if selectedTipGHOSTDAGData.BlueScore() < currentRootGHOSTDAGData.BlueScore() ||
			selectedTipGHOSTDAGData.BlueScore()-currentRootGHOSTDAGData.BlueScore() < rt.reindexSlack {
			// Return current - this indicates no change
			return currentReindexRoot, currentReindexRoot, nil
		}

		// The common ancestor is where we should start concentrating the interval from
		commonAncestor, err := rt.findCommonAncestor(selectedTip, currentReindexRoot)
		if err != nil {
			return nil, nil, err
		}

		reindexRootAncestor = commonAncestor
		newReindexRoot = commonAncestor
	}

	// Iterate from ancestor towards selected tip until passing the reindexWindow threshold,
	// for finding the new reindex root
	for {
		chosenChild, err := rt.FindNextAncestor(selectedTip, newReindexRoot)
		if err != nil {
			return nil, nil, err
		}

		chosenChildGHOSTDAGData, err := rt.ghostdagDataStore.Get(rt.databaseContext, nil, chosenChild)
		if err != nil {
			return nil, nil, err
		}

		if selectedTipGHOSTDAGData.BlueScore() < chosenChildGHOSTDAGData.BlueScore() {
			return nil, nil, errors.Errorf("chosen child %s has blue score greater "+
				"than %s although it's in its selected parent chain", chosenChild, selectedTip)
		}

		if selectedTipGHOSTDAGData.BlueScore()-chosenChildGHOSTDAGData.BlueScore() < rt.reindexWindow {
			break
		}

		newReindexRoot = chosenChild
	}

	return reindexRootAncestor, newReindexRoot, nil
}

/*

Test helper functions

*/

// Helper function (for testing purposes) to validate that all tree intervals
// under a specified subtree root are allocated correctly and as expected
func (rt *reachabilityManager) validateIntervals(root *externalapi.DomainHash) error {
	queue := []*externalapi.DomainHash{root}
	for len(queue) > 0 {
		var current *externalapi.DomainHash
		current, queue = queue[0], queue[1:]

		children, err := rt.children(current)
		if err != nil {
			return err
		}

		if len(children) > 0 {
			queue = append(queue, children...)
		}

		currentInterval, err := rt.interval(current)
		if err != nil {
			return err
		}

		if currentInterval.Start > currentInterval.End {
			err := errors.Errorf("Interval allocation is empty")
			return err
		}

		for i, child := range children {
			childInterval, err := rt.interval(child)
			if err != nil {
				return err
			}

			if i > 0 {
				siblingInterval, err := rt.interval(children[i-1])
				if err != nil {
					return err
				}

				if siblingInterval.End+1 != childInterval.Start {
					err := errors.Errorf("Child intervals are expected be right after each other")
					return err
				}
			}

			if childInterval.Start < currentInterval.Start {
				err := errors.Errorf("Child interval to the left of parent")
				return err
			}

			if childInterval.End >= currentInterval.End {
				err := errors.Errorf("Child interval to the right of parent")
				return err
			}
		}
	}

	return nil
}

// Helper function (for testing purposes) to get all nodes under a specified subtree root
func (rt *reachabilityManager) getAllNodes(root *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	queue := []*externalapi.DomainHash{root}
	nodes := []*externalapi.DomainHash{root}
	for len(queue) > 0 {
		var current *externalapi.DomainHash
		current, queue = queue[0], queue[1:]

		children, err := rt.children(current)
		if err != nil {
			return nil, err
		}

		if len(children) > 0 {
			queue = append(queue, children...)
			nodes = append(nodes, children...)
		}
	}

	return nodes, nil
}
