package domain_test

import (
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"
	"strings"
	"testing"
)

func TestCreateTemporaryConsensus(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		db, err := ldb.NewLevelDB(t.TempDir(), 8)
		if err != nil {
			t.Fatalf("NewLevelDB: %+v", err)
		}

		domainInstance, err := domain.New(consensusConfig, db)
		if err != nil {
			t.Fatalf("New: %+v", err)
		}

		err = domainInstance.CreateTemporaryConsensus()
		if err != nil {
			t.Fatalf("CreateTemporaryConsensus: %+v", err)
		}

		err = domainInstance.CreateTemporaryConsensus()
		if !strings.Contains(err.Error(), "cannot have more than one inactive prefix") {
			t.Fatalf("unexpected error %+v", err)
		}

		coinbaseData := &externalapi.DomainCoinbaseData{
			ScriptPublicKey: &externalapi.ScriptPublicKey{},
			ExtraData:       []byte{},
		}
		block, err := domainInstance.TemporaryConsensus().BuildBlock(coinbaseData, nil)
		if err != nil {
			t.Fatalf("BuildBlock: %+v", err)
		}

		_, err = domainInstance.TemporaryConsensus().ValidateAndInsertBlock(block)
		if err != nil {
			t.Fatalf("ValidateAndInsertBlock: %+v", err)
		}

		blockHash := consensushashing.BlockHash(block)
		blockInfo, err := domainInstance.TemporaryConsensus().GetBlockInfo(blockHash)
		if err != nil {
			t.Fatalf("GetBlockInfo: %+v", err)
		}

		if !blockInfo.Exists {
			t.Fatalf("block not found on temporary consensus")
		}

		blockInfo, err = domainInstance.Consensus().GetBlockInfo(blockHash)
		if err != nil {
			t.Fatalf("GetBlockInfo: %+v", err)
		}

		if blockInfo.Exists {
			t.Fatalf("a block from temporary consensus was found on consensus")
		}

		err = domainInstance.CommitTemporaryConsensus()
		if err != nil {
			t.Fatalf("CommitTemporaryConsensus: %+v", err)
		}

		blockInfo, err = domainInstance.Consensus().GetBlockInfo(blockHash)
		if err != nil {
			t.Fatalf("GetBlockInfo: %+v", err)
		}

		if !blockInfo.Exists {
			t.Fatalf("a block from temporary consensus was not found on consensus after commit")
		}

		// Now we create a new temporary consensus and check that it's deleted once we init a new domain. We also
		// validate that the main consensus persisted the data from the committed temp consensus.
		err = domainInstance.CreateTemporaryConsensus()
		if err != nil {
			t.Fatalf("CreateTemporaryConsensus: %+v", err)
		}

		_, err = domainInstance.TemporaryConsensus().ValidateAndInsertBlock(block)
		if err != nil {
			t.Fatalf("ValidateAndInsertBlock: %+v", err)
		}

		domainInstance2, err := domain.New(consensusConfig, db)
		if err != nil {
			t.Fatalf("New: %+v", err)
		}

		blockInfo, err = domainInstance2.Consensus().GetBlockInfo(blockHash)
		if err != nil {
			t.Fatalf("GetBlockInfo: %+v", err)
		}

		if !blockInfo.Exists {
			t.Fatalf("a block from committed temporary consensus was not persisted to the active consensus")
		}

		err = domainInstance2.CreateTemporaryConsensus()
		if err != nil {
			t.Fatalf("CreateTemporaryConsensus: %+v", err)
		}

		blockInfo, err = domainInstance2.TemporaryConsensus().GetBlockInfo(blockHash)
		if err != nil {
			t.Fatalf("GetBlockInfo: %+v", err)
		}

		if blockInfo.Exists {
			t.Fatalf("block from previous temp consensus shouldn't be found on a fresh temp consensus")
		}
	})
}
