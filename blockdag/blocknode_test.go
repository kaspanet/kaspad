package blockdag

import (
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
	"testing"
)

func TestChainHeight(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	params := dagconfig.SimNetParams
	params.K = 2
	dag, teardownFunc, err := DAGSetup("TestChainHeight", Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("TestChainHeight: Failed to setup DAG instance: %s", err)
	}
	defer teardownFunc()

	prepareAndProcessBlock := func(parents ...*wire.MsgBlock) *wire.MsgBlock {
		parentHashes := make([]*daghash.Hash, len(parents))
		for i, parent := range parents {
			parentHashes[i] = parent.BlockHash()
		}
		daghash.Sort(parentHashes)
		block, err := PrepareBlockForTest(dag, parentHashes, nil)
		if err != nil {
			t.Fatalf("error in PrepareBlockForTest: %s", err)
		}
		utilBlock := util.NewBlock(block)
		isOrphan, delay, err := dag.ProcessBlock(utilBlock, BFNoPoWCheck)
		if err != nil {
			t.Fatalf("error in ProcessBlock: %s", err)
		}
		if delay != 0 {
			t.Fatalf("block is too far in the future")
		}
		if isOrphan {
			t.Fatalf("block was unexpectedly orphan")
		}
		return block
	}

	node0 := dag.dagParams.GenesisBlock
	node1 := prepareAndProcessBlock(dag.dagParams.GenesisBlock)
	node2 := prepareAndProcessBlock(node0)
	node3 := prepareAndProcessBlock(node0)
	node4 := prepareAndProcessBlock(node1, node2, node3)
	node5 := prepareAndProcessBlock(node1, node2, node3)
	node6 := prepareAndProcessBlock(node1, node2, node3)
	node7 := prepareAndProcessBlock(node0)
	node8 := prepareAndProcessBlock(node7)
	node9 := prepareAndProcessBlock(node8)
	node10 := prepareAndProcessBlock(node9, node6)

	// Because nodes 7 & 8 were mined secretly, node10's selected
	// parent will be node6, although node9 is higher. So in this
	// case, node10.height and node10.chainHeight will be different

	tests := []struct {
		block               *wire.MsgBlock
		expectedChainHeight uint64
	}{
		{
			block:               node0,
			expectedChainHeight: 0,
		},
		{
			block:               node1,
			expectedChainHeight: 1,
		},
		{
			block:               node2,
			expectedChainHeight: 1,
		},
		{
			block:               node3,
			expectedChainHeight: 1,
		},
		{
			block:               node4,
			expectedChainHeight: 2,
		},
		{
			block:               node5,
			expectedChainHeight: 2,
		},
		{
			block:               node6,
			expectedChainHeight: 2,
		},
		{
			block:               node7,
			expectedChainHeight: 1,
		},
		{
			block:               node8,
			expectedChainHeight: 2,
		},
		{
			block:               node9,
			expectedChainHeight: 3,
		},
		{
			block:               node10,
			expectedChainHeight: 3,
		},
	}

	for _, test := range tests {
		node := dag.index.LookupNode(test.block.BlockHash())
		if node.chainHeight != test.expectedChainHeight {
			t.Errorf("block %s expected chain height %v but got %v", node, test.expectedChainHeight, node.chainHeight)
		}
		if calculateChainHeight(node) != test.expectedChainHeight {
			t.Errorf("block %s expected calculated chain height %v but got %v", node, test.expectedChainHeight, node.chainHeight)
		}
	}

}
