package integration

import (
	"path/filepath"
	"testing"

	"github.com/kaspanet/kaspad/app"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/dbaccess"
)

type appHarness struct {
	app             *app.App
	rpcClient       *rpcClient
	p2pAddress      string
	rpcAddress      string
	config          *config.Config
	databaseContext *dbaccess.DatabaseContext
}

func setup(t *testing.T) (appHarness1, appHarness2, appHarness3 *appHarness, teardownFunc func()) {
	appHarness1 = &appHarness{p2pAddress: p2pAddress1, rpcAddress: rpcAddress1}
	appHarness2 = &appHarness{p2pAddress: p2pAddress2, rpcAddress: rpcAddress2}
	appHarness3 = &appHarness{p2pAddress: p2pAddress3, rpcAddress: rpcAddress3}

	setConfig(t, appHarness1)
	setConfig(t, appHarness2)
	setConfig(t, appHarness3)

	setDatabaseContext(t, appHarness1)
	setDatabaseContext(t, appHarness2)
	setDatabaseContext(t, appHarness3)

	setApp(t, appHarness1)
	setApp(t, appHarness2)
	setApp(t, appHarness3)

	appHarness1.app.Start()
	appHarness2.app.Start()
	appHarness3.app.Start()

	setRPCClient(t, appHarness1)
	setRPCClient(t, appHarness2)
	setRPCClient(t, appHarness3)

	return appHarness1, appHarness2, appHarness3,
		func() {
			teardown(t, appHarness1)
			teardown(t, appHarness2)
			teardown(t, appHarness3)
		}
}

func setRPCClient(t *testing.T, harness *appHarness) {
	var err error
	harness.rpcClient, err = newRPCClient(harness.rpcAddress)
	if err != nil {
		t.Fatalf("Error getting RPC client %+v", err)
	}
}
func teardown(t *testing.T, harness *appHarness) {
	err := harness.app.Stop()
	if err != nil {
		t.Errorf("Error stopping App: %+v", err)
	}

	harness.app.WaitForShutdown()

	err = harness.databaseContext.Close()
	if err != nil {
		t.Errorf("Error closing database context: %+v", err)
	}
}

func setApp(t *testing.T, harness *appHarness) {
	var err error
	harness.app, err = app.New(harness.config, harness.databaseContext, make(chan struct{}))
	if err != nil {
		t.Fatalf("Error creating app: %+v", err)
	}
}

func setDatabaseContext(t *testing.T, harness *appHarness) {
	var err error
	harness.databaseContext, err = openDB(harness.config)
	if err != nil {
		t.Fatalf("Error openning database: %+v", err)
	}
}

func openDB(cfg *config.Config) (*dbaccess.DatabaseContext, error) {
	dbPath := filepath.Join(cfg.DataDir, "db")
	return dbaccess.New(dbPath)
}
