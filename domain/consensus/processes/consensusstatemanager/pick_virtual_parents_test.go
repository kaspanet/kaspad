package consensusstatemanager_test

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"testing"
	"time"
)

func TestPickVirtualParents(t *testing.T) {
	params := dagconfig.DevnetParams
	params.SkipProofOfWork = true

	factory := consensus.NewFactory()
	testConsensus, teardown, err := factory.NewTestConsensus(&params, false, "TestPickVirtualParents")
	if err != nil {
		t.Fatalf("Error setting up consensus: %+v", err)
	}
	defer teardown(false)

	// Build three chains over the genesis
	for chainIndex := 0; chainIndex < 3; chainIndex++ {
		const chainSize = 1000
		accumulatedValidationTime := time.Duration(0)

		tipHash := params.GenesisHash
		for blockIndex := 0; blockIndex < chainSize; blockIndex++ {
			block, _, err := testConsensus.BuildBlockWithParents([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("Could not build block: %s", err)
			}
			blockHash := consensushashing.BlockHash(block)
			start := time.Now()
			_, err = testConsensus.ValidateAndInsertBlock(block)
			if err != nil {
				t.Fatalf("Failed to validate block %s: %s", blockHash, err)
			}
			validationTime := time.Since(start)
			accumulatedValidationTime += validationTime
			t.Logf("Validated block #%d in chain #%d, took %s", blockIndex, chainIndex, validationTime)
			tipHash = blockHash
		}

		averageValidationTime := accumulatedValidationTime / 1000
		t.Logf("Average validation time for chain #%d: %s", chainIndex, averageValidationTime)
	}
}
