package blockdag

import (
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/util/daghash"
	"testing"
)

func TestKTypeIsUnsigned(t *testing.T) {
	k := dagconfig.KType(0)
	k--

	if k < dagconfig.KType(0) {
		t.Fatalf("KType must be unsigned")
	}
}

// This test is to ensure the size BlueAnticoneSizesSize is serialized to the size of KType.
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
	node, _ := dag.newBlockNode(&blockHeader, newSet())
	hash := daghash.Hash{1}

	// Setting maxKType to maximum value fo KType
	maxKType := ^dagconfig.KType(0)
	node.bluesAnticoneSizes[hash] = maxKType
	serializedNode, _ := serializeBlockNode(node)
	deserializedNode, _ := dag.deserializeBlockNode(serializedNode)
	if deserializedNode.bluesAnticoneSizes[hash] != maxKType {
		t.Fatalf("TestBlueAnticoneSizesSize: BlueAnticoneSize should not change when deserilizing. Expected: %v but got %v",
			maxKType, deserializedNode.bluesAnticoneSizes[hash])
	}

	// Increasing bluesAnticoneSizes by 1 to make it overflow
	node.bluesAnticoneSizes[hash]++
	expected := node.bluesAnticoneSizes[hash]
	serializedNode, _ = serializeBlockNode(node)
	deserializedNode, _ = dag.deserializeBlockNode(serializedNode)
	if deserializedNode.bluesAnticoneSizes[hash] != expected {
		t.Fatalf("TestBlueAnticoneSizesSize: BlueAnticoneSize should not be larger than MaxKType. Expected: %v but got %v",
			deserializedNode.bluesAnticoneSizes[hash], expected)
	}

}
