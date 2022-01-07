package integration

import (
	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"
	"testing"
)

func TestAddressExchange(t *testing.T) {
	appHarness1, appHarness2, appHarness3, teardown := standardSetup(t)
	defer teardown()

	testAddress := "1.2.3.4:6789"
	err := addressmanager.AddAddressByIP(appHarness1.app.AddressManager(), testAddress, nil)
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

	t.Errorf("Didn't find testAddress in list of addresses of appHarness3")
}

func TestAddressExchangeV3V4(t *testing.T) {
	harnesses, teardown := setupHarnesses(t, []*harnessParams{
		{
			p2pAddress:              p2pAddress1,
			rpcAddress:              rpcAddress1,
			miningAddress:           miningAddress1,
			miningAddressPrivateKey: miningAddress1PrivateKey,
		},
		{
			p2pAddress:              p2pAddress2,
			rpcAddress:              rpcAddress2,
			miningAddress:           miningAddress2,
			miningAddressPrivateKey: miningAddress2PrivateKey,
		}, {
			p2pAddress:              p2pAddress3,
			rpcAddress:              rpcAddress3,
			miningAddress:           miningAddress3,
			miningAddressPrivateKey: miningAddress3PrivateKey,
			protocolVersion:         3,
		},
	})
	defer teardown()

	appHarness1, appHarness2, appHarness3 := harnesses[0], harnesses[1], harnesses[2]

	testAddress := "1.2.3.4:6789"
	err := addressmanager.AddAddressByIP(appHarness1.app.AddressManager(), testAddress, nil)
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

	t.Errorf("Didn't find testAddress in list of addresses of appHarness3")
}
