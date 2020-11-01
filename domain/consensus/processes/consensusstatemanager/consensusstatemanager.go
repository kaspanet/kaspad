package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/dagconfig"
)

// consensusStateManager manages the node's consensus state
type consensusStateManager struct {
	dagParams       *dagconfig.Params
	databaseContext model.DBReader

	ghostdagManager       model.GHOSTDAGManager
	dagTopologyManager    model.DAGTopologyManager
	dagTraversalManager   model.DAGTraversalManager
	pruningManager        model.PruningManager
	pastMedianTimeManager model.PastMedianTimeManager
	transactionValidator  model.TransactionValidator
	blockValidator        model.BlockValidator

	blockStatusStore    model.BlockStatusStore
	ghostdagDataStore   model.GHOSTDAGDataStore
	consensusStateStore model.ConsensusStateStore
	multisetStore       model.MultisetStore
	blockStore          model.BlockStore
	utxoDiffStore       model.UTXODiffStore
	blockRelationStore  model.BlockRelationStore
	acceptanceDataStore model.AcceptanceDataStore
	blockHeaderStore    model.BlockHeaderStore
}

// New instantiates a new ConsensusStateManager
func New(
	databaseContext model.DBReader,
	dagParams *dagconfig.Params,
	ghostdagManager model.GHOSTDAGManager,
	dagTopologyManager model.DAGTopologyManager,
	dagTraversalManager model.DAGTraversalManager,
	pruningManager model.PruningManager,
	pastMedianTimeManager model.PastMedianTimeManager,
	transactionValidator model.TransactionValidator,
	blockValidator model.BlockValidator,
	blockStatusStore model.BlockStatusStore,
	ghostdagDataStore model.GHOSTDAGDataStore,
	consensusStateStore model.ConsensusStateStore,
	multisetStore model.MultisetStore,
	blockStore model.BlockStore,
	utxoDiffStore model.UTXODiffStore,
	blockRelationStore model.BlockRelationStore,
	acceptanceDataStore model.AcceptanceDataStore,
	blockHeaderStore model.BlockHeaderStore) (model.ConsensusStateManager, error) {

	csm := &consensusStateManager{
		dagParams:       dagParams,
		databaseContext: databaseContext,

		ghostdagManager:       ghostdagManager,
		dagTopologyManager:    dagTopologyManager,
		dagTraversalManager:   dagTraversalManager,
		pruningManager:        pruningManager,
		pastMedianTimeManager: pastMedianTimeManager,
		transactionValidator:  transactionValidator,
		blockValidator:        blockValidator,

		multisetStore:       multisetStore,
		blockStore:          blockStore,
		blockStatusStore:    blockStatusStore,
		ghostdagDataStore:   ghostdagDataStore,
		consensusStateStore: consensusStateStore,
		utxoDiffStore:       utxoDiffStore,
		blockRelationStore:  blockRelationStore,
		acceptanceDataStore: acceptanceDataStore,
		blockHeaderStore:    blockHeaderStore,
	}

	return csm, nil
}

// VirtualData returns data on the current virtual block
func (csm *consensusStateManager) VirtualData() (virtualData *model.VirtualData, err error) {
	pastMedianTime, err := csm.pastMedianTimeManager.PastMedianTime(model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	ghostdagData, err := csm.ghostdagDataStore.Get(csm.databaseContext, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	virtualParents, err := csm.dagTopologyManager.Parents(model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	return &model.VirtualData{
		PastMedianTime: pastMedianTime,
		BlueScore:      ghostdagData.BlueScore,
		ParentHashes:   virtualParents,
		SelectedParent: ghostdagData.SelectedParent,
	}, nil
}
