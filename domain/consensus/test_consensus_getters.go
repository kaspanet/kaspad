package consensus

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/processes/blockbuilder"
)

func (tc *testConsensus) DatabaseContext() model.DBReader {
	return tc.databaseContext
}

func (tc *testConsensus) AcceptanceDataStore() model.AcceptanceDataStore {
	return tc.AcceptanceDataStore()
}

func (tc *testConsensus) BlockHeaderStore() model.BlockHeaderStore {
	return tc.blockHeaderStore
}

func (tc *testConsensus) BlockRelationStore() model.BlockRelationStore {
	return tc.blockRelationStore
}

func (tc *testConsensus) BlockStatusStore() model.BlockStatusStore {
	return tc.blockStatusStore
}

func (tc *testConsensus) BlockStore() model.BlockStore {
	return tc.blockStore
}

func (tc *testConsensus) ConsensusStateStore() model.ConsensusStateStore {
	return tc.consensusStateStore
}

func (tc *testConsensus) GHOSTDAGDataStore() model.GHOSTDAGDataStore {
	return tc.ghostdagDataStore
}

func (tc *testConsensus) HeaderTipsStore() model.HeaderTipsStore {
	return tc.headerTipsStore
}

func (tc *testConsensus) MultisetStore() model.MultisetStore {
	return tc.multisetStore
}

func (tc *testConsensus) PruningStore() model.PruningStore {
	return tc.pruningStore
}

func (tc *testConsensus) ReachabilityDataStore() model.ReachabilityDataStore {
	return tc.reachabilityDataStore
}

func (tc *testConsensus) UTXODiffStore() model.UTXODiffStore {
	return tc.utxoDiffStore
}

func (tc *testConsensus) BlockBuilder() model.TestBlockBuilder {
	return blockbuilder.NewTestBlockBuilder(tc.blockBuilder)
}

func (tc *testConsensus) BlockProcessor() model.BlockProcessor {
	return tc.blockProcessor
}

func (tc *testConsensus) BlockValidator() model.BlockValidator {
	return tc.blockValidator
}

func (tc *testConsensus) CoinbaseManager() model.CoinbaseManager {
	return tc.coinbaseManager
}

func (tc *testConsensus) ConsensusStateManager() model.TestConsensusStateManager {
	return tc.testConsensusStateManager
}

func (tc *testConsensus) DAGTopologyManager() model.DAGTopologyManager {
	return tc.dagTopologyManager
}

func (tc *testConsensus) DAGTraversalManager() model.DAGTraversalManager {
	return tc.dagTraversalManager
}

func (tc *testConsensus) DifficultyManager() model.DifficultyManager {
	return tc.difficultyManager
}

func (tc *testConsensus) GHOSTDAGManager() model.GHOSTDAGManager {
	return tc.ghostdagManager
}

func (tc *testConsensus) HeaderTipsManager() model.HeaderTipsManager {
	return tc.headerTipsManager
}

func (tc *testConsensus) MergeDepthManager() model.MergeDepthManager {
	return tc.mergeDepthManager
}

func (tc *testConsensus) PastMedianTimeManager() model.PastMedianTimeManager {
	return tc.pastMedianTimeManager
}

func (tc *testConsensus) PruningManager() model.PruningManager {
	return tc.pruningManager
}

func (tc *testConsensus) ReachabilityManager() model.ReachabilityManager {
	return tc.reachabilityManager
}

func (tc *testConsensus) SyncManager() model.SyncManager {
	return tc.syncManager
}

func (tc *testConsensus) TransactionValidator() model.TransactionValidator {
	return tc.transactionValidator
}
