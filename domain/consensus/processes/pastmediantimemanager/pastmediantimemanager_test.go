package pastmediantimemanager_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	consensusdatabase "github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockheaderstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockrelationstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/ghostdagdatastore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/reachabilitydatastore"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/processes/dagtopologymanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/dagtraversalmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/ghostdagmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/pastmediantimemanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/reachabilitymanager"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type testContext struct {
	BlockRelationStore    model.BlockRelationStore
	BlockHeaderStore      model.BlockHeaderStore
	DagTopologyManager    model.DAGTopologyManager
	GhostdagDataStore     model.GHOSTDAGDataStore
	GhostdagManager       model.GHOSTDAGManager
	ReachabilityDataStore model.ReachabilityDataStore
	PastMedianTimeManager model.PastMedianTimeManager
}

func setupTestContext(dbManager model.DBManager, dagParams *dagconfig.Params) *testContext {
	// Data Structures
	blockHeaderStore := blockheaderstore.New()
	blockRelationStore := blockrelationstore.New()
	reachabilityDataStore := reachabilitydatastore.New()
	ghostdagDataStore := ghostdagdatastore.New()

	// Processes
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
	dagTraversalManager := dagtraversalmanager.New(
		dbManager,
		dagTopologyManager,
		ghostdagDataStore,
		ghostdagManager)
	pastMedianTimeManager := pastmediantimemanager.New(
		dagParams.TimestampDeviationTolerance,
		dbManager,
		dagTraversalManager,
		blockHeaderStore)

	return &testContext{
		BlockRelationStore:    blockRelationStore,
		BlockHeaderStore:      blockHeaderStore,
		DagTopologyManager:    dagTopologyManager,
		GhostdagDataStore:     ghostdagDataStore,
		GhostdagManager:       ghostdagManager,
		ReachabilityDataStore: reachabilityDataStore,
		PastMedianTimeManager: pastMedianTimeManager,
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
	testContext.BlockHeaderStore.Stage(blockHash, header)
	testContext.BlockRelationStore.StageBlockRelation(blockHash, &model.BlockRelations{
		Parents: parents,
	})

	if len(parents) > 0 {
		err := testContext.GhostdagManager.GHOSTDAG(blockHash)
		if err != nil {
			return nil, err
		}
	} else {
		testContext.GhostdagDataStore.Stage(blockHash, &model.BlockGHOSTDAGData{})
	}
	return blockHash, nil
}

func TestPastMedianTime(t *testing.T) {
	dbManager, teardownFunc, err := setupDBManager(t.Name())
	if err != nil {
		t.Fatalf("Failed to setup DBManager instance: %v", err)
	}
	defer teardownFunc()

	dagParams := &dagconfig.SimnetParams
	testContext := setupTestContext(dbManager, dagParams)

	blockVersion := int32(0x10000000)
	numBlocks := uint32(300)
	blocks := make([]*externalapi.DomainHash, numBlocks)
	genesisTimestamp := mstime.Now()
	blockTime := genesisTimestamp
	genesisHash, err := newTestBlock(testContext, nil, blockVersion, 0, genesisTimestamp.UnixMilliseconds())
	if err != nil {
		t.Fatalf("newTestBlock: %v", err)
	}
	blocks[0] = genesisHash

	for i := uint32(1); i < numBlocks; i++ {
		blockTime = blockTime.Add(time.Second)
		blockHash, err := newTestBlock(testContext, []*externalapi.DomainHash{blocks[i-1]}, blockVersion, 0, blockTime.UnixMilliseconds())
		if err != nil {
			t.Fatalf("newTestBlock: %v", err)
		}
		blocks[i] = blockHash
	}

	tests := []struct {
		blockNumber                      uint32
		expectedMillisecondsSinceGenesis int64
	}{
		{
			blockNumber:                      262,
			expectedMillisecondsSinceGenesis: 130000,
		},
		{
			blockNumber:                      270,
			expectedMillisecondsSinceGenesis: 138000,
		},
		{
			blockNumber:                      240,
			expectedMillisecondsSinceGenesis: 108000,
		},
		{
			blockNumber:                      5,
			expectedMillisecondsSinceGenesis: 0,
		},
	}

	for _, test := range tests {
		blockTime, err := testContext.PastMedianTimeManager.PastMedianTime(blocks[test.blockNumber])
		if err != nil {
			t.Fatalf("PastMedianTime: %v", err)
		}
		millisecondsSinceGenesis := blockTime - genesisTimestamp.UnixMilliseconds()

		if millisecondsSinceGenesis != test.expectedMillisecondsSinceGenesis {
			t.Fatalf("TestCalcPastMedianTime: expected past median time of block %v to be %v milliseconds "+
				"from genesis but got %v",
				test.blockNumber, test.expectedMillisecondsSinceGenesis, millisecondsSinceGenesis)
		}
	}
}
