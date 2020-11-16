package reachabilitymanager_test

import (
	consensusdatabase "github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockrelationstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/ghostdagdatastore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/reachabilitydatastore"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/processes/dagtopologymanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/ghostdagmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/reachabilitymanager"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type testContext struct {
	BlockRelationStore    model.BlockRelationStore
	GhostdagDataStore     model.GHOSTDAGDataStore
	ReachabilityDataStore model.ReachabilityDataStore
	ReachabilityManager   model.ReachabilityManager
	GhostdagManager       model.GHOSTDAGManager
}

func setupTestContext(dbManager model.DBManager, dagParams *dagconfig.Params) *testContext {
	blockRelationStore := blockrelationstore.New()
	reachabilityDataStore := reachabilitydatastore.New()
	ghostdagDataStore := ghostdagdatastore.New()

	reachabilityManager := reachabilitymanager.New(
		dbManager,
		ghostdagDataStore,
		blockRelationStore,
		reachabilityDataStore)

	dagTopologyManager := dagtopologymanager.New(
		dbManager,
		reachabilityManager,
		blockRelationStore)

	ghostdagManager := ghostdagmanager.New(
		dbManager,
		dagTopologyManager,
		ghostdagDataStore,
		model.KType(dagParams.K))

	return &testContext{
		BlockRelationStore:    blockRelationStore,
		GhostdagDataStore:     ghostdagDataStore,
		ReachabilityDataStore: reachabilityDataStore,
		ReachabilityManager:   reachabilityManager,
		GhostdagManager:       ghostdagManager,
	}
}

func setupDBManager(dbName string) (model.DBManager, func(), error) {
	var err error
	tmpDir, err := ioutil.TempDir("", "setupDBManager")
	if err != nil {
		return nil, nil, errors.Errorf("error creating temp dir: %s", err)
	}

	dbPath := filepath.Join(tmpDir, dbName)
	_ = os.RemoveAll(dbPath)
	db, err := ldb.NewLevelDB(dbPath)
	if err != nil {
		return nil, nil, err
	}

	originalLDBOptions := ldb.Options
	ldb.Options = func() *opt.Options {
		return nil
	}

	teardown := func() {
		db.Close()
		ldb.Options = originalLDBOptions
		os.RemoveAll(dbPath)
	}

	dbManager := consensusdatabase.New(db)
	return dbManager, teardown, err
}

func createBlock(header *externalapi.DomainBlockHeader,
	transactions []*externalapi.DomainTransaction) *externalapi.DomainBlock {
	return &externalapi.DomainBlock{
		Header:       header,
		Transactions: transactions,
	}
}

func newTestBlock(testContext *testContext, parents []*externalapi.DomainHash, blockVersion int32, bits uint32, timeInMilliseconds int64) (*externalapi.DomainHash, error) {
	header := &externalapi.DomainBlockHeader{
		Version:              blockVersion,
		ParentHashes:         parents,
		Bits:                 0,
		TimeInMilliseconds:   timeInMilliseconds,
		HashMerkleRoot:       externalapi.DomainHash{},
		AcceptedIDMerkleRoot: externalapi.DomainHash{},
		UTXOCommitment:       externalapi.DomainHash{},
	}

	blockHash := consensusserialization.HeaderHash(header)

	testContext.BlockRelationStore.StageBlockRelation(blockHash, &model.BlockRelations{
		Parents: parents,
	})

	err := testContext.GhostdagManager.GHOSTDAG(blockHash)
	if err != nil {
		return nil, err
	}

	err = testContext.ReachabilityManager.AddBlock(blockHash)
	if err != nil {
		return nil, err
	}

	if len(parents) > 0 {
		err = testContext.ReachabilityManager.UpdateReindexRoot(blockHash)
		if err != nil {
			return nil, err
		}
	}

	return blockHash, nil
}

func TestIsReachabilityTreeAncestorOf(t *testing.T) {
	dbManager, teardownFunc, err := setupDBManager(t.Name())
	if err != nil {
		t.Fatalf("Failed to setup DBManager instance: %v", err)
	}
	defer teardownFunc()

	dagParams := &dagconfig.SimnetParams
	testContext := setupTestContext(dbManager, dagParams)

	blockVersion := int32(0x10000000)
	numBlocks := uint32(5)

	chainABlocks := make([]*externalapi.DomainHash, numBlocks)
	genesisTimestamp := mstime.Now()
	blockTime := genesisTimestamp
	genesisHash, err := newTestBlock(testContext, nil, blockVersion, 0, genesisTimestamp.UnixMilliseconds())
	if err != nil {
		t.Fatalf("newTestBlock: %v", err)
	}
	chainABlocks[0] = genesisHash

	for i := uint32(1); i < numBlocks; i++ {
		blockTime = blockTime.Add(time.Second)
		blockHash, err := newTestBlock(testContext, []*externalapi.DomainHash{chainABlocks[i-1]}, blockVersion, 0, blockTime.UnixMilliseconds())
		if err != nil {
			t.Fatalf("newTestBlock: %v", err)
		}

		chainABlocks[i] = blockHash
	}

	chainBBlocks := make([]*externalapi.DomainHash, numBlocks)
	chainBBlocks[0] = chainABlocks[0]
	for i := uint32(1); i < numBlocks; i++ {
		blockTime = blockTime.Add(time.Second)
		blockHash, err := newTestBlock(testContext, []*externalapi.DomainHash{chainBBlocks[i-1]}, blockVersion, 0, blockTime.UnixMilliseconds())
		if err != nil {
			t.Fatalf("newTestBlock: %v", err)
		}

		chainBBlocks[i] = blockHash
	}

	tests := []struct {
		BlockA         *externalapi.DomainHash
		BlockB         *externalapi.DomainHash
		ExpectedResult bool
	}{
		{
			BlockA:         chainABlocks[0],
			BlockB:         chainABlocks[0],
			ExpectedResult: true,
		},
		{
			BlockA:         chainABlocks[1],
			BlockB:         chainABlocks[1],
			ExpectedResult: true,
		},
		{
			BlockA:         chainABlocks[1],
			BlockB:         chainABlocks[3],
			ExpectedResult: true,
		},
		{
			BlockA:         chainBBlocks[1],
			BlockB:         chainBBlocks[2],
			ExpectedResult: true,
		},
		{
			BlockA:         chainABlocks[0],
			BlockB:         chainBBlocks[3],
			ExpectedResult: true,
		},
		{
			BlockA:         chainABlocks[3],
			BlockB:         chainBBlocks[3],
			ExpectedResult: false,
		},
		{
			BlockA:         chainABlocks[1],
			BlockB:         chainBBlocks[3],
			ExpectedResult: false,
		},
	}

	for _, test := range tests {
		ok, err := testContext.ReachabilityManager.IsDAGAncestorOf(test.BlockA, test.BlockB)
		if err != nil {
			t.Fatalf("ReachabilityManager.IsDAGAncestorOf: %v", err)
		}
		if ok != test.ExpectedResult {
			t.Fatalf("IsDAGAncestorOf should returns %v but got %v", test.ExpectedResult, ok)
		}
	}
}
