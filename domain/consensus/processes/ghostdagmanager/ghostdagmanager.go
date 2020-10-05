package ghostdagmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/datastructures"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/processes"
	"github.com/kaspanet/kaspad/util/daghash"
)

// GHOSTDAGManager resolves and manages GHOSTDAG block data
type GHOSTDAGManager struct {
	dagTopologyManager processes.DAGTopologyManager
	ghostdagDataStore  datastructures.GHOSTDAGDataStore
}

// New instantiates a new GHOSTDAGManager
func New(
	dagTopologyManager processes.DAGTopologyManager,
	ghostdagDataStore datastructures.GHOSTDAGDataStore) *GHOSTDAGManager {
	return &GHOSTDAGManager{
		dagTopologyManager: dagTopologyManager,
		ghostdagDataStore:  ghostdagDataStore,
	}
}

// GHOSTDAG calculates GHOSTDAG data for the block represented
// by the given blockParents
func (gm *GHOSTDAGManager) GHOSTDAG(blockParents []*daghash.Hash) *model.BlockGHOSTDAGData {
	return nil
}

// BlockData returns previously calculated GHOSTDAG data for
// the given blockHash
func (gm *GHOSTDAGManager) BlockData(blockHash *daghash.Hash) *model.BlockGHOSTDAGData {
	return nil
}
