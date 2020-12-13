package integration

import (
	"testing"
)

func TestUTXOIndex(t *testing.T) {
	harnessParams := &harnessParams{
		p2pAddress:              p2pAddress1,
		rpcAddress:              rpcAddress1,
		miningAddress:           miningAddress1,
		miningAddressPrivateKey: miningAddress1PrivateKey,
		utxoIndex:               true,
	}
	kaspad, teardown := setupHarness(t, harnessParams)
	defer teardown()

	mineNextBlock(t, kaspad)
}
