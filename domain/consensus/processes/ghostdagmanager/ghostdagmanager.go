package ghostdagmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
)

// ghostdagManager resolves and manages GHOSTDAG block data
type ghostdagManager struct {
	databaseContext    *database.DomainDBContext
	dagTopologyManager model.DAGTopologyManager
	ghostdagDataStore  model.GHOSTDAGDataStore
	k                  model.KType
}

// New instantiates a new ghostdagManager
func New(
	databaseContext *dbaccess.DatabaseContext,
	dagTopologyManager model.DAGTopologyManager,
	ghostdagDataStore model.GHOSTDAGDataStore,
	k model.KType) model.GHOSTDAGManager {

	return &ghostdagManager{
		databaseContext:    database.NewDomainDBContext(databaseContext),
		dagTopologyManager: dagTopologyManager,
		ghostdagDataStore:  ghostdagDataStore,
		k:                  k,
	}
}

// BlockData returns previously calculated GHOSTDAG data for
// the given blockHash
func (gm *ghostdagManager) BlockData(blockHash *model.DomainHash) (*model.BlockGHOSTDAGData, error) {
	return gm.ghostdagDataStore.Get(gm.databaseContext, blockHash)
}
