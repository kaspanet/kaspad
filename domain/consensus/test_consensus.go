package consensus

import (
	"encoding/json"
	"io"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/pkg/errors"
)

type testConsensus struct {
	*consensus
	dagParams *dagconfig.Params
	database  database.Database

	testBlockBuilder          testapi.TestBlockBuilder
	testReachabilityManager   testapi.TestReachabilityManager
	testConsensusStateManager testapi.TestConsensusStateManager
	testTransactionValidator  testapi.TestTransactionValidator
}

func (tc *testConsensus) DAGParams() *dagconfig.Params {
	return tc.dagParams
}

func (tc *testConsensus) BuildBlockWithParents(parentHashes []*externalapi.DomainHash,
	coinbaseData *externalapi.DomainCoinbaseData, transactions []*externalapi.DomainTransaction) (
	*externalapi.DomainBlock, externalapi.UTXODiff, error) {

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

func (tc *testConsensus) AddUTXOInvalidHeader(parentHashes []*externalapi.DomainHash) (*externalapi.DomainHash,
	*externalapi.BlockInsertionResult, error) {

	// Require write lock because BuildBlockWithParents stages temporary data
	tc.lock.Lock()
	defer tc.lock.Unlock()

	header, err := tc.testBlockBuilder.BuildUTXOInvalidHeader(parentHashes)
	if err != nil {
		return nil, nil, err
	}

	blockInsertionResult, err := tc.blockProcessor.ValidateAndInsertBlock(&externalapi.DomainBlock{
		Header:       header,
		Transactions: nil,
	})
	if err != nil {
		return nil, nil, err
	}

	return consensushashing.HeaderHash(header), blockInsertionResult, nil
}

func (tc *testConsensus) AddUTXOInvalidBlock(parentHashes []*externalapi.DomainHash) (*externalapi.DomainHash,
	*externalapi.BlockInsertionResult, error) {

	// Require write lock because BuildBlockWithParents stages temporary data
	tc.lock.Lock()
	defer tc.lock.Unlock()

	block, err := tc.testBlockBuilder.BuildUTXOInvalidBlock(parentHashes)
	if err != nil {
		return nil, nil, err
	}

	blockInsertionResult, err := tc.blockProcessor.ValidateAndInsertBlock(block)
	if err != nil {
		return nil, nil, err
	}

	return consensushashing.BlockHash(block), blockInsertionResult, nil
}

func (tc *testConsensus) MineJSON(r io.Reader, blockType testapi.MineJSONBlockType) (tips []*externalapi.DomainHash, err error) {
	// jsonBlock is a json representation of a block in mine format
	type jsonBlock struct {
		ID      string   `json:"id"`
		Parents []string `json:"parents"`
	}

	tipSet := map[externalapi.DomainHash]*externalapi.DomainHash{}
	tipSet[*tc.dagParams.GenesisHash] = tc.dagParams.GenesisHash

	parentsMap := make(map[string]*externalapi.DomainHash)
	parentsMap["0"] = tc.dagParams.GenesisHash

	decoder := json.NewDecoder(r)
	// read open bracket
	_, err = decoder.Token()
	if err != nil {
		return nil, err
	}
	// while the array contains values
	for decoder.More() {
		var block jsonBlock
		// decode an array value (Message)
		err := decoder.Decode(&block)
		if err != nil {
			return nil, err
		}
		if block.ID == "0" {
			continue
		}
		parentHashes := make([]*externalapi.DomainHash, len(block.Parents))
		var ok bool
		for i, parentID := range block.Parents {
			parentHashes[i], ok = parentsMap[parentID]
			if !ok {
				return nil, errors.Errorf("Couldn't find blockID: %s", parentID)
			}
			delete(tipSet, *parentHashes[i])
		}

		var blockHash *externalapi.DomainHash
		switch blockType {
		case testapi.MineJSONBlockTypeUTXOValidBlock:
			blockHash, _, err = tc.AddBlock(parentHashes, nil, nil)
			if err != nil {
				return nil, err
			}
		case testapi.MineJSONBlockTypeUTXOInvalidBlock:
			blockHash, _, err = tc.AddUTXOInvalidBlock(parentHashes)
			if err != nil {
				return nil, err
			}
		case testapi.MineJSONBlockTypeUTXOInvalidHeader:
			blockHash, _, err = tc.AddUTXOInvalidHeader(parentHashes)
			if err != nil {
				return nil, err
			}
		default:
			return nil, errors.Errorf("unknwon block type %v", blockType)
		}

		parentsMap[block.ID] = blockHash
		tipSet[*blockHash] = blockHash
	}

	tips = make([]*externalapi.DomainHash, len(tipSet))
	i := 0
	for _, v := range tipSet {
		tips[i] = v
		i++
	}
	return tips, nil
}

func (tc *testConsensus) BuildUTXOInvalidBlock(parentHashes []*externalapi.DomainHash) (*externalapi.DomainBlock, error) {
	// Require write lock because BuildBlockWithParents stages temporary data
	tc.lock.Lock()
	defer tc.lock.Unlock()

	return tc.testBlockBuilder.BuildUTXOInvalidBlock(parentHashes)
}

func (tc *testConsensus) BuildHeaderWithParents(parentHashes []*externalapi.DomainHash) (externalapi.BlockHeader, error) {
	// Require write lock because BuildUTXOInvalidHeader stages temporary data
	tc.lock.Lock()
	defer tc.lock.Unlock()

	return tc.testBlockBuilder.BuildUTXOInvalidHeader(parentHashes)
}
