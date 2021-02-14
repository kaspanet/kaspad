package addressmanager

import (
	"github.com/kaspanet/kaspad/app/appmessage"
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
