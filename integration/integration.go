package integration

import (
	"path/filepath"
	"testing"

	kaspadpkg "github.com/kaspanet/kaspad/kaspad"

	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/dbaccess"
)

func TestIntegration(t *testing.T) {
	kaspad1Config, kaspad2Config := configs()

	kaspad1DatabaseContext, err := openDB(kaspad1Config)
	if err != nil {
		t.Fatalf("Error openning database for kaspad1: %+v", err)
	}
	defer kaspad1DatabaseContext.Close()

	kaspad2DatabaseContext, err := openDB(kaspad2Config)
	if err != nil {
		t.Fatalf("Error openning database for kaspad2: %+v", err)
	}
	defer kaspad2DatabaseContext.Close()

	kaspad1Interrupt, kaspad2Interrupt := make(chan struct{}), make(chan struct{})

	kaspad1, err := kaspadpkg.New(kaspad1Config, kaspad1DatabaseContext, kaspad1Interrupt)
	if err != nil {
		t.Fatalf("Error creating kaspad1: %+v", err)
	}

	kaspad2, err := kaspadpkg.New(kaspad2Config, kaspad2DatabaseContext, kaspad2Interrupt)
	if err != nil {
		t.Fatalf("Error creating kaspad2: %+v", err)
	}

	defer func() {
		close(kaspad1Interrupt)
		close(kaspad2Interrupt)
	}()

	kaspad1.Start()
	kaspad2.Start()
}

func openDB(cfg *config.Config) (*dbaccess.DatabaseContext, error) {
	dbPath := filepath.Join(cfg.DataDir, "db")
	return dbaccess.New(dbPath)
}
