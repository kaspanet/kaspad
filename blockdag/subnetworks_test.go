package blockdag

import (
	"reflect"
	"testing"
)

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
