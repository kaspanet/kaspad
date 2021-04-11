package reachabilitymanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// IsDAGAncestorOf returns true if blockHashA is an ancestor of
// blockHashB in the DAG.
//
// Note: this method will return true if blockHashA == blockHashB
// The complexity of this method is O(log(|this.futureCoveringTreeNodeSet|))
func (rt *reachabilityManager) IsDAGAncestorOf(stagingArea *model.StagingArea, blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	// Check if this node is a reachability tree ancestor of the
	// other node
	isReachabilityTreeAncestor, err := rt.IsReachabilityTreeAncestorOf(stagingArea, blockHashA, blockHashB)
	if err != nil {
		return false, err
	}
	if isReachabilityTreeAncestor {
		return true, nil
	}

	// Otherwise, use previously registered future blocks to complete the
	// reachability test
	return rt.futureCoveringSetHasAncestorOf(stagingArea, blockHashA, blockHashB)
}

func (rt *reachabilityManager) UpdateReindexRoot(stagingArea *model.StagingArea, selectedTip *externalapi.DomainHash) error {
	return rt.updateReindexRoot(stagingArea, selectedTip)
}
