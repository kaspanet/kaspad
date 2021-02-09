package database_test

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"
)

type databasePrepareFunc func(t *testing.T, testName string) (db database.Database, name string, teardownFunc func())

// databasePrepareFuncs is a set of functions, in which each function
// prepares a separate database type for testing.
// See testForAllDatabaseTypes for further details.
var databasePrepareFuncs = []databasePrepareFunc{
	prepareLDBForTest,
}

func prepareLDBForTest(t *testing.T, testName string) (db database.Database, name string, teardownFunc func()) {
	// Create a temp db to run tests against
	path, err := ioutil.TempDir("", testName)
	if err != nil {
		t.Fatalf("%s: TempDir unexpectedly "+
			"failed: %s", testName, err)
	}
	db, err = ldb.NewLevelDB(path, 8)
	if err != nil {
		t.Fatalf("%s: Open unexpectedly "+
			"failed: %s", testName, err)
	}
	teardownFunc = func() {
		err = db.Close()
		if err != nil {
			t.Fatalf("%s: Close unexpectedly "+
				"failed: %s", testName, err)
		}
	}
	return db, "ldb", teardownFunc
}

// testForAllDatabaseTypes runs the given testFunc for every database
// type defined in databasePrepareFuncs. This is to make sure that
// all supported database types adhere to the assumptions defined in
// the interfaces in this package.
func testForAllDatabaseTypes(t *testing.T, testName string,
	testFunc func(t *testing.T, db database.Database, testName string)) {

	for _, prepareDatabase := range databasePrepareFuncs {
		func() {
			db, dbType, teardownFunc := prepareDatabase(t, testName)
			defer teardownFunc()

			testName := fmt.Sprintf("%s: %s", dbType, testName)
			testFunc(t, db, testName)
		}()
	}
}

type keyValuePair struct {
	key   *database.Key
	value []byte
}

func populateDatabaseForTest(t *testing.T, db database.Database, testName string) []keyValuePair {
	// Prepare a list of key/value pairs
	entries := make([]keyValuePair, 10)
	for i := 0; i < 10; i++ {
		key := database.MakeBucket(nil).Key([]byte(fmt.Sprintf("key%d", i)))
		value := []byte("value")
		entries[i] = keyValuePair{key: key, value: value}
	}

	// Put the pairs into the database
	for _, entry := range entries {
		err := db.Put(entry.key, entry.value)
		if err != nil {
			t.Fatalf("%s: Put unexpectedly "+
				"failed: %s", testName, err)
		}
	}

	return entries
}
