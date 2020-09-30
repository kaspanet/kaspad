package ghostdagmanagerimpl

import (
	"github.com/kaspanet/kaspad/domain/kaspadstate/algorithms/dagtopologymanager"
	"github.com/kaspanet/kaspad/domain/kaspadstate/datastructures"
	"github.com/kaspanet/kaspad/domain/kaspadstate/model"
	"github.com/kaspanet/kaspad/util/daghash"
)

type GHOSTDAGManager struct {
	dagTopologyManager dagtopologymanager.DAGTopologyManager
	ghostdagDataStore  datastructures.GHOSTDAGDataStore
}

func New(
	dagTopologyManager dagtopologymanager.DAGTopologyManager,
	ghostdagDataStore datastructures.GHOSTDAGDataStore) *GHOSTDAGManager {
	return &GHOSTDAGManager{
		dagTopologyManager: dagTopologyManager,
		ghostdagDataStore:  ghostdagDataStore,
	}
}

func (gm *GHOSTDAGManager) GHOSTDAG(blockHash *daghash.Hash) {

}

func (gm *GHOSTDAGManager) BlockData(blockHash *daghash.Hash) *model.BlockGHOSTDAGData {
	return nil
}
