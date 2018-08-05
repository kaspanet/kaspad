package ffldb

import (
	"bytes"
	"testing"

	"github.com/bouk/monkey"
	"github.com/daglabs/btcd/database"
)

// TestCursorDeleteErrors tests all error-cases in *cursor.Delete().
// The non-error-cases are tested in the more general tests.
func TestCursorDeleteErrors(t *testing.T) {
	pdb := newTestDb("TestCursorDeleteErrors", t)

	nestedBucket := []byte("nestedBucket")
	key := []byte("key")
	value := []byte("value")

	err := pdb.Update(func(tx database.Tx) error {
		metadata := tx.Metadata()
		_, err := metadata.CreateBucket(nestedBucket)
		if err != nil {
			return err
		}
		metadata.Put(key, value)
		return nil
	})
	if err != nil {
		t.Fatalf("TestCursorDeleteErrors: Error setting up test-dabase")
	}

	// Check for error when attempted to delete a bucket
	err = pdb.Update(func(tx database.Tx) error {
		cursor := tx.Metadata().Cursor()
		found := false
		for ok := cursor.First(); ok; ok = cursor.Next() {
			if bytes.Equal(cursor.Key(), nestedBucket) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("TestCursorDeleteErrors: Key '%s' not found", string(nestedBucket))
		}

		err := cursor.Delete()
		if !database.IsErrorCode(err, database.ErrIncompatibleValue) {
			t.Errorf("TestCursorDeleteErrors: Expected error of type ErrIncompatibleValue, "+
				"when deleting bucket, but got %v", err)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("TestCursorDeleteErrors: Unexpected error from pdb.Update "+
			"when attempting to delete bucket: %s", err)
	}

	// Check for error when transaction is not writable
	err = pdb.View(func(tx database.Tx) error {
		cursor := tx.Metadata().Cursor()
		if !cursor.First() {
			t.Fatal("TestCursorDeleteErrors: Nothing in cursor when testing for delete in " +
				"non-writable transaction")
		}

		err := cursor.Delete()
		if !database.IsErrorCode(err, database.ErrTxNotWritable) {
			t.Errorf("TestCursorDeleteErrors: Expected error of type ErrTxNotWritable "+
				"when calling .Delete() on non-writable transaction, but got '%v' instead", err)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("TestCursorDeleteErrors: Unexpected error from pdb.Update "+
			"when attempting to delete on non-writable transaction: %s", err)
	}

	// Check for error when cursor was exhausted
	err = pdb.Update(func(tx database.Tx) error {
		cursor := tx.Metadata().Cursor()
		for ok := cursor.First(); ok; ok = cursor.Next() {
		}

		err := cursor.Delete()
		if !database.IsErrorCode(err, database.ErrIncompatibleValue) {
			t.Errorf("TestCursorDeleteErrors: Expected error of type ErrIncompatibleValue "+
				"when calling .Delete() on exhausted cursor, but got '%v' instead", err)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("TestCursorDeleteErrors: Unexpected error from pdb.Update "+
			"when attempting to delete on exhausted cursor: %s", err)
	}

	// Check for error when transaction is closed
	tx, err := pdb.Begin(true)
	if err != nil {
		t.Fatalf("TestCursorDeleteErrors: Error in pdb.Begin(): %s", err)
	}
	cursor := tx.Metadata().Cursor()
	err = tx.Commit()
	if err != nil {
		t.Fatalf("TestCursorDeleteErrors: Error in tx.Commit(): %s", err)
	}

	err = cursor.Delete()
	if !database.IsErrorCode(err, database.ErrTxClosed) {
		t.Errorf("TestCursorDeleteErrors: Expected error of type ErrTxClosed "+
			"when calling .Delete() on with closed transaction, but got '%s' instead", err)
	}
}

func TestSkipPendingUpdates(t *testing.T) {
	pdb := newTestDb("TestSkipPendingUpdates", t)
	defer pdb.Close()

	value := []byte("value")
	// Add numbered prefixes to keys so that they are in expected order, and before any other keys
	firstKey := []byte("1 - first")
	toDeleteKey := []byte("2 - toDelete")
	toUpdateKey := []byte("3 - toUpdate")
	secondKey := []byte("4 - second")

	// create initial metadata for test
	err := pdb.Update(func(tx database.Tx) error {
		metadata := tx.Metadata()
		if err := metadata.Put(firstKey, value); err != nil {
			return err
		}
		if err := metadata.Put(toDeleteKey, value); err != nil {
			return err
		}
		if err := metadata.Put(toUpdateKey, value); err != nil {
			return err
		}
		if err := metadata.Put(secondKey, value); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		t.Fatalf("Error adding to metadata: %s", err)
	}

	// test skips
	err = pdb.Update(func(tx database.Tx) error {
		metadata := tx.Metadata()
		if err := metadata.Delete(toDeleteKey); err != nil {
			return err
		}
		if err := metadata.Put(toUpdateKey, value); err != nil {
			return err
		}
		cursor := metadata.Cursor().(*cursor)
		dbIter := cursor.dbIter

		// Check that first is ok
		dbIter.First()
		expectedKey := bucketizedKey(metadataBucketID, firstKey)
		if !bytes.Equal(dbIter.Key(), expectedKey) {
			t.Errorf("1: key expected to be %v but is %v", expectedKey, dbIter.Key())
		}

		// Go to the next key, which is toDelete
		dbIter.Next()
		expectedKey = bucketizedKey(metadataBucketID, toDeleteKey)
		if !bytes.Equal(dbIter.Key(), expectedKey) {
			t.Errorf("2: key expected to be %s but is %s", expectedKey, dbIter.Key())
		}

		// at this point toDeleteKey and toUpdateKey should be skipped
		cursor.skipPendingUpdates(true)
		expectedKey = bucketizedKey(metadataBucketID, secondKey)
		if !bytes.Equal(dbIter.Key(), expectedKey) {
			t.Errorf("3: key expected to be %s but is %s", expectedKey, dbIter.Key())
		}

		// now traverse backwards - should get toUpdate
		dbIter.Prev()
		expectedKey = bucketizedKey(metadataBucketID, toUpdateKey)
		if !bytes.Equal(dbIter.Key(), expectedKey) {
			t.Errorf("4: key expected to be %s but is %s", expectedKey, dbIter.Key())
		}

		// at this point toUpdateKey and toDeleteKey should be skipped
		cursor.skipPendingUpdates(false)
		expectedKey = bucketizedKey(metadataBucketID, firstKey)
		if !bytes.Equal(dbIter.Key(), expectedKey) {
			t.Errorf("5: key expected to be %s but is %s", expectedKey, dbIter.Key())
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Error running main part of test: %s", err)
	}
}

// TestCursor tests various edge-cases in cursor that were not hit by the more general tests
func TestCursor(t *testing.T) {
	pdb := newTestDb("TestCursor", t)
	defer pdb.Close()

	value := []byte("value")
	// Add numbered prefixes to keys so that they are in expected order, and before any other keys
	firstKey := []byte("1 - first")
	toDeleteKey := []byte("2 - toDelete")
	toUpdateKey := []byte("3 - toUpdate")
	secondKey := []byte("4 - second")

	// create initial metadata for test
	err := pdb.Update(func(tx database.Tx) error {
		metadata := tx.Metadata()
		if err := metadata.Put(firstKey, value); err != nil {
			return err
		}
		if err := metadata.Put(toDeleteKey, value); err != nil {
			return err
		}
		if err := metadata.Put(toUpdateKey, value); err != nil {
			return err
		}
		if err := metadata.Put(secondKey, value); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		t.Fatalf("Error adding to metadata: %s", err)
	}

	// run the actual tests
	err = pdb.Update(func(tx database.Tx) error {
		metadata := tx.Metadata()
		if err := metadata.Delete(toDeleteKey); err != nil {
			return err
		}
		if err := metadata.Put(toUpdateKey, value); err != nil {
			return err
		}
		cursor := metadata.Cursor().(*cursor)

		// Check prev when currentIter == nil
		if ok := cursor.Prev(); ok {
			t.Error("1: .Prev() should have returned false, but have returned true")
		}
		// Same thing for .Next()
		for ok := cursor.First(); ok; ok = cursor.Next() {
		}
		if ok := cursor.Next(); ok {
			t.Error("2: .Next() should have returned false, but have returned true")
		}

		// Check that Key(), rawKey(), Value(), and rawValue() all return nil when currentIter == nil
		if key := cursor.Key(); key != nil {
			t.Errorf("3: .Key() should have returned nil, but have returned '%s' instead", key)
		}
		if key := cursor.rawKey(); key != nil {
			t.Errorf("4: .rawKey() should have returned nil, but have returned '%s' instead", key)
		}
		if value := cursor.Value(); value != nil {
			t.Errorf("5: .Value() should have returned nil, but have returned '%s' instead", value)
		}
		if value := cursor.rawValue(); value != nil {
			t.Errorf("6: .rawValue() should have returned nil, but have returned '%s' instead", value)
		}

		// Check rawValue in normal operation
		cursor.First()
		if rawValue := cursor.rawValue(); !bytes.Equal(rawValue, value) {
			t.Errorf("7: rawValue should have returned '%s' but have returned '%s' instead", value, rawValue)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("Error running the actual tests: %s", err)
	}
}

// TestCreateBucketErrors tests all error-cases in *bucket.CreateBucket().
// The non-error-cases are tested in the more general tests.
func TestCreateBucketErrors(t *testing.T) {
	testKey := []byte("key")

	tests := []struct {
		name        string
		key         []byte
		target      interface{}
		replacement interface{}
		isWritable  bool
		isClosed    bool
		expectedErr database.ErrorCode
	}{
		{"empty key", []byte{}, nil, nil, true, false, database.ErrBucketNameRequired},
		{"transaction is closed", testKey, nil, nil, true, true, database.ErrTxClosed},
		{"transaction is not writable", testKey, nil, nil, false, false, database.ErrTxNotWritable},
		{"key already exists", []byte("ffldb-blockidx"), nil, nil, true, false, database.ErrBucketExists},
		{"nextBucketID error", testKey, (*transaction).nextBucketID,
			func(*transaction) ([4]byte, error) {
				return [4]byte{}, makeDbErr(database.ErrTxClosed, "error in newBucketID", nil)
			},
			true, false, database.ErrTxClosed},
	}

	for _, test := range tests {
		func() {
			pdb := newTestDb("TestCursor", t)
			defer pdb.Close()

			if test.target != nil && test.replacement != nil {
				patch := monkey.Patch(test.target, test.replacement)
				defer patch.Unpatch()
			}

			tx, err := pdb.Begin(test.isWritable)
			defer tx.Commit()
			if err != nil {
				t.Fatalf("TestCreateBucketErrors: %s: error from pdb.Begin: %s", test.name, err)
			}
			if test.isClosed {
				err = tx.Commit()
				if err != nil {
					t.Fatalf("TestCreateBucketErrors: %s: error from tx.Commit: %s", test.name, err)
				}
			}

			metadata := tx.Metadata()

			_, err = metadata.CreateBucket(test.key)

			if !database.IsErrorCode(err, test.expectedErr) {
				t.Errorf("TestCreateBucketErrors: %s: Expected error of type %d "+
					"but got '%v'", test.name, test.expectedErr, err)
			}

		}()
	}
}
