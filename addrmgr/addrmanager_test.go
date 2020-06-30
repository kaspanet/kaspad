// Copyright (c) 2013-2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package addrmgr

import (
	"fmt"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/dbaccess"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/kaspanet/kaspad/util/subnetworkid"
	"io/ioutil"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/wire"
)

// naTest is used to describe a test to be performed against the NetAddressKey
// method.
type naTest struct {
	in   wire.NetAddress
	want string
}

// naTests houses all of the tests to be performed against the NetAddressKey
// method.
var naTests = make([]naTest, 0)

// Put some IP in here for convenience. Points to google.
var someIP = "173.194.115.66"

// addNaTests
func addNaTests() {
	// IPv4
	// Localhost
	addNaTest("127.0.0.1", 16111, "127.0.0.1:16111")
	addNaTest("127.0.0.1", 16110, "127.0.0.1:16110")

	// Class A
	addNaTest("1.0.0.1", 16111, "1.0.0.1:16111")
	addNaTest("2.2.2.2", 16110, "2.2.2.2:16110")
	addNaTest("27.253.252.251", 8335, "27.253.252.251:8335")
	addNaTest("123.3.2.1", 8336, "123.3.2.1:8336")

	// Private Class A
	addNaTest("10.0.0.1", 16111, "10.0.0.1:16111")
	addNaTest("10.1.1.1", 16110, "10.1.1.1:16110")
	addNaTest("10.2.2.2", 8335, "10.2.2.2:8335")
	addNaTest("10.10.10.10", 8336, "10.10.10.10:8336")

	// Class B
	addNaTest("128.0.0.1", 16111, "128.0.0.1:16111")
	addNaTest("129.1.1.1", 16110, "129.1.1.1:16110")
	addNaTest("180.2.2.2", 8335, "180.2.2.2:8335")
	addNaTest("191.10.10.10", 8336, "191.10.10.10:8336")

	// Private Class B
	addNaTest("172.16.0.1", 16111, "172.16.0.1:16111")
	addNaTest("172.16.1.1", 16110, "172.16.1.1:16110")
	addNaTest("172.16.2.2", 8335, "172.16.2.2:8335")
	addNaTest("172.16.172.172", 8336, "172.16.172.172:8336")

	// Class C
	addNaTest("193.0.0.1", 16111, "193.0.0.1:16111")
	addNaTest("200.1.1.1", 16110, "200.1.1.1:16110")
	addNaTest("205.2.2.2", 8335, "205.2.2.2:8335")
	addNaTest("223.10.10.10", 8336, "223.10.10.10:8336")

	// Private Class C
	addNaTest("192.168.0.1", 16111, "192.168.0.1:16111")
	addNaTest("192.168.1.1", 16110, "192.168.1.1:16110")
	addNaTest("192.168.2.2", 8335, "192.168.2.2:8335")
	addNaTest("192.168.192.192", 8336, "192.168.192.192:8336")

	// IPv6
	// Localhost
	addNaTest("::1", 16111, "[::1]:16111")
	addNaTest("fe80::1", 16110, "[fe80::1]:16110")

	// Link-local
	addNaTest("fe80::1:1", 16111, "[fe80::1:1]:16111")
	addNaTest("fe91::2:2", 16110, "[fe91::2:2]:16110")
	addNaTest("fea2::3:3", 8335, "[fea2::3:3]:8335")
	addNaTest("feb3::4:4", 8336, "[feb3::4:4]:8336")

	// Site-local
	addNaTest("fec0::1:1", 16111, "[fec0::1:1]:16111")
	addNaTest("fed1::2:2", 16110, "[fed1::2:2]:16110")
	addNaTest("fee2::3:3", 8335, "[fee2::3:3]:8335")
	addNaTest("fef3::4:4", 8336, "[fef3::4:4]:8336")
}

func addNaTest(ip string, port uint16, want string) {
	nip := net.ParseIP(ip)
	na := *wire.NewNetAddressIPPort(nip, port, wire.SFNodeNetwork)
	test := naTest{na, want}
	naTests = append(naTests, test)
}

func lookupFuncForTest(host string) ([]net.IP, error) {
	return nil, errors.New("not implemented")
}

func newAddrManagerForTest(t *testing.T, testName string,
	localSubnetworkID *subnetworkid.SubnetworkID) (addressManager *AddrManager, teardown func()) {

	dbPath, err := ioutil.TempDir("", testName)
	if err != nil {
		t.Fatalf("Error creating temporary directory: %s", err)
	}

	err = dbaccess.Open(dbPath)
	if err != nil {
		t.Fatalf("error creating db: %s", err)
	}

	addressManager = New(lookupFuncForTest, localSubnetworkID)

	return addressManager, func() {
		err := dbaccess.Close()
		if err != nil {
			t.Fatalf("error closing the database: %s", err)
		}
	}
}

func TestStartStop(t *testing.T) {
	amgr, teardown := newAddrManagerForTest(t, "TestStartStop", nil)
	defer teardown()
	err := amgr.Start()
	if err != nil {
		t.Fatalf("Address Manager failed to start: %v", err)
	}
	err = amgr.Stop()
	if err != nil {
		t.Fatalf("Address Manager failed to stop: %v", err)
	}
}

func TestAddAddressByIP(t *testing.T) {
	originalActiveCfg := config.ActiveConfig()
	config.SetActiveConfig(&config.Config{
		Flags: &config.Flags{
			NetworkFlags: config.NetworkFlags{
				ActiveNetParams: &dagconfig.SimnetParams},
		},
	})
	defer config.SetActiveConfig(originalActiveCfg)

	fmtErr := errors.Errorf("")
	addrErr := &net.AddrError{}
	var tests = []struct {
		addrIP string
		err    error
	}{
		{
			someIP + ":16111",
			nil,
		},
		{
			someIP,
			addrErr,
		},
		{
			someIP[:12] + ":8333",
			fmtErr,
		},
		{
			someIP + ":abcd",
			fmtErr,
		},
	}

	amgr, teardown := newAddrManagerForTest(t, "TestAddAddressByIP", nil)
	defer teardown()
	for i, test := range tests {
		err := amgr.AddAddressByIP(test.addrIP, nil)
		if test.err != nil && err == nil {
			t.Errorf("TestAddAddressByIP test %d failed expected an error and got none", i)
			continue
		}
		if test.err == nil && err != nil {
			t.Errorf("TestAddAddressByIP test %d failed expected no error and got one", i)
			continue
		}
		if reflect.TypeOf(err) != reflect.TypeOf(test.err) {
			t.Errorf("TestAddAddressByIP test %d failed got %v, want %v", i,
				reflect.TypeOf(err), reflect.TypeOf(test.err))
			continue
		}
	}
}

func TestAddLocalAddress(t *testing.T) {
	originalActiveCfg := config.ActiveConfig()
	config.SetActiveConfig(&config.Config{
		Flags: &config.Flags{
			NetworkFlags: config.NetworkFlags{
				ActiveNetParams: &dagconfig.SimnetParams},
		},
	})
	defer config.SetActiveConfig(originalActiveCfg)

	var tests = []struct {
		address  wire.NetAddress
		priority AddressPriority
		valid    bool
	}{
		{
			wire.NetAddress{IP: net.ParseIP("192.168.0.100")},
			InterfacePrio,
			false,
		},
		{
			wire.NetAddress{IP: net.ParseIP("204.124.1.1")},
			InterfacePrio,
			true,
		},
		{
			wire.NetAddress{IP: net.ParseIP("204.124.1.1")},
			BoundPrio,
			true,
		},
		{
			wire.NetAddress{IP: net.ParseIP("::1")},
			InterfacePrio,
			false,
		},
		{
			wire.NetAddress{IP: net.ParseIP("fe80::1")},
			InterfacePrio,
			false,
		},
		{
			wire.NetAddress{IP: net.ParseIP("2620:100::1")},
			InterfacePrio,
			true,
		},
	}
	amgr, teardown := newAddrManagerForTest(t, "TestAddLocalAddress", nil)
	defer teardown()
	for x, test := range tests {
		result := amgr.AddLocalAddress(&test.address, test.priority)
		if result == nil && !test.valid {
			t.Errorf("TestAddLocalAddress test #%d failed: %s should have "+
				"been accepted", x, test.address.IP)
			continue
		}
		if result != nil && test.valid {
			t.Errorf("TestAddLocalAddress test #%d failed: %s should not have "+
				"been accepted", x, test.address.IP)
			continue
		}
	}
}

func TestAttempt(t *testing.T) {
	originalActiveCfg := config.ActiveConfig()
	config.SetActiveConfig(&config.Config{
		Flags: &config.Flags{
			NetworkFlags: config.NetworkFlags{
				ActiveNetParams: &dagconfig.SimnetParams},
		},
	})
	defer config.SetActiveConfig(originalActiveCfg)

	amgr, teardown := newAddrManagerForTest(t, "TestAttempt", nil)
	defer teardown()

	// Add a new address and get it
	err := amgr.AddAddressByIP(someIP+":8333", nil)
	if err != nil {
		t.Fatalf("Adding address failed: %v", err)
	}
	ka := amgr.GetAddress()

	if !ka.LastAttempt().IsZero() {
		t.Errorf("Address should not have attempts, but does")
	}

	na := ka.NetAddress()
	amgr.Attempt(na)

	if ka.LastAttempt().IsZero() {
		t.Errorf("Address should have an attempt, but does not")
	}
}

func TestConnected(t *testing.T) {
	originalActiveCfg := config.ActiveConfig()
	config.SetActiveConfig(&config.Config{
		Flags: &config.Flags{
			NetworkFlags: config.NetworkFlags{
				ActiveNetParams: &dagconfig.SimnetParams},
		},
	})
	defer config.SetActiveConfig(originalActiveCfg)

	amgr, teardown := newAddrManagerForTest(t, "TestConnected", nil)
	defer teardown()

	// Add a new address and get it
	err := amgr.AddAddressByIP(someIP+":8333", nil)
	if err != nil {
		t.Fatalf("Adding address failed: %v", err)
	}
	ka := amgr.GetAddress()
	na := ka.NetAddress()
	// make it an hour ago
	na.Timestamp = mstime.Now().Add(time.Hour * -1)

	amgr.Connected(na)

	if !ka.NetAddress().Timestamp.After(na.Timestamp) {
		t.Errorf("Address should have a new timestamp, but does not")
	}
}

func TestNeedMoreAddresses(t *testing.T) {
	originalActiveCfg := config.ActiveConfig()
	config.SetActiveConfig(&config.Config{
		Flags: &config.Flags{
			NetworkFlags: config.NetworkFlags{
				ActiveNetParams: &dagconfig.SimnetParams},
		},
	})
	defer config.SetActiveConfig(originalActiveCfg)

	amgr, teardown := newAddrManagerForTest(t, "TestNeedMoreAddresses", nil)
	defer teardown()
	addrsToAdd := 1500
	b := amgr.NeedMoreAddresses()
	if !b {
		t.Errorf("Expected that we need more addresses")
	}
	addrs := make([]*wire.NetAddress, addrsToAdd)

	var err error
	for i := 0; i < addrsToAdd; i++ {
		s := fmt.Sprintf("%d.%d.173.147:8333", i/128+60, i%128+60)
		addrs[i], err = amgr.DeserializeNetAddress(s)
		if err != nil {
			t.Errorf("Failed to turn %s into an address: %v", s, err)
		}
	}

	srcAddr := wire.NewNetAddressIPPort(net.IPv4(173, 144, 173, 111), 8333, 0)

	amgr.AddAddresses(addrs, srcAddr, nil)
	numAddrs := amgr.TotalNumAddresses()
	if numAddrs > addrsToAdd {
		t.Errorf("Number of addresses is too many %d vs %d", numAddrs, addrsToAdd)
	}

	b = amgr.NeedMoreAddresses()
	if b {
		t.Errorf("Expected that we don't need more addresses")
	}
}

func TestGood(t *testing.T) {
	originalActiveCfg := config.ActiveConfig()
	config.SetActiveConfig(&config.Config{
		Flags: &config.Flags{
			NetworkFlags: config.NetworkFlags{
				ActiveNetParams: &dagconfig.SimnetParams},
		},
	})
	defer config.SetActiveConfig(originalActiveCfg)

	amgr, teardown := newAddrManagerForTest(t, "TestGood", nil)
	defer teardown()
	addrsToAdd := 64 * 64
	addrs := make([]*wire.NetAddress, addrsToAdd)
	subnetworkCount := 32
	subnetworkIDs := make([]*subnetworkid.SubnetworkID, subnetworkCount)

	var err error
	for i := 0; i < addrsToAdd; i++ {
		s := fmt.Sprintf("%d.173.147.%d:8333", i/64+60, i%64+60)
		addrs[i], err = amgr.DeserializeNetAddress(s)
		if err != nil {
			t.Errorf("Failed to turn %s into an address: %v", s, err)
		}
	}

	for i := 0; i < subnetworkCount; i++ {
		subnetworkIDs[i] = &subnetworkid.SubnetworkID{0xff - byte(i)}
	}

	srcAddr := wire.NewNetAddressIPPort(net.IPv4(173, 144, 173, 111), 8333, 0)

	amgr.AddAddresses(addrs, srcAddr, nil)
	for i, addr := range addrs {
		amgr.Good(addr, subnetworkIDs[i%subnetworkCount])
	}

	numAddrs := amgr.TotalNumAddresses()
	if numAddrs >= addrsToAdd {
		t.Errorf("Number of addresses is too many: %d vs %d", numAddrs, addrsToAdd)
	}

	numCache := len(amgr.AddressCache(true, nil))
	if numCache == 0 || numCache >= numAddrs/4 {
		t.Errorf("Number of addresses in cache: got %d, want positive and less than %d",
			numCache, numAddrs/4)
	}

	for i := 0; i < subnetworkCount; i++ {
		numCache = len(amgr.AddressCache(false, subnetworkIDs[i]))
		if numCache == 0 || numCache >= numAddrs/subnetworkCount {
			t.Errorf("Number of addresses in subnetwork cache: got %d, want positive and less than %d",
				numCache, numAddrs/4/subnetworkCount)
		}
	}
}

func TestGoodChangeSubnetworkID(t *testing.T) {
	originalActiveCfg := config.ActiveConfig()
	config.SetActiveConfig(&config.Config{
		Flags: &config.Flags{
			NetworkFlags: config.NetworkFlags{
				ActiveNetParams: &dagconfig.SimnetParams},
		},
	})
	defer config.SetActiveConfig(originalActiveCfg)

	amgr, teardown := newAddrManagerForTest(t, "TestGoodChangeSubnetworkID", nil)
	defer teardown()
	addr := wire.NewNetAddressIPPort(net.IPv4(173, 144, 173, 111), 8333, 0)
	addrKey := NetAddressKey(addr)
	srcAddr := wire.NewNetAddressIPPort(net.IPv4(173, 144, 173, 111), 8333, 0)

	oldSubnetwork := subnetworkid.SubnetworkIDNative
	amgr.AddAddress(addr, srcAddr, oldSubnetwork)
	amgr.Good(addr, oldSubnetwork)

	// make sure address was saved to addrIndex under oldSubnetwork
	ka := amgr.find(addr)
	if ka == nil {
		t.Fatalf("Address was not found after first time .Good called")
	}
	if !ka.SubnetworkID().IsEqual(oldSubnetwork) {
		t.Fatalf("Address index did not point to oldSubnetwork")
	}

	// make sure address was added to correct bucket under oldSubnetwork
	bucket := amgr.addrTried[*oldSubnetwork][amgr.getTriedBucket(addr)]
	wasFound := false
	for e := bucket.Front(); e != nil; e = e.Next() {
		if NetAddressKey(e.Value.(*KnownAddress).NetAddress()) == addrKey {
			wasFound = true
		}
	}
	if !wasFound {
		t.Fatalf("Address was not found in the correct bucket in oldSubnetwork")
	}

	// now call .Good again with a different subnetwork
	newSubnetwork := subnetworkid.SubnetworkIDRegistry
	amgr.Good(addr, newSubnetwork)

	// make sure address was updated in addrIndex under newSubnetwork
	ka = amgr.find(addr)
	if ka == nil {
		t.Fatalf("Address was not found after second time .Good called")
	}
	if !ka.SubnetworkID().IsEqual(newSubnetwork) {
		t.Fatalf("Address index did not point to newSubnetwork")
	}

	// make sure address was removed from bucket under oldSubnetwork
	bucket = amgr.addrTried[*oldSubnetwork][amgr.getTriedBucket(addr)]
	wasFound = false
	for e := bucket.Front(); e != nil; e = e.Next() {
		if NetAddressKey(e.Value.(*KnownAddress).NetAddress()) == addrKey {
			wasFound = true
		}
	}
	if wasFound {
		t.Fatalf("Address was not removed from bucket in oldSubnetwork")
	}

	// make sure address was added to correct bucket under newSubnetwork
	bucket = amgr.addrTried[*newSubnetwork][amgr.getTriedBucket(addr)]
	wasFound = false
	for e := bucket.Front(); e != nil; e = e.Next() {
		if NetAddressKey(e.Value.(*KnownAddress).NetAddress()) == addrKey {
			wasFound = true
		}
	}
	if !wasFound {
		t.Fatalf("Address was not found in the correct bucket in newSubnetwork")
	}
}

func TestGetAddress(t *testing.T) {
	originalActiveCfg := config.ActiveConfig()
	config.SetActiveConfig(&config.Config{
		Flags: &config.Flags{
			NetworkFlags: config.NetworkFlags{
				ActiveNetParams: &dagconfig.SimnetParams},
		},
	})
	defer config.SetActiveConfig(originalActiveCfg)

	localSubnetworkID := &subnetworkid.SubnetworkID{0xff}
	amgr, teardown := newAddrManagerForTest(t, "TestGetAddress", localSubnetworkID)
	defer teardown()

	// Get an address from an empty set (should error)
	if rv := amgr.GetAddress(); rv != nil {
		t.Errorf("GetAddress failed: got: %v want: %v\n", rv, nil)
	}

	// Add a new address and get it
	err := amgr.AddAddressByIP(someIP+":8332", localSubnetworkID)
	if err != nil {
		t.Fatalf("Adding address failed: %v", err)
	}
	ka := amgr.GetAddress()
	if ka == nil {
		t.Fatalf("Did not get an address where there is one in the pool")
	}
	amgr.Attempt(ka.NetAddress())

	// Checks that we don't get it if we find that it has other subnetwork ID than expected.
	actualSubnetworkID := &subnetworkid.SubnetworkID{0xfe}
	amgr.Good(ka.NetAddress(), actualSubnetworkID)
	ka = amgr.GetAddress()
	if ka != nil {
		t.Errorf("Didn't expect to get an address because there shouldn't be any address from subnetwork ID %s or nil", localSubnetworkID)
	}

	// Checks that the total number of addresses incremented although the new address is not full node or a partial node of the same subnetwork as the local node.
	numAddrs := amgr.TotalNumAddresses()
	if numAddrs != 1 {
		t.Errorf("Wrong number of addresses: got %d, want %d", numAddrs, 1)
	}

	// Now we repeat the same process, but now the address has the expected subnetwork ID.

	// Add a new address and get it
	err = amgr.AddAddressByIP(someIP+":8333", localSubnetworkID)
	if err != nil {
		t.Fatalf("Adding address failed: %v", err)
	}
	ka = amgr.GetAddress()
	if ka == nil {
		t.Fatalf("Did not get an address where there is one in the pool")
	}
	if ka.NetAddress().IP.String() != someIP {
		t.Errorf("Wrong IP: got %v, want %v", ka.NetAddress().IP.String(), someIP)
	}
	if !ka.SubnetworkID().IsEqual(localSubnetworkID) {
		t.Errorf("Wrong Subnetwork ID: got %v, want %v", *ka.SubnetworkID(), localSubnetworkID)
	}
	amgr.Attempt(ka.NetAddress())

	// Mark this as a good address and get it
	amgr.Good(ka.NetAddress(), localSubnetworkID)
	ka = amgr.GetAddress()
	if ka == nil {
		t.Fatalf("Did not get an address where there is one in the pool")
	}
	if ka.NetAddress().IP.String() != someIP {
		t.Errorf("Wrong IP: got %v, want %v", ka.NetAddress().IP.String(), someIP)
	}
	if *ka.SubnetworkID() != *localSubnetworkID {
		t.Errorf("Wrong Subnetwork ID: got %v, want %v", ka.SubnetworkID(), localSubnetworkID)
	}

	numAddrs = amgr.TotalNumAddresses()
	if numAddrs != 2 {
		t.Errorf("Wrong number of addresses: got %d, want %d", numAddrs, 1)
	}
}

func TestGetBestLocalAddress(t *testing.T) {
	originalActiveCfg := config.ActiveConfig()
	config.SetActiveConfig(&config.Config{
		Flags: &config.Flags{
			NetworkFlags: config.NetworkFlags{
				ActiveNetParams: &dagconfig.SimnetParams},
		},
	})
	defer config.SetActiveConfig(originalActiveCfg)

	localAddrs := []wire.NetAddress{
		{IP: net.ParseIP("192.168.0.100")},
		{IP: net.ParseIP("::1")},
		{IP: net.ParseIP("fe80::1")},
		{IP: net.ParseIP("2001:470::1")},
	}

	var tests = []struct {
		remoteAddr wire.NetAddress
		want0      wire.NetAddress
		want1      wire.NetAddress
		want2      wire.NetAddress
		want3      wire.NetAddress
	}{
		{
			// Remote connection from public IPv4
			wire.NetAddress{IP: net.ParseIP("204.124.8.1")},
			wire.NetAddress{IP: net.IPv4zero},
			wire.NetAddress{IP: net.IPv4zero},
			wire.NetAddress{IP: net.ParseIP("204.124.8.100")},
			wire.NetAddress{IP: net.ParseIP("fd87:d87e:eb43:25::1")},
		},
		{
			// Remote connection from private IPv4
			wire.NetAddress{IP: net.ParseIP("172.16.0.254")},
			wire.NetAddress{IP: net.IPv4zero},
			wire.NetAddress{IP: net.IPv4zero},
			wire.NetAddress{IP: net.IPv4zero},
			wire.NetAddress{IP: net.IPv4zero},
		},
		{
			// Remote connection from public IPv6
			wire.NetAddress{IP: net.ParseIP("2602:100:abcd::102")},
			wire.NetAddress{IP: net.IPv6zero},
			wire.NetAddress{IP: net.ParseIP("2001:470::1")},
			wire.NetAddress{IP: net.ParseIP("2001:470::1")},
			wire.NetAddress{IP: net.ParseIP("2001:470::1")},
		},
		/* XXX
		{
			// Remote connection from Tor
			wire.NetAddress{IP: net.ParseIP("fd87:d87e:eb43::100")},
			wire.NetAddress{IP: net.IPv4zero},
			wire.NetAddress{IP: net.ParseIP("204.124.8.100")},
			wire.NetAddress{IP: net.ParseIP("fd87:d87e:eb43:25::1")},
		},
		*/
	}

	amgr, teardown := newAddrManagerForTest(t, "TestGetBestLocalAddress", nil)
	defer teardown()

	// Test against default when there's no address
	for x, test := range tests {
		got := amgr.GetBestLocalAddress(&test.remoteAddr)
		if !test.want0.IP.Equal(got.IP) {
			t.Errorf("TestGetBestLocalAddress test1 #%d failed for remote address %s: want %s got %s",
				x, test.remoteAddr.IP, test.want1.IP, got.IP)
			continue
		}
	}

	for _, localAddr := range localAddrs {
		amgr.AddLocalAddress(&localAddr, InterfacePrio)
	}

	// Test against want1
	for x, test := range tests {
		got := amgr.GetBestLocalAddress(&test.remoteAddr)
		if !test.want1.IP.Equal(got.IP) {
			t.Errorf("TestGetBestLocalAddress test1 #%d failed for remote address %s: want %s got %s",
				x, test.remoteAddr.IP, test.want1.IP, got.IP)
			continue
		}
	}

	// Add a public IP to the list of local addresses.
	localAddr := wire.NetAddress{IP: net.ParseIP("204.124.8.100")}
	amgr.AddLocalAddress(&localAddr, InterfacePrio)

	// Test against want2
	for x, test := range tests {
		got := amgr.GetBestLocalAddress(&test.remoteAddr)
		if !test.want2.IP.Equal(got.IP) {
			t.Errorf("TestGetBestLocalAddress test2 #%d failed for remote address %s: want %s got %s",
				x, test.remoteAddr.IP, test.want2.IP, got.IP)
			continue
		}
	}
	/*
		// Add a Tor generated IP address
		localAddr = wire.NetAddress{IP: net.ParseIP("fd87:d87e:eb43:25::1")}
		amgr.AddLocalAddress(&localAddr, ManualPrio)
		// Test against want3
		for x, test := range tests {
			got := amgr.GetBestLocalAddress(&test.remoteAddr)
			if !test.want3.IP.Equal(got.IP) {
				t.Errorf("TestGetBestLocalAddress test3 #%d failed for remote address %s: want %s got %s",
					x, test.remoteAddr.IP, test.want3.IP, got.IP)
				continue
			}
		}
	*/
}

func TestNetAddressKey(t *testing.T) {
	addNaTests()

	t.Logf("Running %d tests", len(naTests))
	for i, test := range naTests {
		key := NetAddressKey(&test.in)
		if key != test.want {
			t.Errorf("NetAddressKey #%d\n got: %s want: %s", i, key, test.want)
			continue
		}
	}

}
