package ghostdagmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
)

// GHOSTDAGManager resolves and manages GHOSTDAG block data
type GHOSTDAGManager struct {
	databaseContext    *database.DomainDBContext
	dagTopologyManager model.DAGTopologyManager
	ghostdagDataStore  model.GHOSTDAGDataStore
	k                  model.KType
}

// New instantiates a new GHOSTDAGManager
func New(
	databaseContext *dbaccess.DatabaseContext,
	dagTopologyManager model.DAGTopologyManager,
	ghostdagDataStore model.GHOSTDAGDataStore,
	k model.KType) *GHOSTDAGManager {

	return &GHOSTDAGManager{
		databaseContext:    database.NewDomainDBContext(databaseContext),
		dagTopologyManager: dagTopologyManager,
		ghostdagDataStore:  ghostdagDataStore,
		k:                  k,
	}
}

// BlockData returns previously calculated GHOSTDAG data for
// the given blockHash
func (gm *GHOSTDAGManager) BlockData(blockHash *model.DomainHash) *model.BlockGHOSTDAGData {
	return gm.ghostdagDataStore.Get(gm.databaseContext, blockHash)
}
