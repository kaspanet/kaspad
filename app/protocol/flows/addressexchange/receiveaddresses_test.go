package addressexchange

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/flowcontext"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/domain/mempool"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/connmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util/subnetworkid"
)

func TestReceiveAddresses(t *testing.T) {
	const (
		host  = "127.0.0.1"
		portA = 3000
		portB = 3001
	)

	addressA := fmt.Sprintf("%s:%d", host, portA)
	addressB := fmt.Sprintf("%s:%d", host, portB)
	cfgA, cfgB := config.DefaultConfig(), config.DefaultConfig()
	cfgA.Listeners = []string{addressA}
	cfgB.Listeners = []string{addressB}

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

	netAdapterA, err := netadapter.NewNetAdapter(cfgA)
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}

	netAdapterA.SetRouterInitializer(func(router *router.Router, connection *netadapter.NetConnection) {})
	err = netAdapterA.Start()
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}

	// netAdapterB is needed to have a connection with netAdapterA
	netAdapterB, err := netadapter.NewNetAdapter(cfgB)
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}

	netAdapterB.SetRouterInitializer(func(router *router.Router, connection *netadapter.NetConnection) {})
	err = netAdapterB.Start()
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}

	err = netAdapterA.Connect(addressB)
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}

	connManager, err := connmanager.New(defaultCfg, netAdapterA, addressManager)
	if err != nil {
		t.Fatalf("SendAddresses: %s", err)
	}

	peer := peerpkg.New(netAdapterA.Connections()[0])

	ctx := flowcontext.New(defaultCfg, dag, addressManager, txPool, netAdapterA, connManager)
	incomingRoute := router.NewRoute()
	outgoingRoute := router.NewRoute()
	err = incomingRoute.Enqueue(appmessage.NewMsgAddresses(false, nil))
	if err != nil {
		t.Fatalf("ReceiveAddresses: %s", err)
	}

	err = ReceiveAddresses(ctx, incomingRoute, outgoingRoute, peer)
	if err != nil {
		t.Fatalf("ReceiveAddresses: %s", err)
	}
}
