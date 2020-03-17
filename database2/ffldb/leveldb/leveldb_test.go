package leveldb

import (
	"os"
	"reflect"
	"testing"
)

func TestLevelDBSanity(t *testing.T) {
	// Open a test db
	path := os.TempDir()
	name := "test"
	ldb, err := NewLevelDB(path, name)
	if err != nil {
		t.Fatalf("TestLevelDBSanity: NewLevelDB "+
			"unexpectedly failed: %s", err)
	}
	defer func() {
		err := ldb.Close()
		if err != nil {
			t.Fatalf("TestLevelDBSanity: Close "+
				"unexpectedly failed: %s", err)
		}
	}()

	// Put something into the db
	key := []byte("key")
	putData := []byte("Hello world!")
	err = ldb.Put(key, putData)
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
	// Open a test db
	path := os.TempDir()
	name := "test"
	ldb, err := NewLevelDB(path, name)
	if err != nil {
		t.Fatalf("TestLevelDBTransactionSanity: NewLevelDB "+
			"unexpectedly failed: %s", err)
	}
	defer func() {
		err := ldb.Close()
		if err != nil {
			t.Fatalf("TestLevelDBTransactionSanity: Close "+
				"unexpectedly failed: %s", err)
		}
	}()

	// Begin a new transaction
	tx, err := ldb.Begin()
	if err != nil {
		t.Fatalf("TestLevelDBTransactionSanity: Begin "+
			"unexpectedly failed: %s", err)
	}

	// Put something into the transaction
	key := []byte("key")
	putData := []byte("Hello world!")
	err = tx.Put(key, putData)
	if err != nil {
		t.Fatalf("TestLevelDBTransactionSanity: Put "+
			"returned unexpected error: %s", err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		t.Fatalf("TestLevelDBTransactionSanity: Commit "+
			"returned unexpected error: %s", err)
	}

	// Get from the key previously put to
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
}
