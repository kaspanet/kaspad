package main

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"
	"github.com/kaspanet/kaspad/stability-tests/common"
	"github.com/kaspanet/kaspad/stability-tests/common/mine"
	"github.com/pkg/errors"
)

const leveldbCacheSizeMiB = 256

func prepareBlocks() (blocks []*externalapi.DomainBlock, topBlock *externalapi.DomainBlock, err error) {
	config := activeConfig()
	testDatabaseDir, err := common.TempDir("minejson")
	if err != nil {
		return nil, nil, err
	}
	db, err := ldb.NewLevelDB(testDatabaseDir, leveldbCacheSizeMiB)
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()

	testConsensus, tearDownFunc, err := consensus.NewFactory().NewTestConsensus(config.ActiveNetParams, false, "prepareBlocks")
	if err != nil {
		return nil, nil, err
	}
	defer tearDownFunc(true)

	virtualSelectedParent, err := testConsensus.GetVirtualSelectedParent()
	if err != nil {
		return nil, nil, err
	}
	currentParentHash := virtualSelectedParent

	blocksCount := config.OrphanChainLength + 1
	blocks = make([]*externalapi.DomainBlock, 0, blocksCount)

	for i := 0; i < blocksCount; i++ {
		block, _, err := testConsensus.BuildBlockWithParents(
			[]*externalapi.DomainHash{currentParentHash},
			&externalapi.DomainCoinbaseData{ScriptPublicKey: &externalapi.ScriptPublicKey{}},
			[]*externalapi.DomainTransaction{})
		if err != nil {
			return nil, nil, errors.Wrap(err, "error in BuildBlockWithParents")
		}

		mine.SolveBlock(block)
		_, err = testConsensus.ValidateAndInsertBlock(block)
		if err != nil {
			return nil, nil, errors.Wrap(err, "error in ValidateAndInsertBlock")
		}

		blocks = append(blocks, block)
		currentParentHash = consensushashing.BlockHash(block)
	}

	return blocks[:len(blocks)-1], blocks[len(blocks)-1], nil
}
