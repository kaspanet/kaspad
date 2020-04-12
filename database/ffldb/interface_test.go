package ffldb

import (
	"bytes"
	"fmt"
	"github.com/kaspanet/kaspad/database"
	"io/ioutil"
	"reflect"
	"testing"
)

type keyValuePair struct {
	key   *database.Key
	value []byte
}

func prepareCursorForTest(t *testing.T, testName string, entries []keyValuePair) (cursor database.Cursor, teardownFunc func()) {
	// Create a temp db to run tests against
	path, err := ioutil.TempDir("", testName)
	if err != nil {
		t.Fatalf("%s: TempDir unexpectedly "+
			"failed: %s", testName, err)
	}
	db, err := Open(path)
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

	// Put them into the database
	for _, entry := range entries {
		err := db.Put(entry.key, entry.value)
		if err != nil {
			t.Fatalf("%s: Put unexpectedly "+
				"failed: %s", testName, err)
		}
	}

	cursor, err = db.Cursor(database.MakeBucket())
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
