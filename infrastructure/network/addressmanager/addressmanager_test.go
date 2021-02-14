// Copyright (c) 2013-2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package addressmanager

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"
	"github.com/kaspanet/kaspad/util/mstime"
	"net"
	"reflect"
	"testing"

	"github.com/kaspanet/kaspad/infrastructure/config"
)

func newAddrManagerForTest(t *testing.T, testName string) (addressManager *AddressManager, teardown func()) {
	cfg := config.DefaultConfig()

	datadir := t.TempDir()
	database, err := ldb.NewLevelDB(datadir, 8)
	if err != nil {
		t.Fatalf("%s: could not create a database: %s", testName, err)
	}

	addressManager, err = New(NewConfig(cfg), database)
	if err != nil {
		t.Fatalf("%s: error creating address manager: %s", testName, err)
	}

	return addressManager, func() {
		database.Close()
	}
}

func TestBestLocalAddress(t *testing.T) {
	localAddrs := []appmessage.NetAddress{
		{IP: net.ParseIP("192.168.0.100")},
		{IP: net.ParseIP("::1")},
		{IP: net.ParseIP("fe80::1")},
		{IP: net.ParseIP("2001:470::1")},
	}

	var tests = []struct {
		remoteAddr appmessage.NetAddress
		want0      appmessage.NetAddress
		want1      appmessage.NetAddress
		want2      appmessage.NetAddress
		want3      appmessage.NetAddress
	}{
		{
			// Remote connection from public IPv4
			appmessage.NetAddress{IP: net.ParseIP("204.124.8.1")},
			appmessage.NetAddress{IP: net.IPv4zero},
			appmessage.NetAddress{IP: net.IPv4zero},
			appmessage.NetAddress{IP: net.ParseIP("204.124.8.100")},
			appmessage.NetAddress{IP: net.ParseIP("fd87:d87e:eb43:25::1")},
		},
		{
			// Remote connection from private IPv4
			appmessage.NetAddress{IP: net.ParseIP("172.16.0.254")},
			appmessage.NetAddress{IP: net.IPv4zero},
			appmessage.NetAddress{IP: net.IPv4zero},
			appmessage.NetAddress{IP: net.IPv4zero},
			appmessage.NetAddress{IP: net.IPv4zero},
		},
		{
			// Remote connection from public IPv6
			appmessage.NetAddress{IP: net.ParseIP("2602:100:abcd::102")},
			appmessage.NetAddress{IP: net.IPv6zero},
			appmessage.NetAddress{IP: net.ParseIP("2001:470::1")},
			appmessage.NetAddress{IP: net.ParseIP("2001:470::1")},
			appmessage.NetAddress{IP: net.ParseIP("2001:470::1")},
		},
	}

	amgr, teardown := newAddrManagerForTest(t, "TestGetBestLocalAddress")
	defer teardown()

	// Test against default when there's no address
	for x, test := range tests {
		got := amgr.BestLocalAddress(&test.remoteAddr)
		if !test.want0.IP.Equal(got.IP) {
			t.Errorf("TestGetBestLocalAddress test1 #%d failed for remote address %s: want %s got %s",
				x, test.remoteAddr.IP, test.want1.IP, got.IP)
			continue
		}
	}

	for _, localAddr := range localAddrs {
		amgr.localAddresses.addLocalNetAddress(&localAddr, InterfacePrio)
	}

	// Test against want1
	for x, test := range tests {
		got := amgr.BestLocalAddress(&test.remoteAddr)
		if !test.want1.IP.Equal(got.IP) {
			t.Errorf("TestGetBestLocalAddress test1 #%d failed for remote address %s: want %s got %s",
				x, test.remoteAddr.IP, test.want1.IP, got.IP)
			continue
		}
	}

	// Add a public IP to the list of local addresses.
	localAddr := appmessage.NetAddress{IP: net.ParseIP("204.124.8.100")}
	amgr.localAddresses.addLocalNetAddress(&localAddr, InterfacePrio)

	// Test against want2
	for x, test := range tests {
		got := amgr.BestLocalAddress(&test.remoteAddr)
		if !test.want2.IP.Equal(got.IP) {
			t.Errorf("TestGetBestLocalAddress test2 #%d failed for remote address %s: want %s got %s",
				x, test.remoteAddr.IP, test.want2.IP, got.IP)
			continue
		}
	}
}

func TestAddressManager(t *testing.T) {
	addressManager, teardown := newAddrManagerForTest(t, "TestAddressManager")
	defer teardown()

	testAddress1 := &appmessage.NetAddress{IP: net.ParseIP("1.2.3.4"), Timestamp: mstime.Now()}
	testAddress2 := &appmessage.NetAddress{IP: net.ParseIP("5.6.8.8"), Timestamp: mstime.Now()}
	testAddress3 := &appmessage.NetAddress{IP: net.ParseIP("9.0.1.2"), Timestamp: mstime.Now()}
	testAddresses := []*appmessage.NetAddress{testAddress1, testAddress2, testAddress3}

	// Add a few addresses
	err := addressManager.AddAddresses(testAddresses...)
	if err != nil {
		t.Fatalf("AddAddresses() failed: %s", err)
	}

	// Make sure that all the addresses are returned
	addresses := addressManager.Addresses()
	if len(testAddresses) != len(addresses) {
		t.Fatalf("Unexpected amount of addresses returned from Addresses. "+
			"Want: %d, got: %d", len(testAddresses), len(addresses))
	}
	for _, testAddress := range testAddresses {
		found := false
		for _, address := range addresses {
			if reflect.DeepEqual(testAddress, address) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Address %s not returned from Addresses().", testAddress.IP)
		}
	}

	// Remove an address
	addressToRemove := testAddress2
	err = addressManager.RemoveAddress(addressToRemove)
	if err != nil {
		t.Fatalf("RemoveAddress() failed: %s", err)
	}

	// Make sure that the removed address is not returned
	addresses = addressManager.Addresses()
	if len(addresses) != len(testAddresses)-1 {
		t.Fatalf("Unexpected amount of addresses returned from Addresses(). "+
			"Want: %d, got: %d", len(addresses), len(testAddresses)-1)
	}
	for _, address := range addresses {
		if reflect.DeepEqual(addressToRemove, address) {
			t.Fatalf("Removed addresses %s returned from Addresses()", addressToRemove.IP)
		}
	}

	// Add that address back
	err = addressManager.AddAddress(addressToRemove)
	if err != nil {
		t.Fatalf("AddAddress() failed: %s", err)
	}

	// Ban a different address
	addressToBan := testAddress3
	err = addressManager.Ban(addressToBan)
	if err != nil {
		t.Fatalf("Ban() failed: %s", err)
	}

	// Make sure that the banned address is not returned
	addresses = addressManager.Addresses()
	if len(addresses) != len(testAddresses)-1 {
		t.Fatalf("Unexpected amount of addresses returned from Addresses(). "+
			"Want: %d, got: %d", len(addresses), len(testAddresses)-1)
	}
	for _, address := range addresses {
		if reflect.DeepEqual(addressToBan, address) {
			t.Fatalf("Banned addresses %s returned from Addresses()", addressToBan.IP)
		}
	}

	// Check that the address is banned
	isBanned, err := addressManager.IsBanned(addressToBan)
	if err != nil {
		t.Fatalf("IsBanned() failed: %s", err)
	}
	if !isBanned {
		t.Fatalf("Adderss %s is unexpectedly not banned", addressToBan.IP)
	}

	// Check that BannedAddresses() returns the banned address
	bannedAddresses := addressManager.BannedAddresses()
	if len(bannedAddresses) != 1 {
		t.Fatalf("Unexpected amount of addresses returned from BannedAddresses(). "+
			"Want: %d, got: %d", 1, len(bannedAddresses))
	}
	if !reflect.DeepEqual(addressToBan, bannedAddresses[0]) {
		t.Fatalf("Banned address %s not returned from BannedAddresses()", addressToBan.IP)
	}

	// Unban the address
	err = addressManager.Unban(addressToBan)
	if err != nil {
		t.Fatalf("Unban() failed: %s", err)
	}

	// Check that BannedAddresses() not longer returns the banned address
	bannedAddresses = addressManager.BannedAddresses()
	if len(bannedAddresses) != 0 {
		t.Fatalf("Unexpected amount of addresses returned from BannedAddresses(). "+
			"Want: %d, got: %d", 0, len(bannedAddresses))
	}
}
