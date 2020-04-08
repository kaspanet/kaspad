package ffldb

import (
	"bytes"
	"fmt"
	"github.com/kaspanet/kaspad/database"
	"io/ioutil"
	"testing"
)

type keyValuePair struct {
	key   []byte
	value []byte
}

func prepareCursorForTest(t *testing.T, testName string) (cursor database.Cursor, entries []keyValuePair, teardownFunc func()) {
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

	// Prepare a list of key/value pairs
	entries = make([]keyValuePair, 10)
	for i := 0; i < 10; i++ {
		key := []byte(fmt.Sprintf("key%d", i))
		value := []byte("value")
		entries[i] = keyValuePair{key: key, value: value}
	}

	// Put them into the database
	for _, entry := range entries {
		err := db.Put(entry.key, entry.value)
		if err != nil {
			t.Fatalf("%s: Put unexpectedly "+
				"failed: %s", testName, err)
		}
	}

	cursor, err = db.Cursor([]byte{})
	if err != nil {
		t.Fatalf("%s: Cursor unexpectedly "+
			"failed: %s", testName, err)
	}

	return cursor, entries, teardownFunc
}

func TestCursorNext(t *testing.T) {
	cursor, entries, teardownFunc := prepareCursorForTest(t, "TestCursorNext")
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
		if !bytes.Equal(cursorKey, entry.key) {
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
