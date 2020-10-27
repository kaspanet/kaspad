package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/dagconfig"
)

// consensusStateManager manages the node's consensus state
type consensusStateManager struct {
	dagParams       *dagconfig.Params
	databaseContext *database.DomainDBContext

	ghostdagManager    model.GHOSTDAGManager
	dagTopologyManager model.DAGTopologyManager
	pruningManager     model.PruningManager

	blockStatusStore    model.BlockStatusStore
	ghostdagDataStore   model.GHOSTDAGDataStore
	consensusStateStore model.ConsensusStateStore
	multisetStore       model.MultisetStore
	blockStore          model.BlockStore
	utxoDiffStore       model.UTXODiffStore
	blockRelationStore  model.BlockRelationStore
	acceptanceDataStore model.AcceptanceDataStore
}

// New instantiates a new ConsensusStateManager
func New(
	databaseContext *database.DomainDBContext,
	dagParams *dagconfig.Params,
	ghostdagManager model.GHOSTDAGManager,
	dagTopologyManager model.DAGTopologyManager,
	pruningManager model.PruningManager,
	blockStatusStore model.BlockStatusStore,
	ghostdagDataStore model.GHOSTDAGDataStore,
	consensusStateStore model.ConsensusStateStore,
	multisetStore model.MultisetStore,
	blockStore model.BlockStore,
	utxoDiffStore model.UTXODiffStore,
	blockRelationStore model.BlockRelationStore,
	acceptanceDataStore model.AcceptanceDataStore) model.ConsensusStateManager {

	return &consensusStateManager{
		dagParams:       dagParams,
		databaseContext: databaseContext,

		ghostdagManager:    ghostdagManager,
		dagTopologyManager: dagTopologyManager,
		pruningManager:     pruningManager,

		multisetStore:       multisetStore,
		blockStore:          blockStore,
		blockStatusStore:    blockStatusStore,
		ghostdagDataStore:   ghostdagDataStore,
		consensusStateStore: consensusStateStore,
		utxoDiffStore:       utxoDiffStore,
		blockRelationStore:  blockRelationStore,
		acceptanceDataStore: acceptanceDataStore,
	}
}

// AddBlockToVirtual submits the given block to be added to the
// current virtual. This process may result in a new virtual block
// getting created
func (csm *consensusStateManager) AddBlockToVirtual(blockHash *externalapi.DomainHash) error {
	return nil
}

// VirtualData returns data on the current virtual block
func (csm *consensusStateManager) VirtualData() (virtualData *model.VirtualData, err error) {
	panic("implement me")
}

// PopulateTransactionWithUTXOEntries populates the transaction UTXO entries with data from the virtual.
func (csm *consensusStateManager) PopulateTransactionWithUTXOEntries(transaction *externalapi.DomainTransaction) error {
	panic("implement me")
}
