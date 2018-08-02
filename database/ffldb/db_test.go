package ffldb

import (
	"bytes"
	"testing"

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
