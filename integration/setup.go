package integration

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/rpc/client"

	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/dbaccess"
	kaspadpkg "github.com/kaspanet/kaspad/kaspad"
)

func setup(t *testing.T) (kaspad1, kaspad2 *kaspadpkg.Kaspad, client1, client2 *client.Client, teardown func()) {
	kaspad1Config, kaspad2Config := configs(t)

	kaspad1DatabaseContext, err := openDB(kaspad1Config)
	if err != nil {
		t.Fatalf("Error openning database for kaspad1: %+v", err)
	}

	kaspad2DatabaseContext, err := openDB(kaspad2Config)
	if err != nil {
		t.Fatalf("Error openning database for kaspad2: %+v", err)
	}

	kaspad1Interrupt, kaspad2Interrupt := make(chan struct{}), make(chan struct{})

	kaspad1, err = kaspadpkg.New(kaspad1Config, kaspad1DatabaseContext, kaspad1Interrupt)
	if err != nil {
		t.Fatalf("Error creating kaspad1: %+v", err)
	}

	kaspad2, err = kaspadpkg.New(kaspad2Config, kaspad2DatabaseContext, kaspad2Interrupt)
	if err != nil {
		t.Fatalf("Error creating kaspad2: %+v", err)
	}

	kaspad1.Start()
	kaspad2.Start()

	client1, err = rpcClient(kaspad1RPCAddress)
	if err != nil {
		t.Fatalf("Error getting RPC client for kaspad1 %+v", err)
	}
	client2, err = rpcClient(kaspad2RPCAddress)
	if err != nil {
		t.Fatalf("Error getting RPC client for kaspad2: %+v", err)
	}

	return kaspad1, kaspad2, client1, client2, func() {
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
	}
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
