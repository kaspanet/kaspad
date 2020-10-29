package consensusstatemanager

import (
	"errors"

	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashserialization"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/processes/consensusstatemanager/utxoalgebra"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
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
	reachabilityTree      model.ReachabilityTree
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
	reachabilityTree model.ReachabilityTree,
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
		reachabilityTree:      reachabilityTree,
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

	err = csm.updateVirtual(blockHash, newTips)
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
	[]*model.BlockAcceptanceData, model.Multiset, *model.UTXODiff, error) {

	blockGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, blockHash)
	if err != nil {
		return nil, nil, nil, err
	}

	selectedParentPastUTXO, err := csm.restorePastUTXO(blockGHOSTDAGData.SelectedParent)
	if err != nil {
		return nil, nil, nil, err
	}

	acceptanceData, utxoDiff, err := csm.applyBlueBlocks(blockHash, selectedParentPastUTXO, blockGHOSTDAGData)
	if err != nil {
		return nil, nil, nil, err
	}

	multiset, err := csm.calculateMultiset(acceptanceData, blockGHOSTDAGData)
	if err != nil {
		return nil, nil, nil, err
	}

	return acceptanceData, multiset, utxoDiff, nil
}

func (csm *consensusStateManager) restorePastUTXO(blockHash *externalapi.DomainHash) (*model.UTXODiff, error) {
	var err error

	// collect the UTXO diffs
	var utxoDiffs []*model.UTXODiff
	nextBlockHash := blockHash
	for nextBlockHash != nil {
		utxoDiff, err := csm.utxoDiffStore.UTXODiff(csm.databaseContext, nextBlockHash)
		if err != nil {
			return nil, err
		}
		utxoDiffs = append(utxoDiffs, utxoDiff)

		nextBlockHash, err = csm.utxoDiffStore.UTXODiffChild(csm.databaseContext, nextBlockHash)
	}

	// apply the diffs in reverse order
	accumulatedDiff := model.NewUTXODiff()
	for i := len(utxoDiffs) - 1; i >= 0; i-- {
		accumulatedDiff, err = utxoalgebra.WithDiff(accumulatedDiff, utxoDiffs[i])
		if err != nil {
			return nil, err
		}
	}

	return accumulatedDiff, nil
}

func (csm *consensusStateManager) applyBlueBlocks(blockHash *externalapi.DomainHash,
	selectedParentPastUTXODiff *model.UTXODiff, ghostdagData *model.BlockGHOSTDAGData) (
	[]*model.BlockAcceptanceData, *model.UTXODiff, error) {

	blueBlocks, err := csm.blockStore.Blocks(csm.databaseContext, ghostdagData.MergeSetBlues)
	if err != nil {
		return nil, nil, err
	}

	selectedParentMedianTime, err := csm.pastMedianTimeManager.PastMedianTime(ghostdagData.SelectedParent)
	if err != nil {
		return nil, nil, err
	}

	multiblockAcceptanceData := make([]*model.BlockAcceptanceData, len(blueBlocks))
	accumulatedUTXODiff := utxoalgebra.DiffClone(selectedParentPastUTXODiff)
	accumulatedMass := uint64(0)

	for i, blueBlock := range blueBlocks {
		blockAccepanceData := &model.BlockAcceptanceData{
			TransactionAcceptanceData: []*model.TransactionAcceptanceData{},
		}
		isSelectedParent := i == 0

		for j, transaction := range blueBlock.Transactions {
			var isAccepted bool
			var fee uint64

			isAccepted, accumulatedMass, err = csm.maybeAcceptTransaction(transaction, blockHash, isSelectedParent,
				accumulatedUTXODiff, accumulatedMass, selectedParentMedianTime, ghostdagData.BlueScore)
			if err != nil {
				return nil, nil, err
			}

			blockAccepanceData.TransactionAcceptanceData[j] = &model.TransactionAcceptanceData{
				Transaction: transaction,
				Fee:         fee,
				IsAccepted:  isAccepted,
			}
		}
		multiblockAcceptanceData[i] = blockAccepanceData
	}

	return multiblockAcceptanceData, accumulatedUTXODiff, nil
}

func (csm *consensusStateManager) maybeAcceptTransaction(transaction *externalapi.DomainTransaction,
	blockHash *externalapi.DomainHash, isSelectedParent bool, accumulatedUTXODiff *model.UTXODiff,
	accumulatedMassBefore uint64, selectedParentPastMedianTime int64, blockBlueScore uint64) (
	isAccepted bool, accumulatedMassAfter uint64, err error) {

	err = csm.populateTransactionWithUTXOEntriesFromVirtualOrDiff(transaction, accumulatedUTXODiff)
	if err != nil {
		return false, accumulatedMassBefore, err
	}

	// Coinbase transaction outputs are added to the UTXO-set only if they are in the selected parent chain.
	if transactionhelper.IsCoinBase(transaction) {
		if !isSelectedParent {
			return false, accumulatedMassBefore, nil
		}

		err := utxoalgebra.DiffAddTransaction(accumulatedUTXODiff, transaction, blockBlueScore)
		if err != nil {
			return false, accumulatedMassBefore, err
		}

		return true, accumulatedMassBefore, nil
	}

	err = csm.transactionValidator.ValidateTransactionInContextAndPopulateMassAndFee(
		transaction, blockHash, selectedParentPastMedianTime)
	if err != nil {
		if !errors.As(err, &(ruleerrors.RuleError{})) {
			return false, accumulatedMassBefore, err
		}

		return false, accumulatedMassBefore, nil
	}

	isAccepted = true
	isAccepted, accumulatedMassAfter = csm.checkTransactionMass(transaction, accumulatedMassBefore)

	return isAccepted, accumulatedMassAfter, nil
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

func (csm *consensusStateManager) resolveBlockStatus(blockHash *externalapi.DomainHash) (model.BlockStatus, error) {
	// get list of all blocks in the selected parent chain that have not yet resolved their status
	unverifiedBlocks, selectedParentStatus, err := csm.getUnverifiedChainBlocksAndSelectedParentStatus(blockHash)
	if err != nil {
		return 0, err
	}

	// resolve the unverified blocks' statuses in opposite order
	for i := len(unverifiedBlocks); i >= 0; i++ {
		unverifiedBlockHash := unverifiedBlocks[i]

		var blockStatus model.BlockStatus
		if selectedParentStatus == model.StatusDisqualifiedFromChain {
			blockStatus = model.StatusDisqualifiedFromChain
		} else {
			blockStatus, err = csm.resolveSingleBlockStatus(unverifiedBlockHash)
			if err != nil {
				return 0, err
			}
		}

		csm.blockStatusStore.Stage(unverifiedBlockHash, blockStatus)
		selectedParentStatus = blockStatus
	}

	return 0, nil
}

func (csm *consensusStateManager) getUnverifiedChainBlocksAndSelectedParentStatus(blockHash *externalapi.DomainHash) (
	[]*externalapi.DomainHash, model.BlockStatus, error) {

	unverifiedBlocks := []*externalapi.DomainHash{blockHash}
	currentHash := blockHash
	for {
		ghostdagData, err := csm.ghostdagDataStore.Get(csm.databaseContext, currentHash)
		if err != nil {
			return nil, 0, err
		}

		selectedParentStatus, err := csm.blockStatusStore.Get(csm.databaseContext, ghostdagData.SelectedParent)
		if err != nil {
			return nil, 0, err
		}

		if selectedParentStatus != model.StatusUTXOPendingVerification {
			return unverifiedBlocks, selectedParentStatus, nil
		}

		unverifiedBlocks = append(unverifiedBlocks, ghostdagData.SelectedParent)

		currentHash = ghostdagData.SelectedParent
	}
}

func (csm *consensusStateManager) addTip(newTipHash *externalapi.DomainHash) (newTips []*externalapi.DomainHash, err error) {
	currentTips, err := csm.consensusStateStore.Tips(csm.databaseContext)
	if err != nil {
		return nil, err
	}

	newTipParents, err := csm.dagTopologyManager.Parents(newTipHash)
	if err != nil {
		return nil, err
	}

	newTips = []*externalapi.DomainHash{newTipHash}

	for _, currentTip := range currentTips {
		isCurrentTipInNewTipParents := false
		for _, newTipParent := range newTipParents {
			if *currentTip == *newTipParent {
				isCurrentTipInNewTipParents = true
				break
			}
		}
		if !isCurrentTipInNewTipParents {
			newTips = append(newTips, currentTip)
		}
	}

	err = csm.consensusStateStore.StageTips(newTips)
	if err != nil {
		return nil, err
	}

	return newTips, nil
}

func (csm *consensusStateManager) calculateMultiset(
	acceptanceData []*model.BlockAcceptanceData, blockGHOSTDAGData *model.BlockGHOSTDAGData) (model.Multiset, error) {

	selectedParentMultiset, err := csm.multisetStore.Get(csm.databaseContext, blockGHOSTDAGData.SelectedParent)
	if err != nil {
		return nil, err
	}

	multiset := selectedParentMultiset.Clone()

	for _, blockAcceptanceData := range acceptanceData {
		for _, transactionAcceptanceData := range blockAcceptanceData.TransactionAcceptanceData {
			if !transactionAcceptanceData.IsAccepted {
				continue
			}

			transaction := transactionAcceptanceData.Transaction

			var err error
			err = addTransactionToMultiset(multiset, transaction, blockGHOSTDAGData.BlueScore)
			if err != nil {
				return nil, err
			}
		}
	}

	return multiset, nil
}

func addTransactionToMultiset(multiset model.Multiset, transaction *externalapi.DomainTransaction,
	blockBlueScore uint64) error {

	for _, input := range transaction.Inputs {
		err := removeUTXOFromMultiset(multiset, input.UTXOEntry, &input.PreviousOutpoint)
		if err != nil {
			return err
		}
	}

	for i, output := range transaction.Outputs {
		outpoint := &externalapi.DomainOutpoint{
			ID:    *hashserialization.TransactionID(transaction),
			Index: uint32(i),
		}
		utxoEntry := &externalapi.UTXOEntry{
			Amount:          output.Value,
			ScriptPublicKey: output.ScriptPublicKey,
			BlockBlueScore:  blockBlueScore,
			IsCoinbase:      false,
		}
		err := addUTXOToMultiset(multiset, utxoEntry, outpoint)
		if err != nil {
			return err
		}
	}

	return nil
}

func addUTXOToMultiset(multiset model.Multiset, entry *externalapi.UTXOEntry,
	outpoint *externalapi.DomainOutpoint) error {

	serializedUTXO, err := hashserialization.SerializeUTXO(entry, outpoint)
	if err != nil {
		return err
	}
	multiset.Add(serializedUTXO)

	return nil
}

func removeUTXOFromMultiset(multiset model.Multiset, entry *externalapi.UTXOEntry,
	outpoint *externalapi.DomainOutpoint) error {

	serializedUTXO, err := hashserialization.SerializeUTXO(entry, outpoint)
	if err != nil {
		return err
	}
	multiset.Remove(serializedUTXO)

	return nil
}

func (csm *consensusStateManager) checkTransactionMass(
	transaction *externalapi.DomainTransaction, accumulatedMassBefore uint64) (
	isAccepted bool, accumulatedMassAfter uint64) {

	accumulatedMassAfter = accumulatedMassBefore + transaction.Mass

	// We could potentially overflow the accumulator so check for
	// overflow as well.
	if accumulatedMassAfter < transaction.Mass || accumulatedMassAfter > constants.MaxMassAcceptedByBlock {
		return false, 0
	}

	return true, accumulatedMassAfter
}

func (csm *consensusStateManager) resolveSingleBlockStatus(blockHash *externalapi.DomainHash) (model.BlockStatus, error) {
	acceptanceData, multiset, pastUTXODiff, err := csm.calculateAcceptanceDataAndMultiset(blockHash)
	if err != nil {
		return 0, err
	}

	block, err := csm.blockStore.Block(csm.databaseContext, blockHash)
	if err != nil {
		return 0, err
	}

	err = csm.verifyAndBuildUTXO(block, blockHash, pastUTXODiff, acceptanceData, multiset)
	if err != nil {
		if errors.As(err, (&ruleerrors.RuleError{})) {
			return model.StatusDisqualifiedFromChain, nil
		}
		return 0, err
	}

	csm.multisetStore.Stage(blockHash, multiset)
	csm.acceptanceDataStore.Stage(blockHash, acceptanceData)
	csm.utxoDiffStore.Stage(blockHash, pastUTXODiff, nil)

	err = csm.updateParentDiffs(blockHash, pastUTXODiff)
	if err != nil {
		return 0, err
	}

	return model.StatusValid, nil
}

func (csm *consensusStateManager) updateParentDiffs(
	blockHash *externalapi.DomainHash, pastUTXODiff *model.UTXODiff) error {
	parentHashes, err := csm.dagTopologyManager.Parents(blockHash)
	if err != nil {
		return err
	}
	for _, parentHash := range parentHashes {
		parentDiffChildHash, err := csm.utxoDiffStore.UTXODiffChild(csm.databaseContext, parentHash)
		if err != nil {
			return err
		}
		if parentDiffChildHash == nil {
			parentCurrentDiff, err := csm.utxoDiffStore.UTXODiff(csm.databaseContext, parentHash)
			if err != nil {
				return err
			}
			parentNewDiff, err := utxoalgebra.DiffFrom(pastUTXODiff, parentCurrentDiff)
			if err != nil {
				return err
			}

			csm.utxoDiffStore.Stage(parentHash, parentNewDiff, blockHash)
		}
	}

	return nil
}
