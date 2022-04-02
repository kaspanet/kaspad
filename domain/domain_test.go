package domain_test

import (
	"fmt"
	"math/big"

	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/miningmanager/mempool"
	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestCreateStagingConsensus(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		dataDir, err := ioutil.TempDir("", fmt.Sprintf("TestCreateStagingConsensus-%s", consensusConfig.Name))
		if err != nil {
			t.Fatalf("ioutil.TempDir: %+v", err)
		}
		defer os.RemoveAll(dataDir)

		db, err := ldb.NewLevelDB(dataDir, 8)
		if err != nil {
			t.Fatalf("NewLevelDB: %+v", err)
		}

		domainInstance, err := domain.New(consensusConfig, mempool.DefaultConfig(&consensusConfig.Params), db)
		if err != nil {
			t.Fatalf("New: %+v", err)
		}

		err = domainInstance.InitStagingConsensusWithoutGenesis()
		if err != nil {
			t.Fatalf("InitStagingConsensusWithoutGenesis: %+v", err)
		}

		err = domainInstance.InitStagingConsensusWithoutGenesis()
		if !strings.Contains(err.Error(), "cannot create staging consensus when a staging consensus already exists") {
			t.Fatalf("unexpected error: %+v", err)
		}

		addGenesisToStagingConsensus := func() {
			genesisWithTrustedData := &externalapi.BlockWithTrustedData{
				Block:     consensusConfig.GenesisBlock,
				DAAWindow: nil,
				GHOSTDAGData: []*externalapi.BlockGHOSTDAGDataHashPair{
					{
						GHOSTDAGData: externalapi.NewBlockGHOSTDAGData(0, big.NewInt(0), model.VirtualGenesisBlockHash, nil, nil, make(map[externalapi.DomainHash]externalapi.KType)),
						Hash:         consensusConfig.GenesisHash,
					},
				},
			}
			_, err = domainInstance.StagingConsensus().ValidateAndInsertBlockWithTrustedData(genesisWithTrustedData, true)
			if err != nil {
				t.Fatalf("ValidateAndInsertBlockWithTrustedData: %+v", err)
			}
		}

		addGenesisToStagingConsensus()

		coinbaseData := &externalapi.DomainCoinbaseData{
			ScriptPublicKey: &externalapi.ScriptPublicKey{},
			ExtraData:       []byte{},
		}
		block, err := domainInstance.StagingConsensus().BuildBlock(coinbaseData, nil)
		if err != nil {
			t.Fatalf("BuildBlock: %+v", err)
		}

		_, err = domainInstance.StagingConsensus().ValidateAndInsertBlock(block, true)
		if err != nil {
			t.Fatalf("ValidateAndInsertBlock: %+v", err)
		}

		blockHash := consensushashing.BlockHash(block)
		blockInfo, err := domainInstance.StagingConsensus().GetBlockInfo(blockHash)
		if err != nil {
			t.Fatalf("GetBlockInfo: %+v", err)
		}

		if !blockInfo.Exists {
			t.Fatalf("block not found on staging consensus")
		}

		blockInfo, err = domainInstance.Consensus().GetBlockInfo(blockHash)
		if err != nil {
			t.Fatalf("GetBlockInfo: %+v", err)
		}

		if blockInfo.Exists {
			t.Fatalf("a block from staging consensus was found on consensus")
		}

		err = domainInstance.CommitStagingConsensus()
		if err != nil {
			t.Fatalf("CommitStagingConsensus: %+v", err)
		}

		blockInfo, err = domainInstance.Consensus().GetBlockInfo(blockHash)
		if err != nil {
			t.Fatalf("GetBlockInfo: %+v", err)
		}

		if !blockInfo.Exists {
			t.Fatalf("a block from staging consensus was not found on consensus after commit")
		}

		// Now we create a new staging consensus and check that it's deleted once we init a new domain. We also
		// validate that the main consensus persisted the data from the committed temp consensus.
		err = domainInstance.InitStagingConsensusWithoutGenesis()
		if err != nil {
			t.Fatalf("InitStagingConsensusWithoutGenesis: %+v", err)
		}

		addGenesisToStagingConsensus()
		_, err = domainInstance.StagingConsensus().ValidateAndInsertBlock(block, true)
		if err != nil {
			t.Fatalf("ValidateAndInsertBlock: %+v", err)
		}

		domainInstance2, err := domain.New(consensusConfig, mempool.DefaultConfig(&consensusConfig.Params), db)
		if err != nil {
			t.Fatalf("New: %+v", err)
		}

		blockInfo, err = domainInstance2.Consensus().GetBlockInfo(blockHash)
		if err != nil {
			t.Fatalf("GetBlockInfo: %+v", err)
		}

		if !blockInfo.Exists {
			t.Fatalf("a block from committed staging consensus was not persisted to the active consensus")
		}

		err = domainInstance2.InitStagingConsensusWithoutGenesis()
		if err != nil {
			t.Fatalf("InitStagingConsensusWithoutGenesis: %+v", err)
		}

		blockInfo, err = domainInstance2.StagingConsensus().GetBlockInfo(blockHash)
		if err != nil {
			t.Fatalf("GetBlockInfo: %+v", err)
		}

		if blockInfo.Exists {
			t.Fatalf("block from previous temp consensus shouldn't be found on a fresh temp consensus")
		}
	})
}
