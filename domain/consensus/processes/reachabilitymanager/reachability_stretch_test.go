package reachabilitymanager_test

import (
	"compress/gzip"
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
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
	validateMining    = false
)

func LoadJsonDAG(t *testing.T, fileName, testName string, addArbitraryBlocks, useSmallReindexSlack bool) {
	t.Parallel()

	logger.SetLogLevels(logLevel)
	params := dagconfig.SimnetParams
	params.SkipProofOfWork = true
	tc, teardown, err := consensus.NewFactory().NewTestConsensus(&params, false, testName)
	if err != nil {
		t.Fatalf("Error setting up consensus: %+v", err)
	}
	defer teardown(false)

	if useSmallReindexSlack {
		tc.ReachabilityManager().SetReachabilityReindexSlack(10)
	}

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

	err = tc.MineJSON(gzipReader)
	if err != nil {
		t.Fatal(err)
	}

	err = tc.ReachabilityManager().ValidateIntervals(params.GenesisHash)
	if err != nil {
		t.Fatal(err)
	}

	if addArbitraryBlocks {
		// After loading json, add arbitrary blocks all over the DAG to stretch reindex logic
		// and validate intervals post each addition

		blocks, err := tc.ReachabilityManager().GetAllNodes(params.GenesisHash)
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
				validationRoot = params.GenesisHash
			}
			err = tc.ReachabilityManager().ValidateIntervals(validationRoot)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestNoAttack(t *testing.T) {
	fileName := fmt.Sprintf(
		"../../testdata/reachability/noattack-dag-blocks--2^%d-delay-factor--1-k--18.json.gz",
		numBlocksExponent)
	LoadJsonDAG(t, fileName, "TestNoAttack", false, false)
}

func TestAttack(t *testing.T) {
	fileName := fmt.Sprintf(
		"../../testdata/reachability/attack-dag-blocks--2^%d-delay-factor--1-k--18.json.gz",
		numBlocksExponent)
	LoadJsonDAG(t, fileName, "TestAttack", false, false)
}

func TestArbitraryDAG(t *testing.T) {
	fileName := fmt.Sprintf(
		"../../testdata/reachability/noattack-dag-blocks--2^%d-delay-factor--1-k--18.json.gz",
		numBlocksExponent)
	LoadJsonDAG(t, fileName, "TestArbitraryDAG", true, true)
}

func TestArbitraryAttackDAG(t *testing.T) {
	fileName := fmt.Sprintf(
		"../../testdata/reachability/attack-dag-blocks--2^%d-delay-factor--1-k--18.json.gz",
		numBlocksExponent)
	LoadJsonDAG(t, fileName, "TestArbitraryAttackDAG", true, true)
}
