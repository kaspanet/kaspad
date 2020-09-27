package integration

import (
	"testing"

	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"
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
}
