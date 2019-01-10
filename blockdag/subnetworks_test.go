package blockdag

import (
	"reflect"
	"testing"

	"github.com/daglabs/btcd/dagconfig"
)

// TestSubNetworkRegistry tests the full sub-network registry flow
func TestSubNetworkRegistry(t *testing.T) {
	params := dagconfig.SimNetParams
	params.K = 1
	dag, teardownFunc, err := DAGSetup("TestSubNetworkRegistry", Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	gasLimit := uint64(12345)
	subNetworkID, err := RegisterSubNetworkForTest(dag, gasLimit)
	if err != nil {
		t.Fatalf("could not register network: %s", err)
	}
	limit, err := dag.GasLimit(subNetworkID)
	if err != nil {
		t.Fatalf("could not retrieve gas limit: %s", err)
	}
	if limit != gasLimit {
		t.Fatalf("unexpected gas limit. want: %d, got: %d", gasLimit, limit)
	}
}

func TestSerializeSubNetwork(t *testing.T) {
	sNet := &subNetwork{
		gasLimit: 1000,
	}

	serializedSNet, err := serializeSubNetwork(sNet)
	if err != nil {
		t.Fatalf("sub-network serialization unexpectedly failed: %s", err)
	}

	deserializedSNet, err := deserializeSubNetwork(serializedSNet)
	if err != nil {
		t.Fatalf("sub-network deserialization unexpectedly failed: %s", err)
	}

	if !reflect.DeepEqual(sNet, deserializedSNet) {
		t.Errorf("original sub-network and deserialized sub-network are not equal")
	}
}
