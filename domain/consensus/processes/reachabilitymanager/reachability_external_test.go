package reachabilitymanager_test

import (
	"math"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/model"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
)

func TestAddChildThatPointsDirectlyToTheSelectedParentChainBelowReindexRoot(t *testing.T) {
	reachabilityReindexWindow := uint64(10)
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()
		tc, tearDown, err := factory.NewTestConsensus(consensusConfig,
			"TestAddChildThatPointsDirectlyToTheSelectedParentChainBelowReindexRoot")
		if err != nil {
			t.Fatalf("NewTestConsensus: %+v", err)
		}
		defer tearDown(false)

		tc.ReachabilityManager().SetReachabilityReindexWindow(reachabilityReindexWindow)

		stagingArea := model.NewStagingArea()

		reindexRoot, err := tc.ReachabilityDataStore().ReachabilityReindexRoot(tc.DatabaseContext(), stagingArea)
		if err != nil {
			t.Fatalf("ReachabilityReindexRoot: %s", err)
		}

		if !reindexRoot.Equal(consensusConfig.GenesisHash) {
			t.Fatalf("reindex root is expected to initially be genesis")
		}

		// Add a block on top of the genesis block
		chainRootBlock, _, err := tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		// Add chain of reachabilityReindexWindow blocks above chainRootBlock.
		// This should move the reindex root
		chainRootBlockTipHash := chainRootBlock
		for i := uint64(0); i < reachabilityReindexWindow; i++ {
			chainBlock, _, err := tc.AddBlock([]*externalapi.DomainHash{chainRootBlockTipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}
			chainRootBlockTipHash = chainBlock
		}

		newReindexRoot, err := tc.ReachabilityDataStore().ReachabilityReindexRoot(tc.DatabaseContext(), stagingArea)
		if err != nil {
			t.Fatalf("ReachabilityReindexRoot: %s", err)
		}

		if newReindexRoot.Equal(reindexRoot) {
			t.Fatalf("reindex root is expected to change")
		}

		// Add enough blocks over genesis to test also the case where the first
		// level (genesis in this case) runs out of slack
		slackSize := tc.ReachabilityManager().ReachabilityReindexSlack()
		blocksToAdd := uint64(math.Log2(float64(slackSize))) + 2
		for i := uint64(0); i < blocksToAdd; i++ {
			_, _, err = tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}
		}

		err = tc.ReachabilityManager().ValidateIntervals(consensusConfig.GenesisHash)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestUpdateReindexRoot(t *testing.T) {
	reachabilityReindexWindow := uint64(10)
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()
		tc, tearDown, err := factory.NewTestConsensus(consensusConfig, "TestUpdateReindexRoot")
		if err != nil {
			t.Fatalf("NewTestConsensus: %s", err)
		}
		defer tearDown(false)

		tc.ReachabilityManager().SetReachabilityReindexWindow(reachabilityReindexWindow)

		stagingArea := model.NewStagingArea()

		intervalSize := func(hash *externalapi.DomainHash) uint64 {
			data, err := tc.ReachabilityDataStore().ReachabilityData(tc.DatabaseContext(), stagingArea, hash)
			if err != nil {
				t.Fatalf("ReachabilityData: %s", err)
			}
			return data.Interval().End - data.Interval().Start + 1
		}

		// Add two blocks on top of the genesis block
		chain1RootBlock, _, err := tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		chain2RootBlock, _, err := tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		// Make two chains of size reachabilityReindexWindow and check that the reindex root is not changed.
		chain1Tip, chain2Tip := chain1RootBlock, chain2RootBlock
		for i := uint64(0); i < reachabilityReindexWindow-1; i++ {
			var err error
			chain1Tip, _, err = tc.AddBlock([]*externalapi.DomainHash{chain1Tip}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			chain2Tip, _, err = tc.AddBlock([]*externalapi.DomainHash{chain2Tip}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			reindexRoot, err := tc.ReachabilityDataStore().ReachabilityReindexRoot(tc.DatabaseContext(), stagingArea)
			if err != nil {
				t.Fatalf("ReachabilityReindexRoot: %s", err)
			}

			if !reindexRoot.Equal(consensusConfig.GenesisHash) {
				t.Fatalf("reindex root unexpectedly moved")
			}
		}

		// Add another block over chain1. This will move the reindex root to chain1RootBlock
		_, _, err = tc.AddBlock([]*externalapi.DomainHash{chain1Tip}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		// Make sure that chain1RootBlock is now the reindex root
		reindexRoot, err := tc.ReachabilityDataStore().ReachabilityReindexRoot(tc.DatabaseContext(), stagingArea)
		if err != nil {
			t.Fatalf("ReachabilityReindexRoot: %s", err)
		}

		if !reindexRoot.Equal(chain1RootBlock) {
			t.Fatalf("chain1RootBlock is not the reindex root after reindex")
		}

		// Make sure that tight intervals have been applied to chain2. Since
		// we added reachabilityReindexWindow-1 blocks to chain2, the size
		// of the interval at its root should be equal to reachabilityReindexWindow
		if intervalSize(chain2RootBlock) != reachabilityReindexWindow {
			t.Fatalf("got unexpected chain2RootBlock interval. Want: %d, got: %d",
				intervalSize(chain2RootBlock), reachabilityReindexWindow)
		}

		// Make sure that the rest of the interval has been allocated to
		// chain1RootNode, minus slack from both sides
		expectedChain1RootIntervalSize := intervalSize(consensusConfig.GenesisHash) - 1 -
			intervalSize(chain2RootBlock) - 2*tc.ReachabilityManager().ReachabilityReindexSlack()
		if intervalSize(chain1RootBlock) != expectedChain1RootIntervalSize {
			t.Fatalf("got unexpected chain1RootBlock interval. Want: %d, got: %d",
				intervalSize(chain1RootBlock), expectedChain1RootIntervalSize)
		}

		err = tc.ReachabilityManager().ValidateIntervals(consensusConfig.GenesisHash)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestReindexIntervalsEarlierThanReindexRoot(t *testing.T) {
	reachabilityReindexWindow := uint64(10)
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()
		tc, tearDown, err := factory.NewTestConsensus(consensusConfig, "TestUpdateReindexRoot")
		if err != nil {
			t.Fatalf("NewTestConsensus: %+v", err)
		}
		defer tearDown(false)

		tc.ReachabilityManager().SetReachabilityReindexWindow(reachabilityReindexWindow)

		stagingArea := model.NewStagingArea()

		intervalSize := func(hash *externalapi.DomainHash) uint64 {
			data, err := tc.ReachabilityDataStore().ReachabilityData(tc.DatabaseContext(), stagingArea, hash)
			if err != nil {
				t.Fatalf("ReachabilityData: %s", err)
			}
			return data.Interval().End - data.Interval().Start + 1
		}

		// Add three children to the genesis: leftBlock, centerBlock, rightBlock
		leftBlock, _, err := tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		centerBlock, _, err := tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		rightBlock, _, err := tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		// Add a chain of reachabilityReindexWindow blocks above centerBlock.
		// This will move the reindex root to centerBlock
		centerTipHash := centerBlock
		for i := uint64(0); i < reachabilityReindexWindow; i++ {
			var err error
			centerTipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{centerTipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}
		}

		// Make sure that centerBlock is now the reindex root
		reindexRoot, err := tc.ReachabilityDataStore().ReachabilityReindexRoot(tc.DatabaseContext(), stagingArea)
		if err != nil {
			t.Fatalf("ReachabilityReindexRoot: %s", err)
		}

		if !reindexRoot.Equal(centerBlock) {
			t.Fatalf("centerBlock is not the reindex root after reindex")
		}

		// Get the current interval for leftBlock. The reindex should have
		// resulted in a tight interval there
		if intervalSize(leftBlock) != 1 {
			t.Fatalf("leftBlock interval not tight after reindex")
		}

		// Get the current interval for rightBlock. The reindex should have
		// resulted in a tight interval there
		if intervalSize(rightBlock) != 1 {
			t.Fatalf("rightBlock interval not tight after reindex")
		}

		err = tc.ReachabilityManager().ValidateIntervals(consensusConfig.GenesisHash)
		if err != nil {
			t.Fatal(err)
		}

		// Add a chain of reachabilityReindexWindow - 1 blocks above leftBlock.
		// Each addition will trigger a low-than-reindex-root reindex. We
		// expect the centerInterval to shrink by 1 each time, but its child
		// to remain unaffected

		leftTipHash := leftBlock
		for i := uint64(0); i < reachabilityReindexWindow-1; i++ {
			var err error
			leftTipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{leftTipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			err = tc.ReachabilityManager().ValidateIntervals(consensusConfig.GenesisHash)
			if err != nil {
				t.Fatal(err)
			}
		}

		// Add a chain of reachabilityReindexWindow - 1 blocks above rightBlock.
		// Each addition will trigger a low-than-reindex-root reindex. We
		// expect the centerInterval to shrink by 1 each time, but its child
		// to remain unaffected
		rightTipHash := rightBlock
		for i := uint64(0); i < reachabilityReindexWindow-1; i++ {
			var err error
			rightTipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{rightTipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			err = tc.ReachabilityManager().ValidateIntervals(consensusConfig.GenesisHash)
			if err != nil {
				t.Fatal(err)
			}
		}

		err = tc.ReachabilityManager().ValidateIntervals(consensusConfig.GenesisHash)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestTipsAfterReindexIntervalsEarlierThanReindexRoot(t *testing.T) {
	reachabilityReindexWindow := uint64(10)
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()
		tc, tearDown, err := factory.NewTestConsensus(consensusConfig, "TestUpdateReindexRoot")
		if err != nil {
			t.Fatalf("NewTestConsensus: %s", err)
		}
		defer tearDown(false)

		tc.ReachabilityManager().SetReachabilityReindexWindow(reachabilityReindexWindow)

		// Add a chain of reachabilityReindexWindow + 1 blocks above the genesis.
		// This will set the reindex root to the child of genesis
		chainTipHash := consensusConfig.GenesisHash
		for i := uint64(0); i < reachabilityReindexWindow+1; i++ {
			chainTipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{chainTipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}
		}

		// Add another block above the genesis block. This will trigger an
		// earlier-than-reindex-root reindex
		sideBlock, _, err := tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		// Add a block whose parents are the chain tip and the side block.
		// We expect this not to fail
		_, _, err = tc.AddBlock([]*externalapi.DomainHash{sideBlock}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		err = tc.ReachabilityManager().ValidateIntervals(consensusConfig.GenesisHash)
		if err != nil {
			t.Fatal(err)
		}
	})
}
