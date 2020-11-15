package consensus

import (
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"github.com/kaspanet/kaspad/domain/consensus/utils/mining"
	"math/rand"

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

	if coinbaseData == nil {
		coinbaseData = &externalapi.DomainCoinbaseData{
			ScriptPublicKey: testutils.OpTrueScript(),
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
