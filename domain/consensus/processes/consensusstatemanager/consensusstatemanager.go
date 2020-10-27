package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/dagconfig"
)

// consensusStateManager manages the node's consensus state
type consensusStateManager struct {
	dagParams       *dagconfig.Params
	databaseContext model.DBReader

	ghostdagManager       model.GHOSTDAGManager
	dagTopologyManager    model.DAGTopologyManager
	pruningManager        model.PruningManager
	pastMedianTimeManager model.PastMedianTimeManager

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
	pruningManager model.PruningManager,
	pastMedianTimeManager model.PastMedianTimeManager,
	blockStatusStore model.BlockStatusStore,
	ghostdagDataStore model.GHOSTDAGDataStore,
	consensusStateStore model.ConsensusStateStore,
	multisetStore model.MultisetStore,
	blockStore model.BlockStore,
	utxoDiffStore model.UTXODiffStore,
	blockRelationStore model.BlockRelationStore,
	acceptanceDataStore model.AcceptanceDataStore,
	blockHeaderStore model.BlockHeaderStore) model.ConsensusStateManager {

	return &consensusStateManager{
		dagParams:       dagParams,
		databaseContext: databaseContext,

		ghostdagManager:       ghostdagManager,
		dagTopologyManager:    dagTopologyManager,
		pruningManager:        pruningManager,
		pastMedianTimeManager: pastMedianTimeManager,

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
}

// PopulateTransactionWithUTXOEntries populates the transaction UTXO entries with data from the virtual.
func (csm *consensusStateManager) PopulateTransactionWithUTXOEntries(transaction *externalapi.DomainTransaction) error {
	for _, transactionInput := range transaction.Inputs {
		utxoEntry, err := csm.consensusStateStore.UTXOByOutpoint(csm.databaseContext, &transactionInput.PreviousOutpoint)
		if err != nil {
			return err
		}
		if utxoEntry == nil {
			return ruleerrors.ErrMissingTxOut
		}
		transactionInput.UTXOEntry = utxoEntry
	}

	return nil
}

// AddBlockToVirtual submits the given block to be added to the
// current virtual. This process may result in a new virtual block
// getting created
func (csm *consensusStateManager) AddBlockToVirtual(blockHash *externalapi.DomainHash) error {
	isNextVirtualSelectedParent, err := csm.isNextVirtualSelectedParent(blockHash)
	if err != nil {
		return err
	}

	if isNextVirtualSelectedParent {
		blockStatus, err := csm.resolveBlockStatus(blockHash)
		if err != nil {
			return err
		}
		if blockStatus == model.StatusValid {
			err = csm.checkFinalityViolation(blockHash)
			if err != nil {
				return err
			}
		}
	}

	newTips, err := csm.addTip(blockHash)
	if err != nil {
		return err
	}

	err = csm.updateVirtual(newTips)
	if err != nil {
		return err
	}

	return nil
}

func (csm *consensusStateManager) isNextVirtualSelectedParent(blockHash *externalapi.DomainHash) (bool, error) {

	virtualGhostdagData, err := csm.ghostdagDataStore.Get(csm.databaseContext, model.VirtualBlockHash)
	if err != nil {
		return false, err
	}

	nextVirtualSelectedParent, err := csm.ghostdagManager.ChooseSelectedParent(virtualGhostdagData.SelectedParent, blockHash)
	if err != nil {
		return false, err
	}
	return *blockHash == *nextVirtualSelectedParent, nil
}

func (csm *consensusStateManager) calculateAcceptanceDataAndMultiset(blockHash *externalapi.DomainHash) (
	*model.BlockAcceptanceData, model.Multiset, *model.UTXODiff, error) {

	blockGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, blockHash)
	if err != nil {
		return nil, nil, nil, err
	}

	selectedParentPastUTXO, err := csm.restorePastUTXO(blockGHOSTDAGData.SelectedParent)
	if err != nil {
		return nil, nil, nil, err
	}

	return csm.applyBlueBlocks(selectedParentPastUTXO, blockGHOSTDAGData)
}

func (csm *consensusStateManager) restorePastUTXO(blockHash *externalapi.DomainHash) (*model.UTXODiff, error) {
	// TODO
	return nil, nil
}

func (csm *consensusStateManager) applyBlueBlocks(
	selectedParentPastUTXO *model.UTXODiff, ghostdagData *model.BlockGHOSTDAGData) (
	*model.BlockAcceptanceData, model.Multiset, *model.UTXODiff, error) {

	// TODO
	return nil, nil, nil, nil
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

	return &model.VirtualData{
		PastMedianTime: pastMedianTime,
		BlueScore:      ghostdagData.BlueScore,
		ParentHashes:   nil, // TODO
		SelectedParent: ghostdagData.SelectedParent,
	}, nil
}

func (csm *consensusStateManager) resolveBlockStatus(blockHash *externalapi.DomainHash) (model.BlockStatus, error) {
	// TODO
	return 0, nil
}

func (csm *consensusStateManager) checkFinalityViolation(blockHash *externalapi.DomainHash) error {
	// TODO
	return nil
}

func (csm *consensusStateManager) addTip(hash *externalapi.DomainHash) (newTips []*externalapi.DomainHash, err error) {
	// TODO
	return nil, nil
}

func (csm *consensusStateManager) updateVirtual(tips []*externalapi.DomainHash) error {
	// TODO

	return nil
}
