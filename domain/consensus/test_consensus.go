package consensus

import (
	"errors"
	"math"

	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/kaspanet/kaspad/util"
)

type testConsensus struct {
	*consensus
	testBlockBuilder          model.TestBlockBuilder
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

	solveBlock(block)

	// Use blockProcessor.ValidateAndInsertBlock instead of tc.ValidateAndInsertBlock to avoid double-locking
	// the conscensus lock.
	err = tc.blockProcessor.ValidateAndInsertBlock(block)
	if err != nil {
		return nil, err
	}

	return consensusserialization.BlockHash(block), nil
}

func solveBlock(block *externalapi.DomainBlock) {
	targetDifficulty := util.CompactToBig(block.Header.Bits)

	for i := uint64(0); i < math.MaxUint64; i++ {
		block.Header.Nonce = i
		hash := consensusserialization.BlockHash(block)
		if hashes.ToBig(hash).Cmp(targetDifficulty) <= 0 {
			return
		}
	}

	panic(errors.New("went over all the nonce space and couldn't find a single one that gives a valid block"))
}
