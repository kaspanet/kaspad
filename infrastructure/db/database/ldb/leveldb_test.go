package ldb

import (
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/kaspanet/kaspad/infrastructure/db/database"
)

func prepareDatabaseForTest(t *testing.T, testName string) (ldb *LevelDB, teardownFunc func()) {
	// Create a temp db to run tests against
	path, err := ioutil.TempDir("", testName)
	if err != nil {
		t.Fatalf("%s: TempDir unexpectedly "+
			"failed: %s", testName, err)
	}
	ldb, err = NewLevelDB(path)
	if err != nil {
		t.Fatalf("%s: NewLevelDB unexpectedly "+
			"failed: %s", testName, err)
	}
	teardownFunc = func() {
		err = ldb.Close()
		if err != nil {
			t.Fatalf("%s: Close unexpectedly "+
				"failed: %s", testName, err)
		}
	}
	return ldb, teardownFunc
}

func TestLevelDBSanity(t *testing.T) {
	ldb, teardownFunc := prepareDatabaseForTest(t, "TestLevelDBSanity")
	defer teardownFunc()

	// Put something into the db
	key := database.MakeBucket(nil).Key([]byte("key"))
	putData := []byte("Hello world!")
	err := ldb.Put(key, putData)
	if err != nil {
		t.Fatalf("TestLevelDBSanity: Put returned "+
			"unexpected error: %s", err)
	}

	// Get from the key previously put to
	getData, err := ldb.Get(key)
	if err != nil {
		t.Fatalf("TestLevelDBSanity: Get returned "+
			"unexpected error: %s", err)
	}

	// Make sure that the put data and the get data are equal
	if !reflect.DeepEqual(getData, putData) {
		t.Fatalf("TestLevelDBSanity: get data and "+
			"put data are not equal. Put: %s, got: %s",
			string(putData), string(getData))
	}
}

func TestLevelDBTransactionSanity(t *testing.T) {
	ldb, teardownFunc := prepareDatabaseForTest(t, "TestLevelDBTransactionSanity")
	defer teardownFunc()

	// Case 1. Write in tx and then read directly from the DB
	// Begin a new transaction
	tx, err := ldb.Begin()
	if err != nil {
		t.Fatalf("TestLevelDBTransactionSanity: Begin "+
			"unexpectedly failed: %s", err)
	}

	// Put something into the transaction
	key := database.MakeBucket(nil).Key([]byte("key"))
	putData := []byte("Hello world!")
	err = tx.Put(key, putData)
	if err != nil {
		t.Fatalf("TestLevelDBTransactionSanity: Put "+
			"returned unexpected error: %s", err)
	}

	// Get from the key previously put to. Since the tx is not
	// yet committed, this should return ErrNotFound.
	_, err = ldb.Get(key)
	if err == nil {
		t.Fatalf("TestLevelDBTransactionSanity: Get " +
			"unexpectedly succeeded")
	}
	if !database.IsNotFoundError(err) {
		t.Fatalf("TestLevelDBTransactionSanity: Get "+
			"returned wrong error: %s", err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		t.Fatalf("TestLevelDBTransactionSanity: Commit "+
			"returned unexpected error: %s", err)
	}

	// Get from the key previously put to. Now that the tx was
	// committed, this should succeed.
	getData, err := ldb.Get(key)
	if err != nil {
		t.Fatalf("TestLevelDBTransactionSanity: Get "+
			"returned unexpected error: %s", err)
	}

	// Make sure that the put data and the get data are equal
	if !reflect.DeepEqual(getData, putData) {
		t.Fatalf("TestLevelDBTransactionSanity: get "+
			"data and put data are not equal. Put: %s, got: %s",
			string(putData), string(getData))
	}

	// Case 2. Write directly to the DB and then read from a tx
	// Put something into the db
	key = database.MakeBucket(nil).Key([]byte("key2"))
	putData = []byte("Goodbye world!")
	err = ldb.Put(key, putData)
	if err != nil {
		t.Fatalf("TestLevelDBTransactionSanity: Put "+
			"returned unexpected error: %s", err)
	}

	// Begin a new transaction
	tx, err = ldb.Begin()
	if err != nil {
		t.Fatalf("TestLevelDBTransactionSanity: Begin "+
			"unexpectedly failed: %s", err)
	}

	// Get from the key previously put to
	getData, err = tx.Get(key)
	if err != nil {
		t.Fatalf("TestLevelDBTransactionSanity: Get "+
			"returned unexpected error: %s", err)
	}

	// Make sure that the put data and the get data are equal
	if !reflect.DeepEqual(getData, putData) {
		t.Fatalf("TestLevelDBTransactionSanity: get "+
			"data and put data are not equal. Put: %s, got: %s",
			string(putData), string(getData))
	}

	// Rollback the transaction
	err = tx.Rollback()
	if err != nil {
		t.Fatalf("TestLevelDBTransactionSanity: rollback "+
			"returned unexpected error: %s", err)
	}
}
