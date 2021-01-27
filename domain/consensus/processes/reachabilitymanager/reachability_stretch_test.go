package reachabilitymanager_test

import (
	"compress/gzip"
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/pkg/errors"
	"math"
	"math/rand"
	"os"
	"testing"
)

// Test configuration
const (
	numBlocksExponent = 12
	logLevel          = "warn"
)

func initializeTest(t *testing.T, testName string) (tc testapi.TestConsensus, teardown func(keepDataDir bool)) {
	t.Parallel()
	logger.SetLogLevels(logLevel)
	params := dagconfig.SimnetParams
	params.SkipProofOfWork = true
	tc, teardown, err := consensus.NewFactory().NewTestConsensus(&params, false, testName)
	if err != nil {
		t.Fatalf("Error setting up consensus: %+v", err)
	}
	return tc, teardown
}

func buildJsonDAG(t *testing.T, tc testapi.TestConsensus, attackJson bool) (tips []*externalapi.DomainHash) {
	filePrefix := "noattack"
	if attackJson {
		filePrefix = "attack"
	}
	fileName := fmt.Sprintf(
		"../../testdata/reachability/%s-dag-blocks--2^%d-delay-factor--1-k--18.json.gz",
		filePrefix, numBlocksExponent)

	f, err := os.Open(fileName)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gzipReader.Close()

	tips, err = tc.MineJSON(gzipReader)
	if err != nil {
		t.Fatal(err)
	}

	err = tc.ReachabilityManager().ValidateIntervals(tc.DAGParams().GenesisHash)
	if err != nil {
		t.Fatal(err)
	}

	return tips
}

func addArbitraryBlocks(t *testing.T, tc testapi.TestConsensus) {
	// After loading json, add arbitrary blocks all over the DAG to stretch
	// reindex logic, and validate intervals post each addition

	blocks, err := tc.ReachabilityManager().GetAllNodes(tc.DAGParams().GenesisHash)
	if err != nil {
		t.Fatal(err)
	}

	numChainsToAdd := len(blocks) / 2 // Multiply the size of the DAG with arbitrary blocks
	maxBlocksInChain := 20
	validationFreq := int(math.Max(1, float64(numChainsToAdd/100)))

	randSource := rand.New(rand.NewSource(33233))

	for i := 0; i < numChainsToAdd; i++ {
		randomIndex := randSource.Intn(len(blocks))
		randomParent := blocks[randomIndex]
		newBlock, _, err := tc.AddUTXOInvalidHeader([]*externalapi.DomainHash{randomParent})
		if err != nil {
			t.Fatal(err)
		}
		blocks = append(blocks, newBlock)
		// Add a random-length chain every few blocks
		if randSource.Intn(8) == 0 {
			numBlocksInChain := randSource.Intn(maxBlocksInChain)
			chainBlock := newBlock
			for j := 0; j < numBlocksInChain; j++ {
				chainBlock, _, err = tc.AddUTXOInvalidHeader([]*externalapi.DomainHash{chainBlock})
				if err != nil {
					t.Fatal(err)
				}
				blocks = append(blocks, chainBlock)
			}
		}
		// Normally, validate intervals for new chain only
		validationRoot := newBlock
		// However every 'validation frequency' blocks validate intervals for entire DAG
		if i%validationFreq == 0 || i == numChainsToAdd-1 {
			validationRoot = tc.DAGParams().GenesisHash
		}
		err = tc.ReachabilityManager().ValidateIntervals(validationRoot)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func addAlternatingReorgBlocks(t *testing.T, tc testapi.TestConsensus, tips []*externalapi.DomainHash) {
	// Create alternating reorgs to test the cases where
	// reindex root is out of current header selected tip chain

	reindexRoot, err := tc.ReachabilityDataStore().ReachabilityReindexRoot(tc.DatabaseContext())
	if err != nil {
		t.Fatal(err)
	}

	// Try finding two tips; one which has reindex root on it's chain (chainTip), and one which
	// does not (reorgTip). The latter is expected to exist in json attack files.
	var chainTip, reorgTip *externalapi.DomainHash
	for _, block := range tips {
		isRootAncestorOfTip, err := tc.ReachabilityManager().IsReachabilityTreeAncestorOf(reindexRoot, block)
		if err != nil {
			t.Fatal(err)
		}
		if isRootAncestorOfTip {
			chainTip = block
		} else {
			reorgTip = block
		}
	}

	if reorgTip == nil {
		t.Fatal(errors.Errorf("DAG from jsom file is expected to contain a tip " +
			"disagreeing with reindex root chain"))
	}

	if chainTip == nil {
		t.Fatal(errors.Errorf("reindex root is not on any header tip chain, this is unexpected behavior"))
	}

	chainTipGHOSTDAGData, err := tc.GHOSTDAGDataStore().Get(tc.DatabaseContext(), chainTip)
	if err != nil {
		t.Fatal(err)
	}

	reorgTipGHOSTDAGData, err := tc.GHOSTDAGDataStore().Get(tc.DatabaseContext(), reorgTip)
	if err != nil {
		t.Fatal(err)
	}

	// Get both chains close to each other (we care about blue score and not
	// blue work because we have SkipProofOfWork=true)
	if chainTipGHOSTDAGData.BlueScore() > reorgTipGHOSTDAGData.BlueScore() {
		blueScoreDiff := int(chainTipGHOSTDAGData.BlueScore() - reorgTipGHOSTDAGData.BlueScore())
		for i := 0; i < blueScoreDiff+5; i++ {
			reorgTip, _, err = tc.AddUTXOInvalidHeader([]*externalapi.DomainHash{reorgTip})
			if err != nil {
				t.Fatal(err)
			}
		}
	} else {
		blueScoreDiff := int(reorgTipGHOSTDAGData.BlueScore() - chainTipGHOSTDAGData.BlueScore())
		for i := 0; i < blueScoreDiff+5; i++ {
			chainTip, _, err = tc.AddUTXOInvalidHeader([]*externalapi.DomainHash{chainTip})
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	err = tc.ReachabilityManager().ValidateIntervals(tc.DAGParams().GenesisHash)
	if err != nil {
		t.Fatal(err)
	}

	// Alternate between the chains 200 times
	for i := 0; i < 200; i++ {
		if i%2 == 0 {
			for j := 0; j < 10; j++ {
				chainTip, _, err = tc.AddUTXOInvalidHeader([]*externalapi.DomainHash{chainTip})
				if err != nil {
					t.Fatal(err)
				}
			}
		} else {
			for j := 0; j < 10; j++ {
				reorgTip, _, err = tc.AddUTXOInvalidHeader([]*externalapi.DomainHash{reorgTip})
				if err != nil {
					t.Fatal(err)
				}
			}
		}

		err = tc.ReachabilityManager().ValidateIntervals(tc.DAGParams().GenesisHash)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Since current logic switches reindex root chain with reindex slack threshold - at last make the switch happen
	for i := 0; i < int(tc.ReachabilityManager().ReachabilityReindexSlack())+10; i++ {
		reorgTip, _, err = tc.AddUTXOInvalidHeader([]*externalapi.DomainHash{reorgTip})
		if err != nil {
			t.Fatal(err)
		}
	}

	err = tc.ReachabilityManager().ValidateIntervals(tc.DAGParams().GenesisHash)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNoAttack(t *testing.T) {
	tc, teardown := initializeTest(t, "TestNoAttack")
	defer teardown(false)
	buildJsonDAG(t, tc, false)
}

func TestAttack(t *testing.T) {
	tc, teardown := initializeTest(t, "TestAttack")
	defer teardown(false)
	buildJsonDAG(t, tc, true)
}

func TestNoAttackFuzzy(t *testing.T) {
	tc, teardown := initializeTest(t, "TestNoAttackFuzzy")
	defer teardown(false)
	tc.ReachabilityManager().SetReachabilityReindexSlack(10)
	buildJsonDAG(t, tc, false)
	addArbitraryBlocks(t, tc)
}

func TestAttackFuzzy(t *testing.T) {
	tc, teardown := initializeTest(t, "TestAttackFuzzy")
	defer teardown(false)
	tc.ReachabilityManager().SetReachabilityReindexSlack(10)
	buildJsonDAG(t, tc, true)
	addArbitraryBlocks(t, tc)
}

func TestAttackAlternateReorg(t *testing.T) {
	tc, teardown := initializeTest(t, "TestAttackAlternateReorg")
	defer teardown(false)
	tc.ReachabilityManager().SetReachabilityReindexSlack(256)
	tips := buildJsonDAG(t, tc, true)
	addAlternatingReorgBlocks(t, tc, tips)
}
