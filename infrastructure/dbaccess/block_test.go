package dbaccess

import (
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util"
)

func TestBlockStoreSanity(t *testing.T) {
	// Create a temp db to run tests against
	path, err := ioutil.TempDir("", "TestBlockStoreSanity")
	if err != nil {
		t.Fatalf("TestBlockStoreSanity: TempDir unexpectedly "+
			"failed: %s", err)
	}
	databaseContext, err := New(path)
	if err != nil {
		t.Fatalf("TestBlockStoreSanity: Open unexpectedly "+
			"failed: %s", err)
	}
	defer func() {
		err := databaseContext.Close()
		if err != nil {
			t.Fatalf("TestBlockStoreSanity: Close unexpectedly "+
				"failed: %s", err)
		}
	}()

	// Store the genesis block
	genesis := util.NewBlock(dagconfig.MainnetParams.GenesisBlock)
	genesisHash := genesis.Hash()
	genesisBytes, err := genesis.Bytes()
	if err != nil {
		t.Fatalf("TestBlockStoreSanity: util.Block.Bytes unexpectedly "+
			"failed: %s", err)
	}
	dbTx, err := databaseContext.NewTx()
	if err != nil {
		t.Fatalf("Failed to open database "+
			"transaction: %s", err)
	}
	defer dbTx.RollbackUnlessClosed()
	err = StoreBlock(dbTx, genesisHash, genesisBytes)
	if err != nil {
		t.Fatalf("TestBlockStoreSanity: StoreBlock unexpectedly "+
			"failed: %s", err)
	}
	err = dbTx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit database "+
			"transaction: %s", err)
	}

	// Make sure the genesis block now exists in the db
	exists, err := HasBlock(databaseContext, genesisHash)
	if err != nil {
		t.Fatalf("TestBlockStoreSanity: HasBlock unexpectedly "+
			"failed: %s", err)
	}
	if !exists {
		t.Fatalf("TestBlockStoreSanity: just-inserted block is " +
			"missing from the database")
	}

	// Fetch the genesis block back from the db and make sure
	// that it's equal to the original
	fetchedGenesisBytes, err := FetchBlock(databaseContext, genesisHash)
	if err != nil {
		t.Fatalf("TestBlockStoreSanity: FetchBlock unexpectedly "+
			"failed: %s", err)
	}
	fetchedGenesis, err := util.NewBlockFromBytes(fetchedGenesisBytes)
	if err != nil {
		t.Fatalf("TestBlockStoreSanity: NewBlockFromBytes unexpectedly "+
			"failed: %s", err)
	}
	if !reflect.DeepEqual(genesis.MsgBlock(), fetchedGenesis.MsgBlock()) {
		t.Fatalf("TestBlockStoreSanity: just-inserted block is " +
			"not equal to its database counterpart.")
	}
}
