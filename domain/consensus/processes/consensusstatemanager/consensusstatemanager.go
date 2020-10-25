package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/dagconfig"
)

// consensusStateManager manages the node's consensus state
type consensusStateManager struct {
	dagParams *dagconfig.Params

	databaseContext     *database.DomainDBContext
	consensusStateStore model.ConsensusStateStore
	multisetStore       model.MultisetStore
	blockStore          model.BlockStore
	ghostdagManager     model.GHOSTDAGManager
	acceptanceManager   model.AcceptanceManager

	blockStatusStore  model.BlockStatusStore
	ghostdagDataStore model.GHOSTDAGDataStore
}

// New instantiates a new ConsensusStateManager
func New(
	databaseContext *database.DomainDBContext,
	dagParams *dagconfig.Params,
	consensusStateStore model.ConsensusStateStore,
	multisetStore model.MultisetStore,
	blockStore model.BlockStore,
	ghostdagManager model.GHOSTDAGManager,
	acceptanceManager model.AcceptanceManager,
	blockStatusStore model.BlockStatusStore,
	ghostdagDataStore model.GHOSTDAGDataStore) model.ConsensusStateManager {

	return &consensusStateManager{
		dagParams: dagParams,

		databaseContext:     databaseContext,
		consensusStateStore: consensusStateStore,
		multisetStore:       multisetStore,
		blockStore:          blockStore,
		ghostdagManager:     ghostdagManager,
		acceptanceManager:   acceptanceManager,

		blockStatusStore:  blockStatusStore,
		ghostdagDataStore: ghostdagDataStore,
	}
}

// AddBlockToVirtual submits the given block to be added to the
// current virtual. This process may result in a new virtual block
// getting created
func (csm *consensusStateManager) AddBlockToVirtual(blockHash *externalapi.DomainHash) error {
	return nil
}

// VirtualData returns the medianTime and blueScore of the current virtual block
func (csm *consensusStateManager) VirtualData() (medianTime int64, blueScore uint64, err error) {
	return 0, 0, nil
}

// RestoreUTXOSet calculates and returns the UTXOSet of the given blockHash
func (csm *consensusStateManager) RestorePastUTXOSet(blockHash *externalapi.DomainHash) (model.ReadOnlyUTXOSet, error) {
	return nil, nil
}

// RestoreDiffFromVirtual restores the diff between the given virtualDiffParentHash
// and the virtual
func (csm *consensusStateManager) RestoreDiffFromVirtual(utxoDiff *model.UTXODiff, virtualDiffParentHash *externalapi.DomainHash) (*model.UTXODiff, error) {
	panic("implement me")
}
