package integration

import (
	"testing"
)

func TestAddressExchange(t *testing.T) {
	appHarness1, appHarness2, appHarness3, teardown := standardSetup(t)
	defer teardown()

	testAddress := "1.2.3.4:6789"
	err := appHarness1.app.AddressManager().AddAddressByIP(testAddress, nil)
	if err != nil {
		t.Fatalf("Error adding address to addressManager: %+v", err)
	}

	connect(t, appHarness1, appHarness2)
	connect(t, appHarness2, appHarness3)

	peerAddresses, err := appHarness3.rpcClient.GetPeerAddresses()
	if err != nil {
		t.Fatalf("Error getting peer addresses: %+v", err)
	}

	for _, peerAddress := range peerAddresses.Addresses {
		if peerAddress.Addr == testAddress {
			return
		}
	}

	t.Errorf("Didn't testAddress in list of addresses of appHarness3")
}
