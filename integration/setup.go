package integration

import (
	"path/filepath"
	"testing"

	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/dbaccess"
	kaspadpkg "github.com/kaspanet/kaspad/kaspad"
)

func setup(t *testing.T) (kaspad1, kaspad2 *kaspadpkg.Kaspad, teardown func()) {
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

	return kaspad1, kaspad2, func() {
		err := kaspad1DatabaseContext.Close()
		if err != nil {
			t.Errorf("Error closing kaspad1DatabaseContext: %+v", err)
		}
		err = kaspad2DatabaseContext.Close()
		if err != nil {
			t.Errorf("Error closing kaspad2DatabaseContext: %+v", err)
		}
		close(kaspad1Interrupt)
		close(kaspad2Interrupt)
	}
}

func openDB(cfg *config.Config) (*dbaccess.DatabaseContext, error) {
	dbPath := filepath.Join(cfg.DataDir, "db")
	return dbaccess.New(dbPath)
}
