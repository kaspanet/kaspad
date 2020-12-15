package consensusstatemanager_test

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"testing"
)

func TestFindSelectedParentChainChanges(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		factory := consensus.NewFactory()
		consensus, teardown, err := factory.NewTestConsensus(params, "TestFindSelectedParentChainChanges")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		// Add block A over the genesis
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
			t.Fatalf("The `added` slice contains an unexpected amount of items after inserting block A. "+
				"Want: %d, got: %d", 1, len(blockASelectedParentChainChanges.Added))
		}
		if *blockASelectedParentChainChanges.Added[0] != *blockAHash {
			t.Fatalf("The `added` slice contains an unexpected hash. Want: %s, got: %s",
				blockAHash, blockASelectedParentChainChanges.Added[0])
		}

		// Add block B over the genesis
		blockBHash, _, err := consensus.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error adding block B: %+v", err)
		}

		// Figure out which among blocks A and B is NOT the virtual selected parent
		virtualGHOSTDAGData, err := consensus.GHOSTDAGDataStore().Get(consensus.DatabaseContext(), model.VirtualBlockHash)
		if err != nil {
			t.Fatalf("Error getting virtual GHOSTDAG data: %+v", err)
		}
		virtualSelectedParent := virtualGHOSTDAGData.SelectedParent()
		notVirtualSelectedParent := blockAHash
		if *virtualSelectedParent == *blockAHash {
			notVirtualSelectedParent = blockBHash
		}

		// Add block C over the block that isn't the current virtual's selected parent
		// We expect this to cause a reorg
		blockCHash, blockCInsertionResult, err := consensus.AddBlock([]*externalapi.DomainHash{notVirtualSelectedParent}, nil, nil)
		if err != nil {
			t.Fatalf("Error adding block C: %+v", err)
		}
		blockCSelectedParentChainChanges := blockCInsertionResult.SelectedParentChainChanges

		// Make sure that the removed slice contains only the block that was previously
		// the selected parent
		if len(blockCSelectedParentChainChanges.Removed) != 1 {
			t.Fatalf("The `removed` slice contains an unexpected amount of items after inserting block C. "+
				"Want: %d, got: %d", 1, len(blockCSelectedParentChainChanges.Removed))
		}
		if *blockCSelectedParentChainChanges.Removed[0] != *virtualSelectedParent {
			t.Fatalf("The `removed` slice contains an unexpected hash. "+
				"Want: %s, got: %s", virtualSelectedParent, blockCSelectedParentChainChanges.Removed[0])
		}

		// Make sure that the added slice contains the block that was NOT previously
		// the selected parent and blockCHash, in that order
		if len(blockCSelectedParentChainChanges.Added) != 2 {
			t.Fatalf("The `added` slice contains an unexpected amount of items after inserting block C. "+
				"Want: %d, got: %d", 2, len(blockCSelectedParentChainChanges.Added))
		}
		if *blockCSelectedParentChainChanges.Added[0] != *notVirtualSelectedParent {
			t.Fatalf("The `added` slice contains an unexpected hash as the first item. "+
				"Want: %s, got: %s", notVirtualSelectedParent, blockCSelectedParentChainChanges.Added[0])
		}
		if *blockCSelectedParentChainChanges.Added[1] != *blockCHash {
			t.Fatalf("The `added` slice contains an unexpected hash as the second item. "+
				"Want: %s, got: %s", blockCHash, blockCSelectedParentChainChanges.Added[1])
		}
	})
}
