package addressexchange

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kaspanet/kaspad/util/subnetworkid"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/flowcontext"
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/domain/mempool"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/connmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

func TestSendAddresses(t *testing.T) {
	tempDir := os.TempDir()
	dbPath := filepath.Join(tempDir, "TestNew")
	_ = os.RemoveAll(dbPath)
	databaseContext, err := dbaccess.New(dbPath)
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}

	defer func() {
		databaseContext.Close()
		os.RemoveAll(dbPath)
	}()

	dagCfg.DatabaseContext = databaseContext
	dag, err := blockdag.New(dagCfg)
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}

	defaultCfg.ActiveNetParams.AcceptUnroutable = true
	addressManager, err := addressmanager.New(defaultCfg, databaseContext)
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}
	err = addressManager.Start()
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}

	sourceAddress := generateAddressesForTest(1)[0]
	addresses := generateAddressesForTest(5)
	addressManager.AddAddresses(addresses, sourceAddress, subnetworkid.SubnetworkIDNative)

	memPoolCfg.DAG = dag
	txPool := mempool.New(memPoolCfg)

	netAdapter, err := netadapter.NewNetAdapter(defaultCfg)
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}

	connManager, err := connmanager.New(defaultCfg, netAdapter, addressManager)
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}

	ctx := flowcontext.New(defaultCfg, dag, addressManager, txPool, netAdapter, connManager)
	incomingRoute := router.NewRoute()
	outgoingRoute := router.NewRoute()
	err = incomingRoute.Enqueue(appmessage.NewMsgRequestAddresses(true, subnetworkid.SubnetworkIDNative))
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}

	err = SendAddresses(ctx, incomingRoute, outgoingRoute)
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}
}
