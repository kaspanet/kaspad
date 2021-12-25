package addressmanager

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util/mstime"
	"net"
	"reflect"
	"testing"
)

func TestAddressKeySerialization(t *testing.T) {
	addressManager, teardown := newAddressManagerForTest(t, "TestAddressKeySerialization")
	defer teardown()
	addressStore := addressManager.store

	testAddress := &appmessage.NetAddress{IP: net.ParseIP("2602:100:abcd::102"), Port: 12345}
	testAddressKey := netAddressKey(testAddress)

	serializedTestAddressKey := addressStore.serializeAddressKey(testAddressKey)
	deserializedTestAddressKey := addressStore.deserializeAddressKey(serializedTestAddressKey)
	if !reflect.DeepEqual(testAddressKey, deserializedTestAddressKey) {
		t.Fatalf("testAddressKey and deserializedTestAddressKey are not equal\n"+
			"testAddressKey:%+v\ndeserializedTestAddressKey:%+v", testAddressKey, deserializedTestAddressKey)
	}
}

func TestAddressSerialization(t *testing.T) {
	addressManager, teardown := newAddressManagerForTest(t, "TestAddressSerialization")
	defer teardown()
	addressStore := addressManager.store

	testAddress := &address{
		netAddress: &appmessage.NetAddress{
			IP:        net.ParseIP("2602:100:abcd::102"),
			Port:      12345,
			Timestamp: mstime.Now(),
		},
		level: level3,
	}

	serializedTestAddress := addressStore.serializeAddress(testAddress)
	deserializedTestAddress := addressStore.deserializeAddress(serializedTestAddress)
	if !reflect.DeepEqual(testAddress, deserializedTestAddress) {
		t.Fatalf("testAddress and deserializedTestAddress are not equal\n"+
			"testAddress:%+v\ndeserializedTestAddress:%+v", testAddress, deserializedTestAddress)
	}
}
