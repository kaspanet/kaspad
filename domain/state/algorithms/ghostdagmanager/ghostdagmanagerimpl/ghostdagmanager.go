package ghostdagmanagerimpl

import (
	"github.com/kaspanet/kaspad/domain/state/algorithms/dagtopologymanager"
	"github.com/kaspanet/kaspad/domain/state/datastructures/ghostdagdatastore"
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
