package ghostdagmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/datastructures"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/processes"
	"github.com/kaspanet/kaspad/util/daghash"
)

// GHOSTDAGManager ...
type GHOSTDAGManager struct {
	dagTopologyManager processes.DAGTopologyManager
	ghostdagDataStore  datastructures.GHOSTDAGDataStore
}

// New ...
func New(
	dagTopologyManager processes.DAGTopologyManager,
	ghostdagDataStore datastructures.GHOSTDAGDataStore) *GHOSTDAGManager {
	return &GHOSTDAGManager{
		dagTopologyManager: dagTopologyManager,
		ghostdagDataStore:  ghostdagDataStore,
	}
}

// GHOSTDAG ...
func (gm *GHOSTDAGManager) GHOSTDAG(blockParents []*daghash.Hash) *model.BlockGHOSTDAGData {
	return nil
}

// BlockData ...
func (gm *GHOSTDAGManager) BlockData(blockHash *daghash.Hash) *model.BlockGHOSTDAGData {
	return nil
}
