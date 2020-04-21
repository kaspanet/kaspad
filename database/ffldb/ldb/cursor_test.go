package ldb

import (
	"bytes"
	"fmt"
	"github.com/kaspanet/kaspad/database"
	"reflect"
	"strings"
	"testing"
)

func validateCurrentCursorKeyAndValue(t *testing.T, testName string, cursor *LevelDBCursor,
	expectedKey *database.Key, expectedValue []byte) {

	cursorKey, err := cursor.Key()
	if err != nil {
		t.Fatalf("%s: Key "+
			"unexpectedly failed: %s", testName, err)
	}
	if !reflect.DeepEqual(cursorKey, expectedKey) {
		t.Fatalf("%s: Key "+
			"returned wrong key. Want: %s, got: %s",
			testName, string(expectedKey.Bytes()), string(cursorKey.Bytes()))
	}
	cursorValue, err := cursor.Value()
	if err != nil {
		t.Fatalf("%s: Value "+
			"unexpectedly failed: %s", testName, err)
	}
	if !bytes.Equal(cursorValue, expectedValue) {
		t.Fatalf("%s: Value "+
			"returned wrong value. Want: %s, got: %s",
			testName, string(expectedValue), string(cursorValue))
	}
}

// TestCursorSanity validates typical cursor usage, including
// opening a cursor over some existing data, seeking back
// and forth over that data, and getting some keys/values out
// of the cursor.
func TestCursorSanity(t *testing.T) {
	ldb, teardownFunc := prepareDatabaseForTest(t, "TestCursorSanity")
	defer teardownFunc()

	// Write some data to the database
	bucket := database.MakeBucket([]byte("bucket"))
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		err := ldb.Put(bucket.Key([]byte(key)), []byte(value))
		if err != nil {
			t.Fatalf("TestCursorSanity: Put "+
				"unexpectedly failed: %s", err)
		}
	}

	// Open a new cursor
	cursor := ldb.Cursor(bucket)
	defer func() {
		err := cursor.Close()
		if err != nil {
			t.Fatalf("TestCursorSanity: Close "+
				"unexpectedly failed: %s", err)
		}
	}()

	// Seek to first key and make sure its key and value are correct
	hasNext := cursor.First()
	if !hasNext {
		t.Fatalf("TestCursorSanity: First " +
			"unexpectedly returned non-existance")
	}
	expectedKey := bucket.Key([]byte("key0"))
	expectedValue := []byte("value0")
	validateCurrentCursorKeyAndValue(t, "TestCursorSanity", cursor, expectedKey, expectedValue)

	// Seek to a non-existant key
	err := cursor.Seek(database.MakeBucket().Key([]byte("doesn't exist")))
	if err == nil {
		t.Fatalf("TestCursorSanity: Seek " +
			"unexpectedly succeeded")
	}
	if !database.IsNotFoundError(err) {
		t.Fatalf("TestCursorSanity: Seek "+
			"returned wrong error: %s", err)
	}

	// Seek to the last key
	err = cursor.Seek(bucket.Key([]byte("key9")))
	if err != nil {
		t.Fatalf("TestCursorSanity: Seek "+
			"unexpectedly failed: %s", err)
	}
	expectedKey = bucket.Key([]byte("key9"))
	expectedValue = []byte("value9")
	validateCurrentCursorKeyAndValue(t, "TestCursorSanity", cursor, expectedKey, expectedValue)

	// Call Next to get to the end of the cursor. This should
	// return false to signify that there are no items after that.
	// Key and Value calls should return ErrNotFound.
	hasNext = cursor.Next()
	if hasNext {
		t.Fatalf("TestCursorSanity: Next " +
			"after last value is unexpectedly not done")
	}
	_, err = cursor.Key()
	if err == nil {
		t.Fatalf("TestCursorSanity: Key " +
			"unexpectedly succeeded")
	}
	if !database.IsNotFoundError(err) {
		t.Fatalf("TestCursorSanity: Key "+
			"returned wrong error: %s", err)
	}
	_, err = cursor.Value()
	if err == nil {
		t.Fatalf("TestCursorSanity: Value " +
			"unexpectedly succeeded")
	}
	if !database.IsNotFoundError(err) {
		t.Fatalf("TestCursorSanity: Value "+
			"returned wrong error: %s", err)
	}
}

func TestCursorCloseErrors(t *testing.T) {
	tests := []struct {
		name string

		// function is the LevelDBCursor function that we're
		// verifying returns an error after the cursor had
		// been closed.
		function func(dbTx *LevelDBCursor) error
	}{
		{
			name: "Seek",
			function: func(cursor *LevelDBCursor) error {
				return cursor.Seek(database.MakeBucket().Key([]byte{}))
			},
		},
		{
			name: "Key",
			function: func(cursor *LevelDBCursor) error {
				_, err := cursor.Key()
				return err
			},
		},
		{
			name: "Value",
			function: func(cursor *LevelDBCursor) error {
				_, err := cursor.Value()
				return err
			},
		},
		{
			name: "Close",
			function: func(cursor *LevelDBCursor) error {
				return cursor.Close()
			},
		},
	}

	for _, test := range tests {
		func() {
			ldb, teardownFunc := prepareDatabaseForTest(t, "TestCursorCloseErrors")
			defer teardownFunc()

			// Open a new cursor
			cursor := ldb.Cursor(database.MakeBucket())

			// Close the cursor
			err := cursor.Close()
			if err != nil {
				t.Fatalf("TestCursorCloseErrors: Close "+
					"unexpectedly failed: %s", err)
			}

			expectedErrContainsString := "closed cursor"

			// Make sure that the test function returns a "closed transaction" error
			err = test.function(cursor)
			if err == nil {
				t.Fatalf("TestCursorCloseErrors: %s "+
					"unexpectedly succeeded", test.name)
			}
			if !strings.Contains(err.Error(), expectedErrContainsString) {
				t.Fatalf("TestCursorCloseErrors: %s "+
					"returned wrong error. Want: %s, got: %s",
					test.name, expectedErrContainsString, err)
			}
		}()
	}
}

func TestCursorCloseFirstAndNext(t *testing.T) {
	ldb, teardownFunc := prepareDatabaseForTest(t, "TestCursorCloseFirstAndNext")
	defer teardownFunc()

	// Write some data to the database
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		err := ldb.Put(database.MakeBucket([]byte("bucket")).Key([]byte(key)), []byte(value))
		if err != nil {
			t.Fatalf("TestCursorCloseFirstAndNext: Put "+
				"unexpectedly failed: %s", err)
		}
	}

	// Open a new cursor
	cursor := ldb.Cursor(database.MakeBucket([]byte("bucket")))

	// Close the cursor
	err := cursor.Close()
	if err != nil {
		t.Fatalf("TestCursorCloseFirstAndNext: Close "+
			"unexpectedly failed: %s", err)
	}

	result := cursor.First()
	if result {
		t.Fatalf("TestCursorCloseFirstAndNext: First " +
			"unexpectedly returned true")
	}

	result = cursor.Next()
	if result {
		t.Fatalf("TestCursorCloseFirstAndNext: Next " +
			"unexpectedly returned true")
	}
}
