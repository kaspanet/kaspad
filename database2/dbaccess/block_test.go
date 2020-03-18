package dbaccess

import (
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/database2"
	"github.com/kaspanet/kaspad/util"
	"os"
	"reflect"
	"testing"
)

func TestStoreBlock(t *testing.T) {
	// Create a temp db to run tests against
	path := os.TempDir()
	err := database2.Open(path)
	if err != nil {
		t.Fatalf("TestStoreBlock: Open unexpectedly "+
			"failed: %s", err)
	}
	defer func() {
		err := database2.Close()
		if err != nil {
			t.Fatalf("TestStoreBlock: Close unexpectedly "+
				"failed: %s", err)
		}
	}()

	// Store the genesis block
	genesis := util.NewBlock(dagconfig.MainnetParams.GenesisBlock)
	err = StoreBlock(NoTx(), genesis)
	if err != nil {
		t.Fatalf("TestStoreBlock: StoreBlock unexpectedly "+
			"failed: %s", err)
	}

	// Make sure the genesis block now exists in the db
	exists, err := HasBlock(NoTx(), genesis.Hash())
	if err != nil {
		t.Fatalf("TestStoreBlock: HasBlock unexpectedly "+
			"failed: %s", err)
	}
	if !exists {
		t.Fatalf("TestStoreBlock: just-inserted block is " +
			"missing from the database")
	}

	// Fetch the genesis block back from the db and make sure
	// that it's equal to the original
	fetchedGenesis, err := FetchBlock(NoTx(), genesis.Hash())
	if err != nil {
		t.Fatalf("TestStoreBlock: FetchBlock unexpectedly "+
			"failed: %s", err)
	}
	if !reflect.DeepEqual(genesis.MsgBlock(), fetchedGenesis.MsgBlock()) {
		t.Fatalf("TestStoreBlock: just-inserted block is " +
			"not equal to its database counterpart.")
	}
}
