package consensus

import (
	"encoding/json"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/pkg/errors"
	"io"
)

type testConsensus struct {
	*consensus
	dagParams  *dagconfig.Params
	testParams *testapi.TestParams
	database   database.Database

	testBlockBuilder          testapi.TestBlockBuilder
	testReachabilityManager   testapi.TestReachabilityManager
	testConsensusStateManager testapi.TestConsensusStateManager
	testTransactionValidator  testapi.TestTransactionValidator
}

func (tc *testConsensus) DAGParams() *dagconfig.Params {
	return tc.dagParams
}

func (tc *testConsensus) TestParams() *testapi.TestParams {
	return tc.testParams
}

func (tc *testConsensus) BuildBlockWithParents(parentHashes []*externalapi.DomainHash,
	coinbaseData *externalapi.DomainCoinbaseData, transactions []*externalapi.DomainTransaction) (
	*externalapi.DomainBlock, model.UTXODiff, error) {

	// Require write lock because BuildBlockWithParents stages temporary data
	tc.lock.Lock()
	defer tc.lock.Unlock()

	block, diff, err := tc.testBlockBuilder.BuildBlockWithParents(parentHashes, coinbaseData, transactions)
	if err != nil {
		return nil, nil, err
	}

	return block, diff, nil
}

func (tc *testConsensus) AddBlock(parentHashes []*externalapi.DomainHash, coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (*externalapi.DomainHash, *externalapi.BlockInsertionResult, error) {

	// Require write lock because BuildBlockWithParents stages temporary data
	tc.lock.Lock()
	defer tc.lock.Unlock()

	block, _, err := tc.testBlockBuilder.BuildBlockWithParents(parentHashes, coinbaseData, transactions)
	if err != nil {
		return nil, nil, err
	}

	blockInsertionResult, err := tc.blockProcessor.ValidateAndInsertBlock(block)
	if err != nil {
		return nil, nil, err
	}

	return consensushashing.BlockHash(block), blockInsertionResult, nil
}

func (tc *testConsensus) AddHeader(parentHashes []*externalapi.DomainHash, coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (*externalapi.DomainHash, *externalapi.BlockInsertionResult, error) {

	// Require write lock because BuildBlockWithParents stages temporary data
	tc.lock.Lock()
	defer tc.lock.Unlock()

	block, _, err := tc.testBlockBuilder.BuildBlockWithParents(parentHashes, coinbaseData, transactions)
	if err != nil {
		return nil, nil, err
	}

	block.Transactions = nil

	blockInsertionResult, err := tc.blockProcessor.ValidateAndInsertBlock(block)
	if err != nil {
		return nil, nil, err
	}

	return consensushashing.BlockHash(block), blockInsertionResult, nil
}

func (tc *testConsensus) MineJSON(r io.Reader) error {
	// JSONBlock is a json representation of a block in mine format
	type JSONBlock struct {
		ID      string   `json:"id"`
		Parents []string `json:"parents"`
	}

	parentsMap := make(map[string]*externalapi.DomainHash)
	parentsMap["0"] = tc.dagParams.GenesisHash

	decoder := json.NewDecoder(r)
	// read open bracket
	_, err := decoder.Token()
	if err != nil {
		return err
	}
	// while the array contains values
	for decoder.More() {
		var block JSONBlock
		// decode an array value (Message)
		err := decoder.Decode(&block)
		if err != nil {
			return err
		}
		if block.ID == "0" {
			continue
		}
		parentHashes := make([]*externalapi.DomainHash, len(block.Parents))
		var ok bool
		for i, parentID := range block.Parents {
			parentHashes[i], ok = parentsMap[parentID]
			if !ok {
				return errors.Errorf("Couldn't find blockID: %s", parentID)
			}
		}
		blockHash, _, err := tc.AddBlock(parentHashes, nil, nil)
		if err != nil {
			return err
		}
		parentsMap[block.ID] = blockHash
	}
	return nil
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
