package reachabilitymanager_test

import (
	"compress/gzip"
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/logger"
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

func buildJsonDAG(t *testing.T, tc testapi.TestConsensus, attackJson bool) []*externalapi.DomainHash {
	filePrefix := "noattack"
	if attackJson {
		filePrefix = "attack"
	}
	fileName := fmt.Sprintf(
		"../../testdata/reachability/%s-dag-blocks--2^%d-delay-factor--1-k--18.json.gz",
		filePrefix ,  numBlocksExponent)

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

	tips, err := tc.MineJSON(gzipReader)
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
	// After loading json, add arbitrary blocks all over the DAG to stretch reindex logic,
	// and validate intervals post each addition

	blocks, err := tc.ReachabilityManager().GetAllNodes(tc.DAGParams().GenesisHash)
	if err != nil {
		t.Fatal(err)
	}

	numChainsToAdd := len(blocks) // Multiply the size of the DAG with arbitrary blocks
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

func addReorgBlocks(t *testing.T, tc testapi.TestConsensus, tips []*externalapi.DomainHash)  {
	reindexRoot, err := tc.ReachabilityDataStore().ReachabilityReindexRoot(tc.DatabaseContext())
	if err != nil {
		t.Fatal(err)
	}

	reorgTip := tips[0]
	for _, block := range tips {
		isRootAncestorOfTip, err := tc.ReachabilityManager().IsReachabilityTreeAncestorOf(reindexRoot, block)
		if err != nil {
			t.Fatal(err)
		}
		if !isRootAncestorOfTip {
			reorgTip = block
			break
		}
	}

	//print(reorgTip)
	current := reorgTip
	for i := 0; i < 1500; i++ {
		current, _, err = tc.AddUTXOInvalidHeader([]*externalapi.DomainHash{current})
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

func TestNoAttackArbitraryDAG(t *testing.T) {
	tc, teardown := initializeTest(t, "TestNoAttackArbitraryDAG")
	defer teardown(false)
	tc.ReachabilityManager().SetReachabilityReindexSlack(10)
	buildJsonDAG(t, tc, false)
	addArbitraryBlocks(t, tc)
}

func TestAttackArbitraryDAG(t *testing.T) {
	tc, teardown := initializeTest(t, "TestAttackArbitraryDAG")
	defer teardown(false)
	tc.ReachabilityManager().SetReachabilityReindexSlack(10)
	buildJsonDAG(t, tc, true)
	addArbitraryBlocks(t, tc)
}

func TestNoAttackReorgDAG(t *testing.T) {
	tc, teardown := initializeTest(t, "TestNoAttackReorgDAG")
	defer teardown(false)
	tips := buildJsonDAG(t, tc, false)
	addReorgBlocks(t, tc, tips)
}

func TestAttackReorgDAG(t *testing.T) {
	tc, teardown := initializeTest(t, "TestAttackReorgDAG")
	defer teardown(false)
	tips := buildJsonDAG(t, tc, true)
	addReorgBlocks(t, tc, tips)
}
