package consensus

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
)

type testConsensus struct {
	*consensus

	testBlockBuilder          testapi.TestBlockBuilder
	testReachabilityManager   testapi.TestReachabilityManager
	testConsensusStateManager testapi.TestConsensusStateManager
	testTransactionValidator  testapi.TestTransactionValidator
}

func (tc *testConsensus) BuildBlockWithParents(parentHashes []*externalapi.DomainHash,
	coinbaseData *externalapi.DomainCoinbaseData, transactions []*externalapi.DomainTransaction) (
	*externalapi.DomainBlock, model.UTXODiff, error) {

	// Require write lock because BuildBlockWithParents stages temporary data
	tc.lock.Lock()
	defer tc.lock.Unlock()

	return tc.testBlockBuilder.BuildBlockWithParents(parentHashes, coinbaseData, transactions)
}

func (tc *testConsensus) AddBlock(parentHashes []*externalapi.DomainHash, coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (*externalapi.DomainHash, error) {

	// Require write lock because BuildBlockWithParents stages temporary data
	tc.lock.Lock()
	defer tc.lock.Unlock()

	block, _, err := tc.testBlockBuilder.BuildBlockWithParents(parentHashes, coinbaseData, transactions)
	if err != nil {
		return nil, err
	}

	err = tc.blockProcessor.ValidateAndInsertBlock(block)
	if err != nil {
		return nil, err
	}

	return consensusserialization.BlockHash(block), nil
}

func (tc *testConsensus) DiscardAllStores() {
	tc.AcceptanceDataStore().Discard()
	tc.BlockHeaderStore().Discard()
	tc.BlockRelationStore().Discard()
	tc.BlockStatusStore().Discard()
	tc.BlockStore().Discard()
	tc.ConsensusStateStore().Discard()
	tc.GHOSTDAGDataStore().Discard()
	tc.HeaderTipsStore().Discard()
	tc.MultisetStore().Discard()
	tc.PruningStore().Discard()
	tc.ReachabilityDataStore().Discard()
	tc.UTXODiffStore().Discard()
}
