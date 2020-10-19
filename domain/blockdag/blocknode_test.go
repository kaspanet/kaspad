package blockdag

import (
	"github.com/kaspanet/kaspad/util/mstime"
	"testing"

	"github.com/kaspanet/kaspad/domain/blocknode"

	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util/daghash"
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
	node, _ := dag.newBlockNode(&blockHeader, blocknode.NewSet())
	fakeBlue := &blocknode.Node{Hash: &daghash.Hash{1}}
	dag.Index.AddNode(fakeBlue)
	// Setting maxKType to maximum value of KType.
	// As we verify above that KType is unsigned we can be sure that maxKType is indeed the maximum value of KType.
	maxKType := ^dagconfig.KType(0)
	node.BluesAnticoneSizes[fakeBlue] = maxKType
	serializedNode, _ := blocknode.SerializeNode(node)
	deserializedNode, _ := dag.deserializeBlockNode(serializedNode)
	if deserializedNode.BluesAnticoneSizes[fakeBlue] != maxKType {
		t.Fatalf("TestBlueAnticoneSizesSize: BlueAnticoneSize should not change when deserializing. Expected: %v but got %v",
			maxKType, deserializedNode.BluesAnticoneSizes[fakeBlue])
	}
}

func TestAncestorErrors(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	params := dagconfig.SimnetParams
	dag, teardownFunc, err := DAGSetup("TestAncestorErrors", true, Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("TestAncestorErrors: Failed to setup DAG instance: %s", err)
	}
	defer teardownFunc()

	node := newTestNode(dag, blocknode.NewSet(), int32(0x10000000), 0, mstime.Now())
	node.BlueScore = 2
	ancestor := node.SelectedAncestor(3)
	if ancestor != nil {
		t.Errorf("TestAncestorErrors: Ancestor() unexpectedly returned a node. Expected: <nil>")
	}
}
