package blockdag

import (
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/util/daghash"
	"testing"
)

// This test is to ensure the size BlueAnticoneSizesSize is serialized to the size of KType.
// We verify that by serializing and deserializing the block while making sure that we stay within the expected range.
func TestBlueAnticoneSizesSize(t *testing.T) {
	dag, teardownFunc, err := DAGSetup("TestBlueAnticoneSizesSize", true, Config{
		DAGParams: &dagconfig.SimnetParams,
	})
	if err != nil {
		t.Fatalf("TestBlueAnticoneSizesSize: Failed to setup DAG instance: %s", err)
	}
	defer teardownFunc()

	k := dagconfig.KType(0)
	k--

	if k < dagconfig.KType(0) {
		t.Fatalf("KType must be unsigned")
	}

	blockHeader := dagconfig.SimnetParams.GenesisBlock.Header
	node, _ := dag.newBlockNode(&blockHeader, newBlockSet())
	fakeBlue := &blockNode{hash: &daghash.Hash{1}}
	dag.index.AddNode(fakeBlue)
	// Setting maxKType to maximum value of KType.
	// As we verify above that KType is unsigned we can be sure that maxKType is indeed the maximum value of KType.
	maxKType := ^dagconfig.KType(0)
	node.bluesAnticoneSizes[fakeBlue] = maxKType
	serializedNode, _ := serializeBlockNode(node)
	deserializedNode, _ := dag.deserializeBlockNode(serializedNode)
	if deserializedNode.bluesAnticoneSizes[fakeBlue] != maxKType {
		t.Fatalf("TestBlueAnticoneSizesSize: BlueAnticoneSize should not change when deserializing. Expected: %v but got %v",
			maxKType, deserializedNode.bluesAnticoneSizes[fakeBlue])
	}
}
