// Copyright (c) 2013-2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package addrmgr

import (
	"net"
	"testing"

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

func lookupFunc(host string) ([]net.IP, error) {
	return nil, errors.New("not implemented")
}

func TestStartStop(t *testing.T) {
	n := New("teststartstop", lookupFunc, nil)
	n.Start()
	err := n.Stop()
	if err != nil {
		t.Fatalf("Address Manager failed to stop: %v", err)
	}
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
