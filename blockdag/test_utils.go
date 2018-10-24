package blockdag

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/database"
	_ "github.com/daglabs/btcd/database/ffldb"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/wire"
)

const (
	// testDbType is the database backend type to use for the tests.
	testDbType = "ffldb"

	// testDbRoot is the root directory used to create all test databases.
	testDbRoot = "testdbs"

	// blockDataNet is the expected network in the test block data.
	blockDataNet = wire.MainNet
)

// isSupportedDbType returns whether or not the passed database type is
// currently supported.
func isSupportedDbType(dbType string) bool {
	supportedDrivers := database.SupportedDrivers()
	for _, driver := range supportedDrivers {
		if dbType == driver {
			return true
		}
	}

	return false
}

// filesExists returns whether or not the named file or directory exists.
func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// DAGSetup is used to create a new db and chain instance with the genesis
// block already inserted.  In addition to the new chain instance, it returns
// a teardown function the caller should invoke when done testing to clean up.
func DAGSetup(dbName string, params *dagconfig.Params, config Config) (*BlockDAG, func(), error) {
	if !isSupportedDbType(testDbType) {
		return nil, nil, fmt.Errorf("unsupported db type %v", testDbType)
	}

	var teardown func()

	if config.DB == nil {
		// Handle memory database specially since it doesn't need the disk
		// specific handling.
		var db database.DB
		// Create the root directory for test databases.
		if !fileExists(testDbRoot) {
			if err := os.MkdirAll(testDbRoot, 0700); err != nil {
				err := fmt.Errorf("unable to create test db "+
					"root: %v", err)
				return nil, nil, err
			}
		}

		// Create a new database to store the accepted blocks into.
		dbPath := filepath.Join(testDbRoot, dbName)
		_ = os.RemoveAll(dbPath)
		ndb, err := database.Create(testDbType, dbPath, blockDataNet)
		if err != nil {
			return nil, nil, fmt.Errorf("error creating db: %v", err)
		}
		db = ndb

		config.DB = db

		// Setup a teardown function for cleaning up.  This function is
		// returned to the caller to be invoked when it is done testing.
		teardown = func() {
			db.Close()
			os.RemoveAll(dbPath)
			os.RemoveAll(testDbRoot)
		}
	}

	// Copy the chain params to ensure any modifications the tests do to
	// the chain parameters do not affect the global instance.
	paramsCopy := *params

	config.DAGParams = &paramsCopy
	config.TimeSource = NewMedianTime()
	config.SigCache = txscript.NewSigCache(1000)

	// Create the DAG instance.
	dag, err := New(&config)
	if err != nil {
		teardown()
		err := fmt.Errorf("failed to create dag instance: %v", err)
		return nil, nil, err
	}
	return dag, teardown, nil
}
