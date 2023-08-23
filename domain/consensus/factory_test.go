package consensus

import (
	"io/ioutil"
	"testing"

	"github.com/c4ei/kaspad/domain/prefixmanager/prefix"

	"github.com/c4ei/kaspad/domain/dagconfig"
	"github.com/c4ei/kaspad/infrastructure/db/database/ldb"
)

func TestNewConsensus(t *testing.T) {
	f := NewFactory()

	config := &Config{Params: dagconfig.DevnetParams}

	tmpDir, err := ioutil.TempDir("", "TestNewConsensus")
	if err != nil {
		return
	}

	db, err := ldb.NewLevelDB(tmpDir, 8)
	if err != nil {
		t.Fatalf("error in NewLevelDB: %s", err)
	}

	_, shouldMigrate, err := f.NewConsensus(config, db, &prefix.Prefix{}, nil)
	if err != nil {
		t.Fatalf("error in NewConsensus: %+v", err)
	}

	if shouldMigrate {
		t.Fatalf("A fresh consensus should never return shouldMigrate=true")
	}
}
