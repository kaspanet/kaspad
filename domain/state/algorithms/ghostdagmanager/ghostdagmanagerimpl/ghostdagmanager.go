package ghostdagmanagerimpl

import (
	"github.com/kaspanet/kaspad/domain/state/algorithms/dagtopologymanager"
	"github.com/kaspanet/kaspad/domain/state/datastructures/ghostdagdatastore"
	"github.com/kaspanet/kaspad/util/daghash"
)

type GHOSTDAGManager struct {
	dagTopologyManager dagtopologymanager.DAGTopologyManager
	ghostdagDataStore  ghostdagdatastore.GHOSTDAGDataStore
}

func New(
	dagTopologyManager dagtopologymanager.DAGTopologyManager,
	ghostdagDataStore ghostdagdatastore.GHOSTDAGDataStore) *GHOSTDAGManager {
	return &GHOSTDAGManager{
		dagTopologyManager: dagTopologyManager,
		ghostdagDataStore:  ghostdagDataStore,
	}
}

func (gm *GHOSTDAGManager) GHOSTDAG(blockHash *daghash.Hash) {

}

func (gm *GHOSTDAGManager) BlockData(blockHash *daghash.Hash) *ghostdagdatastore.BlockGHOSTDAGData {
	return nil
}
