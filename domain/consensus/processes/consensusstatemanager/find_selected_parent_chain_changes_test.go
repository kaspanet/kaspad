package consensusstatemanager_test

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"testing"
)

func TestFindSelectedParentChainChanges(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		factory := consensus.NewFactory()
		consensus, teardown, err := factory.NewTestConsensus(params, "TestUTXOCommitment")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown()

		// Add block A over the virtual
		blockAHash, blockAInsertionResult, err := consensus.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error adding block A: %+v", err)
		}
		blockASelectedParentChainChanges := blockAInsertionResult.SelectedParentChainChanges

		// Make sure that the removed slice is empty
		if len(blockASelectedParentChainChanges.Removed) > 0 {
			t.Fatalf("The `removed` slice is not empty after inserting block A")
		}

		// Make sure that the added slice contains only blockAHash
		if len(blockASelectedParentChainChanges.Added) != 1 {
			t.Fatalf("The `added` slice contains an unexpected amount of items after inserting block A")
		}
		if *blockASelectedParentChainChanges.Added[0] != *blockAHash {
			t.Fatalf("The `added` slice contains an unexpected hash")
		}
	})
}
