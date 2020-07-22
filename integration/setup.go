package integration

import (
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/netadapter/server"

	"github.com/kaspanet/kaspad/rpc/client"
	rpcclient "github.com/kaspanet/kaspad/rpc/client"

	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/dbaccess"
	kaspadpkg "github.com/kaspanet/kaspad/kaspad"
)

func setup(t *testing.T) (kaspad1, kaspad2 *kaspadpkg.Kaspad, client1, client2 *client.Client, teardownFunc func()) {
	kaspad1Config, kaspad2Config := configs(t)

	kaspad1DatabaseContext, kaspad2DatabaseContext := openDBs(t, kaspad1Config, kaspad2Config)

	kaspad1, kaspad2 = NewKaspads(t,
		kaspad1Config, kaspad2Config,
		kaspad1DatabaseContext, kaspad2DatabaseContext)

	StartAndWaitForConnection(t, kaspad1, kaspad2)

	client1, client2 = rpcClients(t)

	verifyConnected(t, client1)
	verifyConnected(t, client2)

	return kaspad1, kaspad2, client1, client2,
		func() { teardown(t, kaspad1DatabaseContext, kaspad2DatabaseContext, kaspad1, kaspad2) }
}

func rpcClients(t *testing.T) (client1, client2 *rpcclient.Client) {
	client1, err := rpcClient(kaspad1RPCAddress)
	if err != nil {
		t.Fatalf("Error getting RPC client for kaspad1 %+v", err)
	}
	client2, err = rpcClient(kaspad2RPCAddress)
	if err != nil {
		t.Fatalf("Error getting RPC client for kaspad2: %+v", err)
	}
	return client1, client2
}

func StartAndWaitForConnection(t *testing.T, kaspad1, kaspad2 *kaspadpkg.Kaspad) {
	kaspad1OnConnectedChan := make(chan struct{})
	kaspad1OldOnConnectedHandler := kaspad1.NetAdapter.Server.OnConnectedHandler()
	kaspad1.NetAdapter.Server.SetOnConnectedHandler(func(connection server.Connection) error {
		err := kaspad1OldOnConnectedHandler(connection)
		close(kaspad1OnConnectedChan)
		return err
	})

	kaspad1.Start()
	kaspad2.Start()

	onConnectedWG := sync.WaitGroup{}
	onConnectedWG.Add(2)

	select {
	case <-kaspad1OnConnectedChan:
		kaspad1.NetAdapter.Server.SetOnConnectedHandler(kaspad1OldOnConnectedHandler)
	case <-time.After(10 * time.Second):
		t.Fatalf("Timed out waiting for the kaspads to connect")
	}
}

func teardown(t *testing.T,
	kaspad1DatabaseContext, kaspad2DatabaseContext *dbaccess.DatabaseContext,
	kaspad1, kaspad2 *kaspadpkg.Kaspad) {

	err := kaspad1DatabaseContext.Close()
	if err != nil {
		t.Errorf("Error closing kaspad1DatabaseContext: %+v", err)
	}
	err = kaspad2DatabaseContext.Close()
	if err != nil {
		t.Errorf("Error closing kaspad2DatabaseContext: %+v", err)
	}

	err = kaspad1.Stop()
	if err != nil {
		t.Errorf("Error stopping kaspad1 %+v", err)
	}
	err = kaspad2.Stop()
	if err != nil {
		t.Errorf("Error stopping kaspad2: %+v", err)
	}

	kaspad1.WaitForShutdown()
	kaspad2.WaitForShutdown()
}

func NewKaspads(t *testing.T,
	kaspad1Config, kaspad2Config *config.Config,
	kaspad1DatabaseContext, kaspad2DatabaseContext *dbaccess.DatabaseContext,
) (kaspad1, kaspad2 *kaspadpkg.Kaspad) {

	kaspad1, err := kaspadpkg.New(kaspad1Config, kaspad1DatabaseContext, make(chan struct{}))
	if err != nil {
		t.Fatalf("Error creating kaspad1: %+v", err)
	}

	kaspad2, err = kaspadpkg.New(kaspad2Config, kaspad2DatabaseContext, make(chan struct{}))
	if err != nil {
		t.Fatalf("Error creating kaspad2: %+v", err)
	}
	return kaspad1, kaspad2
}

func openDBs(t *testing.T, kaspad1Config *config.Config, kaspad2Config *config.Config) (
	kaspad1DatabaseContext *dbaccess.DatabaseContext,
	kaspad2DatabaseContext *dbaccess.DatabaseContext) {

	kaspad1DatabaseContext, err := openDB(kaspad1Config)
	if err != nil {
		t.Fatalf("Error openning database for kaspad1: %+v", err)
	}

	kaspad2DatabaseContext, err = openDB(kaspad2Config)
	if err != nil {
		t.Fatalf("Error openning database for kaspad2: %+v", err)
	}
	return kaspad1DatabaseContext, kaspad2DatabaseContext
}

func rpcClient(rpcAddress string) (*client.Client, error) {
	connConfig := &client.ConnConfig{
		Host:           rpcAddress,
		Endpoint:       "ws",
		User:           rpcUser,
		Pass:           rpcPass,
		DisableTLS:     true,
		RequestTimeout: time.Second * 10,
	}

	return client.New(connConfig, nil)
}

func openDB(cfg *config.Config) (*dbaccess.DatabaseContext, error) {
	dbPath := filepath.Join(cfg.DataDir, "db")
	return dbaccess.New(dbPath)
}

func verifyConnected(t *testing.T, client *rpcclient.Client) {
	connectedPeerInfo, err := client.GetConnectedPeerInfo()
	if err != nil {
		t.Fatalf("Error getting connected peer info from kaspad1")
	}
	if len(connectedPeerInfo) != 1 {
		t.Errorf("Expected 1 connected peer, but got %d", len(connectedPeerInfo))
	}
}
