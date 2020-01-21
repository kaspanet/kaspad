package blockdag

import (
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
	"math"
	"testing"
)

func TestChainHeight(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	params := dagconfig.SimnetParams
	params.K = 2
	dag, teardownFunc, err := DAGSetup("TestChainHeight", Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("TestChainHeight: Failed to setup DAG instance: %s", err)
	}
	defer teardownFunc()

	block0 := dag.dagParams.GenesisBlock
	block1 := prepareAndProcessBlock(t, dag, block0)
	block2 := prepareAndProcessBlock(t, dag, block0)
	block3 := prepareAndProcessBlock(t, dag, block0)
	block4 := prepareAndProcessBlock(t, dag, block1, block2, block3)
	block5 := prepareAndProcessBlock(t, dag, block1, block2, block3)
	block6 := prepareAndProcessBlock(t, dag, block1, block2, block3)
	block7 := prepareAndProcessBlock(t, dag, block0)
	block8 := prepareAndProcessBlock(t, dag, block7)
	block9 := prepareAndProcessBlock(t, dag, block8)
	block10 := prepareAndProcessBlock(t, dag, block9, block6)

	// Because nodes 7 & 8 were mined secretly, block10's selected
	// parent will be block6, although block9 is higher. So in this
	// case, block10.height and block10.chainHeight will be different

	tests := []struct {
		block               *wire.MsgBlock
		expectedChainHeight uint64
	}{
		{
			block:               block0,
			expectedChainHeight: 0,
		},
		{
			block:               block1,
			expectedChainHeight: 1,
		},
		{
			block:               block2,
			expectedChainHeight: 1,
		},
		{
			block:               block3,
			expectedChainHeight: 1,
		},
		{
			block:               block4,
			expectedChainHeight: 2,
		},
		{
			block:               block5,
			expectedChainHeight: 2,
		},
		{
			block:               block6,
			expectedChainHeight: 2,
		},
		{
			block:               block7,
			expectedChainHeight: 1,
		},
		{
			block:               block8,
			expectedChainHeight: 2,
		},
		{
			block:               block9,
			expectedChainHeight: 3,
		},
		{
			block:               block10,
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

// This test is to ensure the size BlueAnticoneSizesSize is limited to uint8.
// We verify that by serializing and deserializing the block while we make sure that we stay within the expected range.
func TestBlueAnticoneSizesSize(t *testing.T) {
	dag, teardownFunc, err := DAGSetup("TestBlueAnticoneSizesSize", Config{
		DAGParams: &dagconfig.SimnetParams,
	})
	if err != nil {
		t.Fatalf("TestBlueAnticoneSizesSize: Failed to setup DAG instance: %s", err)
	}
	defer teardownFunc()
	blockHeader := dagconfig.SimnetParams.GenesisBlock.Header
	block, _ := dag.newBlockNode(&blockHeader, newSet())

	maxKsize := dagconfig.KType(math.MaxUint8)
	expected := maxKsize
	block.bluesAnticoneSizes[daghash.Hash{1}] = expected
	serializedNode, _ := serializeBlockNode(block)
	deserializedNode, _ := dag.deserializeBlockNode(serializedNode)
	if deserializedNode.bluesAnticoneSizes[daghash.Hash{1}] != expected {
		t.Fatalf("TestBlueAnticoneSizesSize: BlueAnticoneSize should not change when deserilizing. Expected: %v but got %v",
			expected, deserializedNode.bluesAnticoneSizes[daghash.Hash{1}])
	}

	block.bluesAnticoneSizes[daghash.Hash{1}]++
	size := block.bluesAnticoneSizes[daghash.Hash{1}]
	serializedNode, _ = serializeBlockNode(block)
	deserializedNode, _ = dag.deserializeBlockNode(serializedNode)
	if deserializedNode.bluesAnticoneSizes[daghash.Hash{1}] != size {
		t.Fatalf("TestBlueAnticoneSizesSize: BlueAnticoneSize should not be larger than MaxUint8. Expected: %v but got %v",
			deserializedNode.bluesAnticoneSizes[daghash.Hash{1}], size)
	}
}
