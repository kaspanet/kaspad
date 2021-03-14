// All tests within this file should call testForAllDatabaseTypes
// over the actual test. This is to make sure that all supported
// database types adhere to the assumptions defined in the
// interfaces in this package.

package database_test

import (
	"bytes"
	"fmt"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"reflect"
	"strings"
	"testing"
)

func prepareCursorForTest(t *testing.T, db database.Database, testName string) database.Cursor {
	cursor, err := db.Cursor(database.MakeBucket(nil))
	if err != nil {
		t.Fatalf("%s: Cursor unexpectedly "+
			"failed: %s", testName, err)
	}

	return cursor
}

func recoverFromClosedCursorPanic(t *testing.T, testName string) {
	panicErr := recover()
	if panicErr == nil {
		t.Fatalf("%s: cursor unexpectedly "+
			"didn't panic after being closed", testName)
	}
	expectedPanicErr := "closed cursor"
	if !strings.Contains(fmt.Sprintf("%v", panicErr), expectedPanicErr) {
		t.Fatalf("%s: cursor panicked "+
			"with wrong message. Want: %v, got: %s",
			testName, expectedPanicErr, panicErr)
	}
}

func TestCursorNext(t *testing.T) {
	testForAllDatabaseTypes(t, "TestCursorNext", testCursorNext)
}

func testCursorNext(t *testing.T, db database.Database, testName string) {
	entries := populateDatabaseForTest(t, db, testName)
	cursor := prepareCursorForTest(t, db, testName)

	// Make sure that all the entries exist in the cursor, in their
	// correct order
	for _, entry := range entries {
		hasNext := cursor.Next()
		if !hasNext {
			t.Fatalf("%s: cursor unexpectedly "+
				"done", testName)
		}
		cursorKey, err := cursor.Key()
		if err != nil {
			t.Fatalf("%s: Key unexpectedly "+
				"failed: %s", testName, err)
		}
		if !reflect.DeepEqual(cursorKey, entry.key) {
			t.Fatalf("%s: Cursor returned "+
				"wrong key. Want: %s, got: %s", testName, entry.key, cursorKey)
		}
		cursorValue, err := cursor.Value()
		if err != nil {
			t.Fatalf("%s: Value unexpectedly "+
				"failed: %s", testName, err)
		}
		if !bytes.Equal(cursorValue, entry.value) {
			t.Fatalf("%s: Cursor returned "+
				"wrong value. Want: %s, got: %s", testName, entry.value, cursorValue)
		}
	}

	// The cursor should now be exhausted. Make sure Next now
	// returns false
	hasNext := cursor.Next()
	if hasNext {
		t.Fatalf("%s: cursor unexpectedly "+
			"not done", testName)
	}

	// Rewind the cursor and close it
	cursor.First()
	err := cursor.Close()
	if err != nil {
		t.Fatalf("%s: Close unexpectedly "+
			"failed: %s", testName, err)
	}

	// Call Next on the cursor. This time it should panic
	// because it's closed.
	func() {
		defer recoverFromClosedCursorPanic(t, testName)
		cursor.Next()
	}()
}

func TestCursorFirst(t *testing.T) {
	testForAllDatabaseTypes(t, "TestCursorFirst", testCursorFirst)
}

func testCursorFirst(t *testing.T, db database.Database, testName string) {
	entries := populateDatabaseForTest(t, db, testName)
	cursor := prepareCursorForTest(t, db, testName)

	// Make sure that First returns true when the cursor is not empty
	exists := cursor.First()
	if !exists {
		t.Fatalf("%s: Cursor unexpectedly "+
			"returned false", testName)
	}

	// Make sure that the first key and value are as expected
	firstEntryKey := entries[0].key
	firstCursorKey, err := cursor.Key()
	if err != nil {
		t.Fatalf("%s: Key unexpectedly "+
			"failed: %s", testName, err)
	}
	if !reflect.DeepEqual(firstCursorKey, firstEntryKey) {
		t.Fatalf("%s: Cursor returned "+
			"wrong key. Want: %s, got: %s", testName, firstEntryKey, firstCursorKey)
	}
	firstEntryValue := entries[0].value
	firstCursorValue, err := cursor.Value()
	if err != nil {
		t.Fatalf("%s: Value unexpectedly "+
			"failed: %s", testName, err)
	}
	if !bytes.Equal(firstCursorValue, firstEntryValue) {
		t.Fatalf("%s: Cursor returned "+
			"wrong value. Want: %s, got: %s", testName, firstEntryValue, firstCursorValue)
	}

	// Exhaust the cursor
	for cursor.Next() {
		// Do nothing
	}

	// Call first again and make sure it still returns true
	exists = cursor.First()
	if !exists {
		t.Fatalf("%s: First unexpectedly "+
			"returned false", testName)
	}

	// Call next and make sure it returns true as well
	exists = cursor.Next()
	if !exists {
		t.Fatalf("%s: Next unexpectedly "+
			"returned false", testName)
	}

	// Remove all the entries from the database
	for _, entry := range entries {
		err := db.Delete(entry.key)
		if err != nil {
			t.Fatalf("%s: Delete unexpectedly "+
				"failed: %s", testName, err)
		}
	}

	// Create a new cursor over an empty dataset
	cursor = prepareCursorForTest(t, db, testName)

	// Make sure that First returns false when the cursor is empty
	exists = cursor.First()
	if exists {
		t.Fatalf("%s: Cursor unexpectedly "+
			"returned true", testName)
	}
}

func TestCursorSeek(t *testing.T) {
	testForAllDatabaseTypes(t, "TestCursorSeek", testCursorSeek)
}

func testCursorSeek(t *testing.T, db database.Database, testName string) {
	entries := populateDatabaseForTest(t, db, testName)
	cursor := prepareCursorForTest(t, db, testName)

	// Seek to the fourth entry and make sure it exists
	fourthEntry := entries[3]
	err := cursor.Seek(fourthEntry.key)
	if err != nil {
		t.Fatalf("%s: Cursor unexpectedly "+
			"failed: %s", testName, err)
	}

	// Make sure that the key and value are as expected
	fourthEntryKey := entries[3].key
	fourthCursorKey, err := cursor.Key()
	if err != nil {
		t.Fatalf("%s: Key unexpectedly "+
			"failed: %s", testName, err)
	}
	if !reflect.DeepEqual(fourthCursorKey, fourthEntryKey) {
		t.Fatalf("%s: Cursor returned "+
			"wrong key. Want: %s, got: %s", testName, fourthEntryKey, fourthCursorKey)
	}
	fourthEntryValue := entries[3].value
	fourthCursorValue, err := cursor.Value()
	if err != nil {
		t.Fatalf("%s: Value unexpectedly "+
			"failed: %s", testName, err)
	}
	if !bytes.Equal(fourthCursorValue, fourthEntryValue) {
		t.Fatalf("%s: Cursor returned "+
			"wrong value. Want: %s, got: %s", testName, fourthEntryValue, fourthCursorValue)
	}

	// Call Next and make sure that we are now on the fifth entry
	exists := cursor.Next()
	if !exists {
		t.Fatalf("%s: Next unexpectedly "+
			"returned false", testName)
	}
	fifthEntryKey := entries[4].key
	fifthCursorKey, err := cursor.Key()
	if err != nil {
		t.Fatalf("%s: Key unexpectedly "+
			"failed: %s", testName, err)
	}
	if !reflect.DeepEqual(fifthCursorKey, fifthEntryKey) {
		t.Fatalf("%s: Cursor returned "+
			"wrong key. Want: %s, got: %s", testName, fifthEntryKey, fifthCursorKey)
	}
	fifthEntryValue := entries[4].value
	fifthCursorValue, err := cursor.Value()
	if err != nil {
		t.Fatalf("%s: Value unexpectedly "+
			"failed: %s", testName, err)
	}
	if !bytes.Equal(fifthCursorValue, fifthEntryValue) {
		t.Fatalf("%s: Cursor returned "+
			"wrong value. Want: %s, got: %s", testName, fifthEntryValue, fifthCursorValue)
	}

	// Seek to a value that doesn't exist and make sure that
	// the returned error is ErrNotFound
	err = cursor.Seek(database.MakeBucket(nil).Key([]byte("doesn't exist")))
	if err == nil {
		t.Fatalf("%s: Seek unexpectedly "+
			"succeeded", testName)
	}
	if !database.IsNotFoundError(err) {
		t.Fatalf("%s: Seek returned "+
			"wrong error: %s", testName, err)
	}
}

func TestCursorCloseErrors(t *testing.T) {
	testForAllDatabaseTypes(t, "TestCursorCloseErrors", testCursorCloseErrors)
}

func testCursorCloseErrors(t *testing.T, db database.Database, testName string) {
	populateDatabaseForTest(t, db, testName)
	cursor := prepareCursorForTest(t, db, testName)

	// Close the cursor
	err := cursor.Close()
	if err != nil {
		t.Fatalf("%s: Close "+
			"unexpectedly failed: %s", testName, err)
	}

	tests := []struct {
		name     string
		function func() error
	}{
		{
			name: "Seek",
			function: func() error {
				return cursor.Seek(database.MakeBucket(nil).Key([]byte{}))
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
			t.Fatalf("%s: %s "+
				"unexpectedly succeeded", testName, test.name)
		}
		if !strings.Contains(err.Error(), expectedErrContainsString) {
			t.Fatalf("%s: %s "+
				"returned wrong error. Want: %s, got: %s",
				testName, test.name, expectedErrContainsString, err)
		}
	}
}

func TestCursorCloseFirstAndNext(t *testing.T) {
	testForAllDatabaseTypes(t, "TestCursorCloseFirstAndNext", testCursorCloseFirstAndNext)
}

func testCursorCloseFirstAndNext(t *testing.T, db database.Database, testName string) {
	populateDatabaseForTest(t, db, testName)
	cursor := prepareCursorForTest(t, db, testName)

	// Close the cursor
	err := cursor.Close()
	if err != nil {
		t.Fatalf("%s: Close "+
			"unexpectedly failed: %s", testName, err)
	}

	// We expect First to panic
	func() {
		defer recoverFromClosedCursorPanic(t, testName)
		cursor.First()
	}()

	// We expect Next to panic
	func() {
		defer recoverFromClosedCursorPanic(t, testName)
		cursor.Next()
	}()
}
