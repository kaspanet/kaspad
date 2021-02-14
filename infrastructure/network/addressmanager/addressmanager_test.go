// Copyright (c) 2013-2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package addressmanager

import (
	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"
	"io/ioutil"
	"net"
	"testing"

	"github.com/kaspanet/kaspad/app/appmessage"

	"github.com/kaspanet/kaspad/infrastructure/config"
)

func newAddrManagerForTest(t *testing.T, testName string) (addressManager *AddressManager, teardown func()) {
	cfg := config.DefaultConfig()

	datadir, err := ioutil.TempDir("", testName)
	if err != nil {
		t.Fatalf("%s: could not create a temp directory: %s", testName, err)
	}
	database, err := ldb.NewLevelDB(datadir, 8)
	if err != nil {
		t.Fatalf("%s: could not create a database: %s", testName, err)
	}

	addressManager, err = New(NewConfig(cfg), database)
	if err != nil {
		t.Fatalf("%s: error creating address manager: %s", testName, err)
	}

	return addressManager, func() {}
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
