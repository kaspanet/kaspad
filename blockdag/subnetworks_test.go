package blockdag

import (
	"github.com/daglabs/btcd/wire"
	"reflect"
	"testing"

	"github.com/daglabs/btcd/dagconfig"
)

// TestSubnetworkRegistry tests the full subnetwork registry flow
func TestSubnetworkRegistry(t *testing.T) {
	params := dagconfig.SimNetParams
	params.K = 1
	dag, teardownFunc, err := DAGSetup("TestSubnetworkRegistry", Config{
		DAGParams:    &params,
		SubnetworkID: &wire.SubnetworkIDSupportsAll,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	gasLimit := uint64(12345)
	subnetworkID, err := RegisterSubnetworkForTest(dag, gasLimit)
	if err != nil {
		t.Fatalf("could not register network: %s", err)
	}
	limit, err := dag.SubnetworkStore.GasLimit(subnetworkID)
	if err != nil {
		t.Fatalf("could not retrieve gas limit: %s", err)
	}
	if limit != gasLimit {
		t.Fatalf("unexpected gas limit. want: %d, got: %d", gasLimit, limit)
	}
}

func TestSerializeSubnetwork(t *testing.T) {
	sNet := &subnetwork{
		gasLimit: 1000,
	}

	serializedSNet, err := serializeSubnetwork(sNet)
	if err != nil {
		t.Fatalf("subnetwork serialization unexpectedly failed: %s", err)
	}

	deserializedSNet, err := deserializeSubnetwork(serializedSNet)
	if err != nil {
		t.Fatalf("subnetwork deserialization unexpectedly failed: %s", err)
	}

	if !reflect.DeepEqual(sNet, deserializedSNet) {
		t.Errorf("original subnetwork and deserialized subnetwork are not equal")
	}
}
