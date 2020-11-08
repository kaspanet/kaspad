package consensus

import (
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"
	"io/ioutil"
	"testing"
)

func TestNewConsensus(t *testing.T) {
	f := NewFactory()

	dagParams := &dagconfig.DevnetParams

	tmpDir, err := ioutil.TempDir("", "TestNewConsensus")
	if err != nil {
		return
	}

	db, err := ldb.NewLevelDB(tmpDir)
	if err != nil {
		t.Fatalf("error in NewLevelDB: %s", err)
	}

	_, err = f.NewConsensus(dagParams, db)
	if err != nil {
		t.Fatalf("error in NewConsensus: %+v", err)
	}
}
