package database_test

import (
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/database/ffldb"
	"io/ioutil"
	"testing"
)

var databasePrepareFuncs = []func(t *testing.T, testName string) (db database.Database, name string, teardownFunc func()){
	prepareFFLDBForTest,
}

func prepareFFLDBForTest(t *testing.T, testName string) (db database.Database, name string, teardownFunc func()) {
	// Create a temp db to run tests against
	path, err := ioutil.TempDir("", testName)
	if err != nil {
		t.Fatalf("%s: TempDir unexpectedly "+
			"failed: %s", testName, err)
	}
	db, err = ffldb.Open(path)
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
	return db, "ffldb", teardownFunc
}
