package integration

import (
	"path/filepath"
	"testing"

	"github.com/kaspanet/kaspad/app"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/dbaccess"
)

func setup(t *testing.T) (app1, app2, app3 *app.App, client1, client2, client3 *rpcClient, teardownFunc func()) {
	config1, config2, config3 := configs(t)

	databaseContext1, databaseContext2, databaseContext3 := openDBs(t, config1, config2, config3)

	app1, app2, app3 = newApps(t, config1, config2, config3, databaseContext1, databaseContext2, databaseContext3)

	app1.Start()
	app2.Start()
	app3.Start()

	client1, client2, client3 = rpcClients(t)

	return app1, app2, app3, client1, client2, client3,
		func() { teardown(t, app1, app2, app3, databaseContext1, databaseContext2, databaseContext3) }
}

func rpcClients(t *testing.T) (client1, client2, client3 *rpcClient) {
	client1, err := newRPCClient(rpcAddress1)
	if err != nil {
		t.Fatalf("Error getting RPC client for app1 %+v", err)
	}

	client2, err = newRPCClient(rpcAddress2)
	if err != nil {
		t.Fatalf("Error getting RPC client for app2: %+v", err)
	}

	client3, err = newRPCClient(rpcAddress3)
	if err != nil {
		t.Fatalf("Error getting RPC client for app3: %+v", err)
	}

	return client1, client2, client3
}

func teardown(t *testing.T, app1, app2, app3 *app.App,
	databaseContext1, databaseContext2, databaseContext3 *dbaccess.DatabaseContext) {

	err := app1.Stop()
	if err != nil {
		t.Errorf("Error stopping app1 %+v", err)
	}
	err = app2.Stop()
	if err != nil {
		t.Errorf("Error stopping app2: %+v", err)
	}
	err = app3.Stop()
	if err != nil {
		t.Errorf("Error stopping app3: %+v", err)
	}

	app1.WaitForShutdown()
	app2.WaitForShutdown()
	app3.WaitForShutdown()

	err = databaseContext1.Close()
	if err != nil {
		t.Errorf("Error closing databaseContext1: %+v", err)
	}
	err = databaseContext2.Close()
	if err != nil {
		t.Errorf("Error closing databaseContext2: %+v", err)
	}
	err = databaseContext3.Close()
	if err != nil {
		t.Errorf("Error closing databaseContext3: %+v", err)
	}
}

func newApps(t *testing.T, config1, config2, config3 *config.Config,
	databaseContext1, databaseContext2, databaseContext3 *dbaccess.DatabaseContext,
) (app1, app2, app3 *app.App) {

	app1, err := app.New(config1, databaseContext1, make(chan struct{}))
	if err != nil {
		t.Fatalf("Error creating app1: %+v", err)
	}

	app2, err = app.New(config2, databaseContext2, make(chan struct{}))
	if err != nil {
		t.Fatalf("Error creating app2: %+v", err)
	}

	app3, err = app.New(config3, databaseContext3, make(chan struct{}))
	if err != nil {
		t.Fatalf("Error creating app3: %+v", err)
	}

	return app1, app2, app3
}

func openDBs(t *testing.T, config1, config2, config3 *config.Config) (
	databaseContext1 *dbaccess.DatabaseContext,
	databaseContext2 *dbaccess.DatabaseContext,
	databaseContext3 *dbaccess.DatabaseContext) {

	databaseContext1, err := openDB(config1)
	if err != nil {
		t.Fatalf("Error openning database for app1: %+v", err)
	}

	databaseContext2, err = openDB(config2)
	if err != nil {
		t.Fatalf("Error openning database for app2: %+v", err)
	}

	databaseContext3, err = openDB(config3)
	if err != nil {
		t.Fatalf("Error openning database for app3: %+v", err)
	}

	return databaseContext1, databaseContext2, databaseContext3
}

func openDB(cfg *config.Config) (*dbaccess.DatabaseContext, error) {
	dbPath := filepath.Join(cfg.DataDir, "db")
	return dbaccess.New(dbPath)
}
