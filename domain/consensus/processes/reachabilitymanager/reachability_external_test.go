package reachabilitymanager_test

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"testing"
)

func TestAddChildThatPointsDirectlyToTheSelectedParentChainBelowReindexRoot(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		factory := consensus.NewFactory()
		tc, tearDown, err := factory.NewTestConsensus(params, "TestAddChildThatPointsDirectlyToTheSelectedParentChainBelowReindexRoot")
		if err != nil {
			t.Fatalf("NewTestConsensus: %s", err)
		}
		defer tearDown()

		// Add a block on top of the genesis block
		chainRootBlock, err := tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		// Add chain of reachabilityReindexWindow blocks above chainRootBlock.
		// This should move the reindex root
		chainRootBlockTipHash := chainRootBlock
		for i := uint64(0); i < tc.ReachabilityManager().ReachabilityReindexWindow(); i++ {
			chainBlock, err := tc.AddBlock([]*externalapi.DomainHash{chainRootBlockTipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}
			chainRootBlockTipHash = chainBlock
		}

		// Add another block over genesis
		_, err = tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}
	})
}

func TestUpdateReindexRoot(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		factory := consensus.NewFactory()
		tc, tearDown, err := factory.NewTestConsensus(params, "TestUpdateReindexRoot")
		if err != nil {
			t.Fatalf("NewTestConsensus: %s", err)
		}
		defer tearDown()

		intervalSize := func(hash *externalapi.DomainHash) uint64 {
			data, err := tc.ReachabilityDataStore().ReachabilityData(tc.DBReader(), hash)
			if err != nil {
				t.Fatalf("ReachabilityData: %s", err)
			}
			return data.TreeNode.Interval.End - data.TreeNode.Interval.Start + 1
		}

		// Add two blocks on top of the genesis block
		chain1RootBlock, err := tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		chain2RootBlock, err := tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		for i := uint64(0); i < tc.ReachabilityManager().ReachabilityReindexWindow()-1; i++ {
			_, err := tc.AddBlock([]*externalapi.DomainHash{chain1RootBlock}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			_, err = tc.AddBlock([]*externalapi.DomainHash{chain2RootBlock}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			reindexRoot, err := tc.ReachabilityDataStore().ReachabilityReindexRoot(tc.DBReader())
			if err != nil {
				t.Fatalf("ReachabilityReindexRoot: %s", err)
			}

			if *reindexRoot != *params.GenesisHash {
				t.Fatalf("reindex root unexpectedly moved")
			}
		}

		// Add another block over chain1. This will move the reindex root to chain1RootBlock
		_, err = tc.AddBlock([]*externalapi.DomainHash{chain1RootBlock}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		// Make sure that chain1RootBlock is now the reindex root
		reindexRoot, err := tc.ReachabilityDataStore().ReachabilityReindexRoot(tc.DBReader())
		if err != nil {
			t.Fatalf("ReachabilityReindexRoot: %s", err)
		}

		if *reindexRoot != *chain1RootBlock {
			t.Fatalf("chain1RootBlock is not the reindex root after reindex")
		}

		// Make sure that tight intervals have been applied to chain2. Since
		// we added reachabilityReindexWindow-1 blocks to chain2, the size
		// of the interval at its root should be equal to reachabilityReindexWindow
		if intervalSize(chain2RootBlock) != tc.ReachabilityManager().ReachabilityReindexWindow() {
			t.Fatalf("got unexpected chain2RootBlock interval. Want: %d, got: %d",
				intervalSize(chain2RootBlock), tc.ReachabilityManager().ReachabilityReindexWindow())
		}

		// Make sure that the rest of the interval has been allocated to
		// chain1RootNode, minus slack from both sides
		expectedChain1RootIntervalSize := intervalSize(params.GenesisHash) - 1 -
			intervalSize(chain2RootBlock) - 2*tc.ReachabilityManager().ReachabilityReindexWindow()
		if intervalSize(chain1RootBlock) != expectedChain1RootIntervalSize {
			t.Fatalf("got unexpected chain1RootBlock interval. Want: %d, got: %d",
				intervalSize(chain1RootBlock), expectedChain1RootIntervalSize)
		}
	})
}

func TestReindexIntervalsEarlierThanReindexRoot(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		factory := consensus.NewFactory()
		tc, tearDown, err := factory.NewTestConsensus(params, "TestUpdateReindexRoot")
		if err != nil {
			t.Fatalf("NewTestConsensus: %s", err)
		}
		defer tearDown()

		intervalSize := func(hash *externalapi.DomainHash) uint64 {
			data, err := tc.ReachabilityDataStore().ReachabilityData(tc.DBReader(), hash)
			if err != nil {
				t.Fatalf("ReachabilityData: %s", err)
			}
			return data.TreeNode.Interval.End - data.TreeNode.Interval.Start + 1
		}

		// Add three children to the genesis: leftBlock, centerBlock, rightBlock
		leftBlock, err := tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		centerBlock, err := tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		rightBlock, err := tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		// Add a chain of reachabilityReindexWindow blocks above centerBlock.
		// This will move the reindex root to centerBlock
		centerTipHash := centerBlock
		for i := uint64(0); i < tc.ReachabilityManager().ReachabilityReindexWindow(); i++ {
			var err error
			centerTipHash, err = tc.AddBlock([]*externalapi.DomainHash{centerTipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}
		}

		// Make sure that centerBlock is now the reindex root
		reindexRoot, err := tc.ReachabilityDataStore().ReachabilityReindexRoot(tc.DBReader())
		if err != nil {
			t.Fatalf("ReachabilityReindexRoot: %s", err)
		}

		if *reindexRoot != *centerBlock {
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

		// Get the current interval for centerBlock. Its interval should be:
		// genesisInterval - 1 - leftInterval - leftSlack - rightInterval - rightSlack
		expectedCenterInterval := intervalSize(params.GenesisHash) - 1 -
			intervalSize(leftBlock) - tc.ReachabilityManager().ReachabilityReindexSlack() -
			intervalSize(rightBlock) - tc.ReachabilityManager().ReachabilityReindexSlack()
		if intervalSize(centerBlock) != expectedCenterInterval {
			t.Fatalf("unexpected centerBlock interval. Want: %d, got: %d",
				expectedCenterInterval, intervalSize(centerBlock))
		}

		// Add a chain of reachabilityReindexWindow - 1 blocks above leftBlock.
		// Each addition will trigger a low-than-reindex-root reindex. We
		// expect the centerInterval to shrink by 1 each time, but its child
		// to remain unaffected
		cetnerData, err := tc.ReachabilityDataStore().ReachabilityData(tc.DBReader(), centerBlock)
		if err != nil {
			t.Fatalf("ReachabilityData: %s", err)
		}

		treeChildOfCenterBlock := cetnerData.TreeNode.Children[0]
		treeChildOfCenterBlockOriginalIntervalSize := intervalSize(treeChildOfCenterBlock)
		leftTipHash := leftBlock
		for i := uint64(0); i < tc.ReachabilityManager().ReachabilityReindexWindow()-1; i++ {
			var err error
			leftTipHash, err = tc.AddBlock([]*externalapi.DomainHash{leftTipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			expectedCenterInterval--
			if intervalSize(centerBlock) != expectedCenterInterval {
				t.Fatalf("unexpected centerBlock interval. Want: %d, got: %d",
					expectedCenterInterval, intervalSize(centerBlock))
			}

			if intervalSize(treeChildOfCenterBlock) != treeChildOfCenterBlockOriginalIntervalSize {
				t.Fatalf("the interval of centerBlock's child unexpectedly changed")
			}
		}

		// Add a chain of reachabilityReindexWindow - 1 blocks above rightBlock.
		// Each addition will trigger a low-than-reindex-root reindex. We
		// expect the centerInterval to shrink by 1 each time, but its child
		// to remain unaffected
		rightTipHash := rightBlock
		for i := uint64(0); i < tc.ReachabilityManager().ReachabilityReindexWindow()-1; i++ {
			var err error
			rightTipHash, err = tc.AddBlock([]*externalapi.DomainHash{rightTipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			expectedCenterInterval--
			if intervalSize(centerBlock) != expectedCenterInterval {
				t.Fatalf("unexpected centerBlock interval. Want: %d, got: %d",
					expectedCenterInterval, intervalSize(centerBlock))
			}

			if intervalSize(treeChildOfCenterBlock) != treeChildOfCenterBlockOriginalIntervalSize {
				t.Fatalf("the interval of centerBlock's child unexpectedly changed")
			}
		}
	})
}

func TestTipsAfterReindexIntervalsEarlierThanReindexRoot(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		factory := consensus.NewFactory()
		tc, tearDown, err := factory.NewTestConsensus(params, "TestUpdateReindexRoot")
		if err != nil {
			t.Fatalf("NewTestConsensus: %s", err)
		}
		defer tearDown()

		// Add a chain of reachabilityReindexWindow + 1 blocks above the genesis.
		// This will set the reindex root to the child of genesis
		chainTipHash := params.GenesisHash
		for i := uint64(0); i < tc.ReachabilityManager().ReachabilityReindexWindow()+1; i++ {
			chainTipHash, err = tc.AddBlock([]*externalapi.DomainHash{chainTipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}
		}

		// Add another block above the genesis block. This will trigger an
		// earlier-than-reindex-root reindex
		sideBlock, err := tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		// Add a block whose parents are the chain tip and the side block.
		// We expect this not to fail
		_, err = tc.AddBlock([]*externalapi.DomainHash{sideBlock}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}
	})
}
