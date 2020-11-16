package consensus

import (
	"math/rand"

	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"github.com/kaspanet/kaspad/domain/consensus/utils/mining"

	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type testConsensus struct {
	*consensus
	rd *rand.Rand

	testBlockBuilder          model.TestBlockBuilder
	testReachabilityManager   model.TestReachabilityManager
	testConsensusStateManager model.TestConsensusStateManager
}

func (tc *testConsensus) BuildBlockWithParents(parentHashes []*externalapi.DomainHash, coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error) {

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

	scriptPublicKey, _ := testutils.OpTrueScript()

	if coinbaseData == nil {
		coinbaseData = &externalapi.DomainCoinbaseData{
			ScriptPublicKey: scriptPublicKey,
			ExtraData:       []byte{},
		}
	}

	block, err := tc.testBlockBuilder.BuildBlockWithParents(parentHashes, coinbaseData, transactions)
	if err != nil {
		return nil, err
	}

	return tc.SolveAndAddBlock(block)
}

func (tc *testConsensus) SolveAndAddBlock(block *externalapi.DomainBlock) (*externalapi.DomainHash, error) {
	mining.SolveBlock(block, tc.rd)

	err := tc.blockProcessor.ValidateAndInsertBlock(block)
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
