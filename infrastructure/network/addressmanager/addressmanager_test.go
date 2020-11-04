// Copyright (c) 2013-2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package addressmanager

import (
	"fmt"
	"io/ioutil"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"

	"github.com/kaspanet/kaspad/app/appmessage"

	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/util/mstime"

	"github.com/pkg/errors"
)

// naTest is used to describe a test to be performed against the NetAddressKey
// method.
type naTest struct {
	in   appmessage.NetAddress
	want AddressKey
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

func addNaTest(ip string, port uint16, want AddressKey) {
	nip := net.ParseIP(ip)
	na := *appmessage.NewNetAddressIPPort(nip, port, appmessage.SFNodeNetwork)
	test := naTest{na, want}
	naTests = append(naTests, test)
}

func lookupFuncForTest(host string) ([]net.IP, error) {
	return nil, errors.New("not implemented")
}

func newAddrManagerForTest(t *testing.T, testName string,
	localSubnetworkID *externalapi.DomainSubnetworkID) (addressManager *AddressManager, teardown func()) {

	cfg := config.DefaultConfig()
	cfg.SubnetworkID = localSubnetworkID

	dbPath, err := ioutil.TempDir("", testName)
	if err != nil {
		t.Fatalf("Error creating temporary directory: %s", err)
	}

	databaseContext, err := ldb.NewLevelDB(dbPath)
	if err != nil {
		t.Fatalf("error creating db: %s", err)
	}

	addressManager, err = New(cfg, databaseContext)
	if err != nil {
		t.Fatalf("error creating address manager: %s", err)
	}

	return addressManager, func() {
		err := databaseContext.Close()
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
		err := AddAddressByIP(amgr, test.addrIP, nil)
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
	var tests = []struct {
		address  appmessage.NetAddress
		priority AddressPriority
		valid    bool
	}{
		{
			appmessage.NetAddress{IP: net.ParseIP("192.168.0.100")},
			InterfacePrio,
			false,
		},
		{
			appmessage.NetAddress{IP: net.ParseIP("204.124.1.1")},
			InterfacePrio,
			true,
		},
		{
			appmessage.NetAddress{IP: net.ParseIP("204.124.1.1")},
			BoundPrio,
			true,
		},
		{
			appmessage.NetAddress{IP: net.ParseIP("::1")},
			InterfacePrio,
			false,
		},
		{
			appmessage.NetAddress{IP: net.ParseIP("fe80::1")},
			InterfacePrio,
			false,
		},
		{
			appmessage.NetAddress{IP: net.ParseIP("2620:100::1")},
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
	amgr, teardown := newAddrManagerForTest(t, "TestAttempt", nil)
	defer teardown()

	// Add a new address and get it
	err := AddAddressByIP(amgr, someIP+":8333", nil)
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
	amgr, teardown := newAddrManagerForTest(t, "TestConnected", nil)
	defer teardown()

	// Add a new address and get it
	err := AddAddressByIP(amgr, someIP+":8333", nil)
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
	amgr, teardown := newAddrManagerForTest(t, "TestNeedMoreAddresses", nil)
	defer teardown()
	addrsToAdd := 1500
	b := amgr.NeedMoreAddresses()
	if !b {
		t.Errorf("Expected that we need more addresses")
	}
	addrs := make([]*appmessage.NetAddress, addrsToAdd)

	var err error
	for i := 0; i < addrsToAdd; i++ {
		s := AddressKey(fmt.Sprintf("%d.%d.173.147:8333", i/128+60, i%128+60))
		addrs[i], err = amgr.DeserializeNetAddress(s)
		if err != nil {
			t.Errorf("Failed to turn %s into an address: %v", s, err)
		}
	}

	srcAddr := appmessage.NewNetAddressIPPort(net.IPv4(173, 144, 173, 111), 8333, 0)

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
	amgr, teardown := newAddrManagerForTest(t, "TestGood", nil)
	defer teardown()
	addrsToAdd := 64 * 64
	addrs := make([]*appmessage.NetAddress, addrsToAdd)
	subnetworkCount := 32
	subnetworkIDs := make([]*externalapi.DomainSubnetworkID, subnetworkCount)

	var err error
	for i := 0; i < addrsToAdd; i++ {
		s := AddressKey(fmt.Sprintf("%d.173.147.%d:8333", i/64+60, i%64+60))
		addrs[i], err = amgr.DeserializeNetAddress(s)
		if err != nil {
			t.Errorf("Failed to turn %s into an address: %v", s, err)
		}
	}

	for i := 0; i < subnetworkCount; i++ {
		subnetworkIDs[i] = &externalapi.DomainSubnetworkID{0xff - byte(i)}
	}

	srcAddr := appmessage.NewNetAddressIPPort(net.IPv4(173, 144, 173, 111), 8333, 0)

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
	amgr, teardown := newAddrManagerForTest(t, "TestGoodChangeSubnetworkID", nil)
	defer teardown()
	addr := appmessage.NewNetAddressIPPort(net.IPv4(173, 144, 173, 111), 8333, 0)
	addrKey := NetAddressKey(addr)
	srcAddr := appmessage.NewNetAddressIPPort(net.IPv4(173, 144, 173, 111), 8333, 0)

	oldSubnetwork := &subnetworks.SubnetworkIDNative
	amgr.AddAddress(addr, srcAddr, oldSubnetwork)
	amgr.Good(addr, oldSubnetwork)

	// make sure address was saved to addressIndex under oldSubnetwork
	ka := amgr.knownAddress(addr)
	if ka == nil {
		t.Fatalf("Address was not found after first time .Good called")
	}
	if *ka.SubnetworkID() != *oldSubnetwork {
		t.Fatalf("Address index did not point to oldSubnetwork")
	}

	// make sure address was added to correct bucket under oldSubnetwork
	bucket := amgr.subnetworkTriedAddresBucketArrays[*oldSubnetwork][amgr.triedAddressBucketIndex(addr)]
	wasFound := false
	for _, ka := range bucket {
		if NetAddressKey(ka.NetAddress()) == addrKey {
			wasFound = true
		}
	}
	if !wasFound {
		t.Fatalf("Address was not found in the correct bucket in oldSubnetwork")
	}

	// now call .Good again with a different subnetwork
	newSubnetwork := &subnetworks.SubnetworkIDRegistry
	amgr.Good(addr, newSubnetwork)

	// make sure address was updated in addressIndex under newSubnetwork
	ka = amgr.knownAddress(addr)
	if ka == nil {
		t.Fatalf("Address was not found after second time .Good called")
	}
	if *ka.SubnetworkID() != *newSubnetwork {
		t.Fatalf("Address index did not point to newSubnetwork")
	}

	// make sure address was removed from bucket under oldSubnetwork
	bucket = amgr.subnetworkTriedAddresBucketArrays[*oldSubnetwork][amgr.triedAddressBucketIndex(addr)]
	wasFound = false
	for _, ka := range bucket {
		if NetAddressKey(ka.NetAddress()) == addrKey {
			wasFound = true
		}
	}
	if wasFound {
		t.Fatalf("Address was not removed from bucket in oldSubnetwork")
	}

	// make sure address was added to correct bucket under newSubnetwork
	bucket = amgr.subnetworkTriedAddresBucketArrays[*newSubnetwork][amgr.triedAddressBucketIndex(addr)]
	wasFound = false
	for _, ka := range bucket {
		if NetAddressKey(ka.NetAddress()) == addrKey {
			wasFound = true
		}
	}
	if !wasFound {
		t.Fatalf("Address was not found in the correct bucket in newSubnetwork")
	}
}

func TestGetAddress(t *testing.T) {
	localSubnetworkID := &externalapi.DomainSubnetworkID{0xff}
	amgr, teardown := newAddrManagerForTest(t, "TestGetAddress", localSubnetworkID)
	defer teardown()

	// Get an address from an empty set (should error)
	if rv := amgr.GetAddress(); rv != nil {
		t.Errorf("GetAddress failed: got: %v want: %v\n", rv, nil)
	}

	// Add a new address and get it
	err := AddAddressByIP(amgr, someIP+":8332", localSubnetworkID)
	if err != nil {
		t.Fatalf("Adding address failed: %v", err)
	}
	ka := amgr.GetAddress()
	if ka == nil {
		t.Fatalf("Did not get an address where there is one in the pool")
	}
	amgr.Attempt(ka.NetAddress())

	// Checks that we don't get it if we find that it has other subnetwork ID than expected.
	actualSubnetworkID := &externalapi.DomainSubnetworkID{0xfe}
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
	err = AddAddressByIP(amgr, someIP+":8333", localSubnetworkID)
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
	if *ka.SubnetworkID() != *localSubnetworkID {
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
	localAddr := appmessage.NetAddress{IP: net.ParseIP("204.124.8.100")}
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
		localAddr = appmessage.NetAddress{IP: net.ParseIP("fd87:d87e:eb43:25::1")}
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
