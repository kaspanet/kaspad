package consensus_test

import (
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
const(
	numBlocksExponent = 14
	logLevel = "warn"
	validateMining = false
)

func LoadJsonDAG(t *testing.T, fileName, testName string, addArbitraryBlocks, useSmallReindexSlack bool) {
	logger.SetLogLevels(logLevel)
	params := dagconfig.SimnetParams
	params.SkipProofOfWork = true
	tc, teardown, err := consensus.NewFactory().NewTestConsensus(&params, testName)
	if err != nil {
		t.Fatalf("Error setting up consensus: %+v", err)
	}
	defer teardown(false)

	tc.TestParams().ValidateMining = validateMining

	if useSmallReindexSlack {
		tc.ReachabilityManager().SetReachabilityReindexSlack(10)
	}

	f, err := os.Open(fileName)
	if err != nil {
		t.Fatal(err)
	}
	//now := time.Now()
	err = tc.MineJSON(f)
	if err != nil {
		t.Fatal(err)
	}
	//fmt.Printf("passed %ds\n", time.Since(now).Seconds())

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

		numBlocksToAdd := len(blocks) // Double the size of the DAG with arbitrary blocks
		validationFreq := int(math.Max(1, float64(numBlocksToAdd/100)))

		rand.Seed(33233)

		for i := 0; i < numBlocksToAdd; i++ {
			randomIndex := rand.Intn(len(blocks))
			randomParent := blocks[randomIndex]
			newBlock, _, err := tc.AddBlock([]*externalapi.DomainHash{randomParent}, nil, nil)
			blocks = append(blocks, newBlock)

			// Validate intervals every 'validation frequency' blocks
			if i%validationFreq == 0 {
				err = tc.ReachabilityManager().ValidateIntervals(params.GenesisHash)
				if err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}

func TestNoAttack(t *testing.T) {
	fileName := fmt.Sprintf(
		"../../testdata/dags/noattack-dag-blocks--2^%d-delay-factor--1-k--18.json",
		numBlocksExponent)
	LoadJsonDAG(t, fileName, "TestNoAttack", false, false)
}

func TestAttack(t *testing.T) {
	fileName := fmt.Sprintf(
		"../../testdata/dags/attack-dag-blocks--2^%d-delay-factor--1-k--18.json",
		numBlocksExponent)
	LoadJsonDAG(t, fileName, "TestAttack", false, false)
}

func TestArbitraryDAG(t *testing.T) {
	fileName := fmt.Sprintf(
		"../../testdata/dags/noattack-dag-blocks--2^%d-delay-factor--1-k--18.json",
		numBlocksExponent)
	LoadJsonDAG(t, fileName, "TestNoAttack", true, true)
}
