package ffldb

import (
	"bytes"
	"fmt"
	"github.com/kaspanet/kaspad/database"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

type keyValuePair struct {
	key   *database.Key
	value []byte
}

func prepareDatabaseForTest(t *testing.T, testName string) (db database.Database, teardownFunc func()) {
	// Create a temp db to run tests against
	path, err := ioutil.TempDir("", testName)
	if err != nil {
		t.Fatalf("%s: TempDir unexpectedly "+
			"failed: %s", testName, err)
	}
	db, err = Open(path)
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
	return db, teardownFunc
}

func prepareCursorForTest(t *testing.T, testName string, entries []keyValuePair) (cursor database.Cursor, teardownFunc func()) {
	db, teardownFunc := prepareDatabaseForTest(t, testName)

	// Put the entries into the database
	for _, entry := range entries {
		err := db.Put(entry.key, entry.value)
		if err != nil {
			t.Fatalf("%s: Put unexpectedly "+
				"failed: %s", testName, err)
		}
	}

	cursor, err := db.Cursor(database.MakeBucket())
	if err != nil {
		t.Fatalf("%s: Cursor unexpectedly "+
			"failed: %s", testName, err)
	}

	return cursor, teardownFunc
}

func prepareKeyValuePairsForTest() []keyValuePair {
	// Prepare a list of key/value pairs
	entries := make([]keyValuePair, 10)
	for i := 0; i < 10; i++ {
		key := database.MakeBucket().Key([]byte(fmt.Sprintf("key%d", i)))
		value := []byte("value")
		entries[i] = keyValuePair{key: key, value: value}
	}
	return entries
}

func TestCursorNext(t *testing.T) {
	entries := prepareKeyValuePairsForTest()
	cursor, teardownFunc := prepareCursorForTest(t, "TestCursorNext", entries)
	defer teardownFunc()

	// Make sure that all the entries exist in the cursor, in their
	// correct order
	for _, entry := range entries {
		hasNext := cursor.Next()
		if !hasNext {
			t.Fatalf("TestCursorNext: cursor unexpectedly " +
				"done")
		}
		cursorKey, err := cursor.Key()
		if err != nil {
			t.Fatalf("TestCursorNext: Key unexpectedly "+
				"failed: %s", err)
		}
		if !reflect.DeepEqual(cursorKey, entry.key) {
			t.Fatalf("TestCursorNext: Cursor returned "+
				"wrong key. Want: %s, got: %s", entry.key, cursorKey)
		}
		cursorValue, err := cursor.Value()
		if err != nil {
			t.Fatalf("TestCursorNext: Value unexpectedly "+
				"failed: %s", err)
		}
		if !bytes.Equal(cursorValue, entry.value) {
			t.Fatalf("TestCursorNext: Cursor returned "+
				"wrong value. Want: %s, got: %s", entry.value, cursorValue)
		}
	}

	// The cursor should now be exhausted. Make sure Next now
	// returns false
	hasNext := cursor.Next()
	if hasNext {
		t.Fatalf("TestCursorNext: cursor unexpectedly " +
			"not done")
	}

	// Rewind the cursor, close it, and call Next on it again.
	// This time it should return false because it's closed.
	cursor.First()
	err := cursor.Close()
	if err != nil {
		t.Fatalf("TestCursorNext: Close unexpectedly "+
			"failed: %s", err)
	}
	hasNext = cursor.Next()
	if hasNext {
		t.Fatalf("TestCursorNext: cursor unexpectedly " +
			"returned true after being closed")
	}
}

func TestCursorFirst(t *testing.T) {
	entries := prepareKeyValuePairsForTest()
	cursor, teardownFunc := prepareCursorForTest(t, "TestCursorFirstWithEntries", entries)
	defer teardownFunc()

	// Make sure that First returns true when the cursor is not empty
	exists := cursor.First()
	if !exists {
		t.Fatalf("TestCursorFirst: Cursor unexpectedly " +
			"returned false")
	}

	// Make sure that the first key and value are as expected
	firstEntryKey := entries[0].key
	firstCursorKey, err := cursor.Key()
	if err != nil {
		t.Fatalf("TestCursorFirst: Key unexpectedly "+
			"failed: %s", err)
	}
	if !reflect.DeepEqual(firstCursorKey, firstEntryKey) {
		t.Fatalf("TestCursorFirst: Cursor returned "+
			"wrong key. Want: %s, got: %s", firstEntryKey, firstCursorKey)
	}
	firstEntryValue := entries[0].value
	firstCursorValue, err := cursor.Value()
	if err != nil {
		t.Fatalf("TestCursorFirst: Value unexpectedly "+
			"failed: %s", err)
	}
	if !bytes.Equal(firstCursorValue, firstEntryValue) {
		t.Fatalf("TestCursorFirst: Cursor returned "+
			"wrong value. Want: %s, got: %s", firstEntryValue, firstCursorValue)
	}

	// Create a new cursor over an empty dataset
	cursor, teardownFunc = prepareCursorForTest(t, "TestCursorFirstWithoutEntries", nil)
	defer teardownFunc()

	// Make sure that First returns false when the cursor is empty
	exists = cursor.First()
	if exists {
		t.Fatalf("TestCursorFirst: Cursor unexpectedly " +
			"returned true")
	}
}

func TestCursorSeek(t *testing.T) {
	entries := prepareKeyValuePairsForTest()
	cursor, teardownFunc := prepareCursorForTest(t, "TestCursorSeek", entries)
	defer teardownFunc()

	// Seek to the fourth entry and make sure it exists
	fourthEntry := entries[3]
	err := cursor.Seek(fourthEntry.key)
	if err != nil {
		t.Fatalf("TestCursorSeek: Cursor unexpectedly "+
			"failed: %s", err)
	}

	// Make sure that the key and value are as expected
	fourthEntryKey := entries[3].key
	fourthCursorKey, err := cursor.Key()
	if err != nil {
		t.Fatalf("TestCursorSeek: Key unexpectedly "+
			"failed: %s", err)
	}
	if !reflect.DeepEqual(fourthCursorKey, fourthEntryKey) {
		t.Fatalf("TestCursorSeek: Cursor returned "+
			"wrong key. Want: %s, got: %s", fourthEntryKey, fourthCursorKey)
	}
	fourthEntryValue := entries[3].value
	fourthCursorValue, err := cursor.Value()
	if err != nil {
		t.Fatalf("TestCursorSeek: Value unexpectedly "+
			"failed: %s", err)
	}
	if !bytes.Equal(fourthCursorValue, fourthEntryValue) {
		t.Fatalf("TestCursorSeek: Cursor returned "+
			"wrong value. Want: %s, got: %s", fourthEntryValue, fourthCursorValue)
	}

	// Seek to a value that doesn't exist and make sure that
	// the returned error is ErrNotFound
	err = cursor.Seek(database.MakeBucket().Key([]byte("doesn't exist")))
	if err == nil {
		t.Fatalf("TestCursorSeek: Seek unexpectedly " +
			"succeeded")
	}
	if !database.IsNotFoundError(err) {
		t.Fatalf("TestCursorSeek: Seek returned "+
			"wrong error: %s", err)
	}
}

func TestCursorCloseErrors(t *testing.T) {
	entries := prepareKeyValuePairsForTest()
	cursor, teardownFunc := prepareCursorForTest(t, "TestCursorCloseErrors", entries)
	defer teardownFunc()

	// Close the cursor
	err := cursor.Close()
	if err != nil {
		t.Fatalf("TestCursorCloseErrors: Close "+
			"unexpectedly failed: %s", err)
	}

	tests := []struct {
		name     string
		function func() error
	}{
		{
			name: "Seek",
			function: func() error {
				return cursor.Seek(database.MakeBucket().Key([]byte{}))
			},
		},
		{
			name: "Key",
			function: func() error {
				_, err := cursor.Key()
				return err
			},
		},
		{
			name: "Value",
			function: func() error {
				_, err := cursor.Value()
				return err
			},
		},
		{
			name: "Close",
			function: func() error {
				return cursor.Close()
			},
		},
	}

	for _, test := range tests {
		expectedErrContainsString := "closed cursor"

		// Make sure that the test function returns a "closed cursor" error
		err = test.function()
		if err == nil {
			t.Fatalf("TestCursorCloseErrors: %s "+
				"unexpectedly succeeded", test.name)
		}
		if !strings.Contains(err.Error(), expectedErrContainsString) {
			t.Fatalf("TestCursorCloseErrors: %s "+
				"returned wrong error. Want: %s, got: %s",
				test.name, expectedErrContainsString, err)
		}
	}
}

func TestCursorCloseFirstAndNext(t *testing.T) {
	entries := prepareKeyValuePairsForTest()
	cursor, teardownFunc := prepareCursorForTest(t, "TestCursorCloseFirstAndNext", entries)
	defer teardownFunc()

	// Close the cursor
	err := cursor.Close()
	if err != nil {
		t.Fatalf("TestCursorCloseFirstAndNext: Close "+
			"unexpectedly failed: %s", err)
	}

	// We expect First to return false
	result := cursor.First()
	if result {
		t.Fatalf("TestCursorCloseFirstAndNext: First " +
			"unexpectedly returned true")
	}

	// We expect Next to return false
	result = cursor.Next()
	if result {
		t.Fatalf("TestCursorCloseFirstAndNext: Next " +
			"unexpectedly returned true")
	}
}

func TestDataAccessorPut(t *testing.T) {
	db, teardownFunc := prepareDatabaseForTest(t, "TestDataAccessorPut")
	defer teardownFunc()

	// Put value1 into the database
	key := database.MakeBucket().Key([]byte("key"))
	value1 := []byte("value1")
	err := db.Put(key, value1)
	if err != nil {
		t.Fatalf("TestDataAccessorPut: Put "+
			"unexpectedly failed: %s", err)
	}

	// Put value2 into the database with the same key
	value2 := []byte("value2")
	err = db.Put(key, value2)
	if err != nil {
		t.Fatalf("TestDataAccessorPut: Put "+
			"unexpectedly failed: %s", err)
	}

	// Make sure that the returned value is value2
	returnedValue, err := db.Get(key)
	if err != nil {
		t.Fatalf("TestDataAccessorPut: Get "+
			"unexpectedly failed: %s", err)
	}
	if !bytes.Equal(returnedValue, value2) {
		t.Fatalf("TestDataAccessorPut: Get "+
			"returned wrong value. Want: %s, got: %s",
			string(value2), string(returnedValue))
	}
}

func TestDataAccessorGet(t *testing.T) {
	db, teardownFunc := prepareDatabaseForTest(t, "TestDataAccessorGet")
	defer teardownFunc()

	// Put a value into the database
	key := database.MakeBucket().Key([]byte("key"))
	value := []byte("value")
	err := db.Put(key, value)
	if err != nil {
		t.Fatalf("TestDataAccessorGet: Put "+
			"unexpectedly failed: %s", err)
	}

	// Get the value back and make sure it's the same one
	returnedValue, err := db.Get(key)
	if err != nil {
		t.Fatalf("TestDataAccessorGet: Get "+
			"unexpectedly failed: %s", err)
	}
	if !bytes.Equal(returnedValue, value) {
		t.Fatalf("TestDataAccessorGet: Get "+
			"returned wrong value. Want: %s, got: %s",
			string(value), string(returnedValue))
	}

	// Try getting a non-existent value and make sure
	// the returned error is ErrNotFound
	_, err = db.Get(database.MakeBucket().Key([]byte("doesn't exist")))
	if err == nil {
		t.Fatalf("TestDataAccessorGet: Get " +
			"unexpectedly succeeded")
	}
	if !database.IsNotFoundError(err) {
		t.Fatalf("TestDataAccessorPut: Get "+
			"returned wrong error: %s", err)
	}
}

func TestDataAccessorHas(t *testing.T) {
	db, teardownFunc := prepareDatabaseForTest(t, "TestDataAccessorHas")
	defer teardownFunc()

	// Put a value into the database
	key := database.MakeBucket().Key([]byte("key"))
	value := []byte("value")
	err := db.Put(key, value)
	if err != nil {
		t.Fatalf("TestDataAccessorHas: Put "+
			"unexpectedly failed: %s", err)
	}

	// Make sure that Has returns true for the value we just put
	exists, err := db.Has(key)
	if err != nil {
		t.Fatalf("TestDataAccessorGet: Has "+
			"unexpectedly failed: %s", err)
	}
	if !exists {
		t.Fatalf("TestDataAccessorGet: Has " +
			"unexpectedly returned that the value does not exist")
	}

	// Make sure that Has returns false for a non-existent value
	exists, err = db.Has(database.MakeBucket().Key([]byte("doesn't exist")))
	if err != nil {
		t.Fatalf("TestDataAccessorGet: Has "+
			"unexpectedly failed: %s", err)
	}
	if exists {
		t.Fatalf("TestDataAccessorGet: Has " +
			"unexpectedly returned that the value exists")
	}
}
