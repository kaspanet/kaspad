// All tests within this file should call testForAllDatabaseTypes
// over the actual test. This is to make sure that all supported
// database types adhere to the assumptions defined in the
// interfaces in this package.

package database_test

import (
	"bytes"
	"testing"

	"github.com/kaspanet/kaspad/infrastructure/db/database"
)

func TestDatabasePut(t *testing.T) {
	testForAllDatabaseTypes(t, "TestDatabasePut", testDatabasePut)
}

func testDatabasePut(t *testing.T, db database.Database, testName string) {
	// Put value1 into the database
	key := database.MakeBucket().Key([]byte("key"))
	value1 := []byte("value1")
	err := db.Put(key, value1)
	if err != nil {
		t.Fatalf("%s: Put "+
			"unexpectedly failed: %s", testName, err)
	}

	// Make sure that the returned value is value1
	returnedValue, err := db.Get(key)
	if err != nil {
		t.Fatalf("%s: Get "+
			"unexpectedly failed: %s", testName, err)
	}
	if !bytes.Equal(returnedValue, value1) {
		t.Fatalf("%s: Get "+
			"returned wrong value. Want: %s, got: %s",
			testName, string(value1), string(returnedValue))
	}

	// Put value2 into the database with the same key
	value2 := []byte("value2")
	err = db.Put(key, value2)
	if err != nil {
		t.Fatalf("%s: Put "+
			"unexpectedly failed: %s", testName, err)
	}

	// Make sure that the returned value is value2
	returnedValue, err = db.Get(key)
	if err != nil {
		t.Fatalf("%s: Get "+
			"unexpectedly failed: %s", testName, err)
	}
	if !bytes.Equal(returnedValue, value2) {
		t.Fatalf("%s: Get "+
			"returned wrong value. Want: %s, got: %s",
			testName, string(value2), string(returnedValue))
	}
}

func TestDatabaseGet(t *testing.T) {
	testForAllDatabaseTypes(t, "TestDatabaseGet", testDatabaseGet)
}

func testDatabaseGet(t *testing.T, db database.Database, testName string) {
	// Put a value into the database
	key := database.MakeBucket().Key([]byte("key"))
	value := []byte("value")
	err := db.Put(key, value)
	if err != nil {
		t.Fatalf("%s: Put "+
			"unexpectedly failed: %s", testName, err)
	}

	// Get the value back and make sure it's the same one
	returnedValue, err := db.Get(key)
	if err != nil {
		t.Fatalf("%s: Get "+
			"unexpectedly failed: %s", testName, err)
	}
	if !bytes.Equal(returnedValue, value) {
		t.Fatalf("%s: Get "+
			"returned wrong value. Want: %s, got: %s",
			testName, string(value), string(returnedValue))
	}

	// Try getting a non-existent value and make sure
	// the returned error is ErrNotFound
	_, err = db.Get(database.MakeBucket().Key([]byte("doesn't exist")))
	if err == nil {
		t.Fatalf("%s: Get "+
			"unexpectedly succeeded", testName)
	}
	if !database.IsNotFoundError(err) {
		t.Fatalf("%s: Get "+
			"returned wrong error: %s", testName, err)
	}
}

func TestDatabaseHas(t *testing.T) {
	testForAllDatabaseTypes(t, "TestDatabaseHas", testDatabaseHas)
}

func testDatabaseHas(t *testing.T, db database.Database, testName string) {
	// Put a value into the database
	key := database.MakeBucket().Key([]byte("key"))
	value := []byte("value")
	err := db.Put(key, value)
	if err != nil {
		t.Fatalf("%s: Put "+
			"unexpectedly failed: %s", testName, err)
	}

	// Make sure that Has returns true for the value we just put
	exists, err := db.Has(key)
	if err != nil {
		t.Fatalf("%s: Has "+
			"unexpectedly failed: %s", testName, err)
	}
	if !exists {
		t.Fatalf("%s: Has "+
			"unexpectedly returned that the value does not exist", testName)
	}

	// Make sure that Has returns false for a non-existent value
	exists, err = db.Has(database.MakeBucket().Key([]byte("doesn't exist")))
	if err != nil {
		t.Fatalf("%s: Has "+
			"unexpectedly failed: %s", testName, err)
	}
	if exists {
		t.Fatalf("%s: Has "+
			"unexpectedly returned that the value exists", testName)
	}
}

func TestDatabaseDelete(t *testing.T) {
	testForAllDatabaseTypes(t, "TestDatabaseDelete", testDatabaseDelete)
}

func testDatabaseDelete(t *testing.T, db database.Database, testName string) {
	// Put a value into the database
	key := database.MakeBucket().Key([]byte("key"))
	value := []byte("value")
	err := db.Put(key, value)
	if err != nil {
		t.Fatalf("%s: Put "+
			"unexpectedly failed: %s", testName, err)
	}

	// Delete the value
	err = db.Delete(key)
	if err != nil {
		t.Fatalf("%s: Delete "+
			"unexpectedly failed: %s", testName, err)
	}

	// Make sure that Has returns false for the deleted value
	exists, err := db.Has(key)
	if err != nil {
		t.Fatalf("%s: Has "+
			"unexpectedly failed: %s", testName, err)
	}
	if exists {
		t.Fatalf("%s: Has "+
			"unexpectedly returned that the value exists", testName)
	}
}
