package ldb

import (
	"bytes"
	"fmt"
	"github.com/kaspanet/kaspad/database"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

func TestCursorSanity(t *testing.T) {
	// Open a test db
	path, err := ioutil.TempDir("", "TestCursorSanity")
	if err != nil {
		t.Fatalf("TestCursorSanity: TempDir unexpectedly "+
			"failed: %s", err)
	}
	db, err := NewLevelDB(path)
	if err != nil {
		t.Fatalf("TestCursorSanity: Open "+
			"unexpectedly failed: %s", err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			t.Fatalf("TestCursorSanity: Close "+
				"unexpectedly failed: %s", err)
		}
	}()

	// Write some data to the database
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		err := db.Put(database.MakeBucket([]byte("bucket")).Key([]byte(key)), []byte(value))
		if err != nil {
			t.Fatalf("TestCursorSanity: Put "+
				"unexpectedly failed: %s", err)
		}
	}

	// Open a new cursor
	cursor := db.Cursor(database.MakeBucket([]byte("bucket")))
	defer func() {
		err = cursor.Close()
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
	firstKey, err := cursor.Key()
	if err != nil {
		t.Fatalf("TestCursorSanity: Key "+
			"unexpectedly failed: %s", err)
	}
	expectedKey := database.MakeBucket([]byte("bucket")).Key([]byte("key0"))
	if !reflect.DeepEqual(firstKey, expectedKey) {
		t.Fatalf("TestCursorSanity: Key "+
			"returned wrong key. Want: %s, got: %s",
			string(expectedKey.Bytes()), string(firstKey.Bytes()))
	}
	firstValue, err := cursor.Value()
	if err != nil {
		t.Fatalf("TestCursorSanity: Value "+
			"unexpectedly failed: %s", err)
	}
	expectedValue := []byte("value0")
	if !bytes.Equal(firstValue, expectedValue) {
		t.Fatalf("TestCursorSanity: Value "+
			"returned wrong value. Want: %s, got: %s",
			string(expectedValue), string(firstValue))
	}

	// Seek to a non-existant key
	err = cursor.Seek(database.MakeBucket().Key([]byte("doesn't exist")))
	if err == nil {
		t.Fatalf("TestCursorSanity: Seek " +
			"unexpectedly succeeded")
	}
	if !database.IsNotFoundError(err) {
		t.Fatalf("TestCursorSanity: Seek "+
			"returned wrong error: %s", err)
	}

	// Seek to the last key
	err = cursor.Seek(database.MakeBucket([]byte("bucket")).Key([]byte("key9")))
	if err != nil {
		t.Fatalf("TestCursorSanity: Seek "+
			"unexpectedly failed: %s", err)
	}
	lastKey, err := cursor.Key()
	if err != nil {
		t.Fatalf("TestCursorSanity: Key "+
			"unexpectedly failed: %s", err)
	}
	expectedKey = database.MakeBucket([]byte("bucket")).Key([]byte("key9"))
	if !reflect.DeepEqual(lastKey, expectedKey) {
		t.Fatalf("TestCursorSanity: Key "+
			"returned wrong key. Want: %s, got: %s",
			expectedKey, lastKey)
	}
	lastValue, err := cursor.Value()
	if err != nil {
		t.Fatalf("TestCursorSanity: Value "+
			"unexpectedly failed: %s", err)
	}
	expectedValue = []byte("value9")
	if !bytes.Equal(lastValue, expectedValue) {
		t.Fatalf("TestCursorSanity: Value "+
			"returned wrong value. Want: %s, got: %s",
			string(expectedValue), string(lastValue))
	}

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
		name     string
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
			// Open a test db
			path, err := ioutil.TempDir("", "TestCursorCloseErrors")
			if err != nil {
				t.Fatalf("TestCursorCloseErrors: TempDir unexpectedly "+
					"failed: %s", err)
			}
			db, err := NewLevelDB(path)
			if err != nil {
				t.Fatalf("TestCursorCloseErrors: Open "+
					"unexpectedly failed: %s", err)
			}
			defer func() {
				err := db.Close()
				if err != nil {
					t.Fatalf("TestCursorCloseErrors: Close "+
						"unexpectedly failed: %s", err)
				}
			}()

			// Open a new cursor
			cursor := db.Cursor(database.MakeBucket())

			// Close the cursor
			err = cursor.Close()
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
	// Open a test db
	path, err := ioutil.TempDir("", "TestCursorCloseFirstAndNext")
	if err != nil {
		t.Fatalf("TestCursorCloseFirstAndNext: TempDir unexpectedly "+
			"failed: %s", err)
	}
	db, err := NewLevelDB(path)
	if err != nil {
		t.Fatalf("TestCursorCloseFirstAndNext: Open "+
			"unexpectedly failed: %s", err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			t.Fatalf("TestCursorCloseFirstAndNext: Close "+
				"unexpectedly failed: %s", err)
		}
	}()

	// Write some data to the database
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		err := db.Put(database.MakeBucket([]byte("bucket")).Key([]byte(key)), []byte(value))
		if err != nil {
			t.Fatalf("TestCursorCloseFirstAndNext: Put "+
				"unexpectedly failed: %s", err)
		}
	}

	// Open a new cursor
	cursor := db.Cursor(database.MakeBucket([]byte("bucket")))

	// Close the cursor
	err = cursor.Close()
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
