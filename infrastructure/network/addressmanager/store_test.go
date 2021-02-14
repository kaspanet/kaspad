package addressmanager

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util/mstime"
	"net"
	"reflect"
	"testing"
)

func TestAddressKeySerialization(t *testing.T) {
	addressStore := newAddressStore(nil)

	testAddress := &appmessage.NetAddress{IP: net.ParseIP("2602:100:abcd::102"), Port: 12345}
	testAddressKey := netAddressKey(testAddress)

	serializedTestAddressKey := addressStore.serializeAddressKey(testAddressKey)
	deserializedTestAddressKey := addressStore.deserializeAddressKey(serializedTestAddressKey)
	if !reflect.DeepEqual(testAddressKey, deserializedTestAddressKey) {
		t.Fatalf("testAddressKey and deserializedTestAddressKey are not equal\n"+
			"testAddressKey:%+v\ndeserializedTestAddressKey:%+v", testAddressKey, deserializedTestAddressKey)
	}
}

func TestNetAddressSerialization(t *testing.T) {
	addressStore := newAddressStore(nil)

	testAddress := &appmessage.NetAddress{
		IP:        net.ParseIP("2602:100:abcd::102"),
		Port:      12345,
		Timestamp: mstime.Now(),
		Services:  appmessage.ServiceFlag(6789),
	}

	serializedTestNetAddress := addressStore.serializeNetAddress(testAddress)
	deserializedTestNetAddress := addressStore.deserializeNetAddress(serializedTestNetAddress)
	if !reflect.DeepEqual(testAddress, deserializedTestNetAddress) {
		t.Fatalf("testAddress and deserializedTestNetAddress are not equal\n"+
			"testAddress:%+v\ndeserializedTestNetAddress:%+v", testAddress, deserializedTestNetAddress)
	}
}
