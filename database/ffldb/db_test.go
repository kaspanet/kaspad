package ffldb

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"bou.ke/monkey"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/wire"
)

// TestCursorDeleteErrors tests all error-cases in *cursor.Delete().
// The non-error-cases are tested in the more general tests.
func TestCursorDeleteErrors(t *testing.T) {
	pdb := newTestDb("TestCursorDeleteErrors", t)

	nestedBucket := []byte("nestedBucket")
	key := []byte("key")
	value := []byte("value")

	err := pdb.Update(func(dbTx database.Tx) error {
		metadata := dbTx.Metadata()
		_, err := metadata.CreateBucket(nestedBucket)
		if err != nil {
			return err
		}
		metadata.Put(key, value)
		return nil
	})
	if err != nil {
		t.Fatalf("TestCursorDeleteErrors: Error setting up test-database: %s", err)
	}

	// Check for error when attempted to delete a bucket
	err = pdb.Update(func(dbTx database.Tx) error {
		cursor := dbTx.Metadata().Cursor()
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
	err = pdb.View(func(dbTx database.Tx) error {
		cursor := dbTx.Metadata().Cursor()
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
	err = pdb.Update(func(dbTx database.Tx) error {
		cursor := dbTx.Metadata().Cursor()
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
	err := pdb.Update(func(dbTx database.Tx) error {
		metadata := dbTx.Metadata()
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
		t.Fatalf("TestSkipPendingUpdates: Error adding to metadata: %s", err)
	}

	// test skips
	err = pdb.Update(func(dbTx database.Tx) error {
		metadata := dbTx.Metadata()
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
			t.Errorf("TestSkipPendingUpdates: 1: key expected to be %v but is %v", expectedKey, dbIter.Key())
		}

		// Go to the next key, which is toDelete
		dbIter.Next()
		expectedKey = bucketizedKey(metadataBucketID, toDeleteKey)
		if !bytes.Equal(dbIter.Key(), expectedKey) {
			t.Errorf("TestSkipPendingUpdates: 2: key expected to be %s but is %s", expectedKey, dbIter.Key())
		}

		// at this point toDeleteKey and toUpdateKey should be skipped
		cursor.skipPendingUpdates(true)
		expectedKey = bucketizedKey(metadataBucketID, secondKey)
		if !bytes.Equal(dbIter.Key(), expectedKey) {
			t.Errorf("TestSkipPendingUpdates: 3: key expected to be %s but is %s", expectedKey, dbIter.Key())
		}

		// now traverse backwards - should get toUpdate
		dbIter.Prev()
		expectedKey = bucketizedKey(metadataBucketID, toUpdateKey)
		if !bytes.Equal(dbIter.Key(), expectedKey) {
			t.Errorf("TestSkipPendingUpdates: 4: key expected to be %s but is %s", expectedKey, dbIter.Key())
		}

		// at this point toUpdateKey and toDeleteKey should be skipped
		cursor.skipPendingUpdates(false)
		expectedKey = bucketizedKey(metadataBucketID, firstKey)
		if !bytes.Equal(dbIter.Key(), expectedKey) {
			t.Errorf("TestSkipPendingUpdates: 5: key expected to be %s but is %s", expectedKey, dbIter.Key())
		}
		return nil
	})
	if err != nil {
		t.Fatalf("TestSkipPendingUpdates: Error running main part of test: %s", err)
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
	err := pdb.Update(func(dbTx database.Tx) error {
		metadata := dbTx.Metadata()
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
	err = pdb.Update(func(dbTx database.Tx) error {
		metadata := dbTx.Metadata()
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
		{"key already exists", blockIdxBucketName, nil, nil, true, false, database.ErrBucketExists},
		{"nextBucketID error", testKey, (*transaction).nextBucketID,
			func(*transaction) ([4]byte, error) {
				return [4]byte{}, makeDbErr(database.ErrTxClosed, "error in newBucketID", nil)
			},
			true, false, database.ErrTxClosed},
	}

	for _, test := range tests {
		func() {
			pdb := newTestDb("TestCreateBucketErrors", t)
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

// TestPutErrors tests all error-cases in *bucket.Put().
// The non-error-cases are tested in the more general tests.
func TestPutErrors(t *testing.T) {
	testKey := []byte("key")
	testValue := []byte("value")

	tests := []struct {
		name        string
		key         []byte
		isWritable  bool
		isClosed    bool
		expectedErr database.ErrorCode
	}{
		{"empty key", []byte{}, true, false, database.ErrKeyRequired},
		{"transaction is closed", testKey, true, true, database.ErrTxClosed},
		{"transaction is not writable", testKey, false, false, database.ErrTxNotWritable},
	}

	for _, test := range tests {
		func() {
			pdb := newTestDb("TestPutErrors", t)
			defer pdb.Close()

			tx, err := pdb.Begin(test.isWritable)
			defer tx.Commit()
			if err != nil {
				t.Fatalf("TestPutErrors: %s: error from pdb.Begin: %s", test.name, err)
			}
			if test.isClosed {
				err = tx.Commit()
				if err != nil {
					t.Fatalf("TestPutErrors: %s: error from tx.Commit: %s", test.name, err)
				}
			}

			metadata := tx.Metadata()

			err = metadata.Put(test.key, testValue)

			if !database.IsErrorCode(err, test.expectedErr) {
				t.Errorf("TestPutErrors: %s: Expected error of type %d "+
					"but got '%v'", test.name, test.expectedErr, err)
			}

		}()
	}
}

// TestGetErrors tests all error-cases in *bucket.Get().
// The non-error-cases are tested in the more general tests.
func TestGetErrors(t *testing.T) {
	testKey := []byte("key")

	tests := []struct {
		name     string
		key      []byte
		isClosed bool
	}{
		{"empty key", []byte{}, false},
		{"transaction is closed", testKey, true},
	}

	for _, test := range tests {
		func() {
			pdb := newTestDb("TestGetErrors", t)
			defer pdb.Close()

			tx, err := pdb.Begin(false)
			defer tx.Rollback()
			if err != nil {
				t.Fatalf("TestGetErrors: %s: error from pdb.Begin: %s", test.name, err)
			}
			if test.isClosed {
				err = tx.Rollback()
				if err != nil {
					t.Fatalf("TestGetErrors: %s: error from tx.Commit: %s", test.name, err)
				}
			}

			metadata := tx.Metadata()

			if result := metadata.Get(test.key); result != nil {
				t.Errorf("TestGetErrors: %s: Expected to return nil, but got %v", test.name, result)
			}
		}()
	}
}

// TestDeleteErrors tests all error-cases in *bucket.Delete().
// The non-error-cases are tested in the more general tests.
func TestDeleteErrors(t *testing.T) {
	testKey := []byte("key")

	tests := []struct {
		name        string
		key         []byte
		isWritable  bool
		isClosed    bool
		expectedErr database.ErrorCode
	}{
		{"empty key", []byte{}, true, false, database.ErrKeyRequired},
		{"transaction is closed", testKey, true, true, database.ErrTxClosed},
		{"transaction is not writable", testKey, false, false, database.ErrTxNotWritable},
	}

	for _, test := range tests {
		func() {
			pdb := newTestDb("TestDeleteErrors", t)
			defer pdb.Close()

			tx, err := pdb.Begin(test.isWritable)
			defer tx.Commit()
			if err != nil {
				t.Fatalf("TestDeleteErrors: %s: error from pdb.Begin: %s", test.name, err)
			}
			if test.isClosed {
				err = tx.Commit()
				if err != nil {
					t.Fatalf("TestDeleteErrors: %s: error from tx.Commit: %s", test.name, err)
				}
			}

			metadata := tx.Metadata()

			err = metadata.Delete(test.key)

			if !database.IsErrorCode(err, test.expectedErr) {
				t.Errorf("TestDeleteErrors: %s: Expected error of type %d "+
					"but got '%v'", test.name, test.expectedErr, err)
			}
		}()
	}
}

func TestForEachBucket(t *testing.T) {
	pdb := newTestDb("TestForEachBucket", t)

	// set-up test
	testKey := []byte("key")
	testValue := []byte("value")
	bucketKeys := [][]byte{{1}, {2}, {3}}

	err := pdb.Update(func(dbTx database.Tx) error {
		metadata := dbTx.Metadata()
		for _, bucketKey := range bucketKeys {
			bucket, err := metadata.CreateBucket(bucketKey)
			if err != nil {
				return err
			}

			err = bucket.Put(testKey, testValue)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("TestForEachBucket: Error setting up test-database: %s", err)
	}

	// actual test
	err = pdb.View(func(dbTx database.Tx) error {
		i := 0
		metadata := dbTx.Metadata()

		err := metadata.ForEachBucket(func(bucketKey []byte) error {
			if i >= len(bucketKeys) { // in case there are any other buckets in metadata
				return nil
			}

			expectedBucketKey := bucketKeys[i]
			if !bytes.Equal(expectedBucketKey, bucketKey) {
				t.Errorf("TestForEachBucket: %d: Expected bucket key: %v, but got: %v",
					i, expectedBucketKey, bucketKey)
				return nil
			}
			bucket := metadata.Bucket(bucketKey)
			if bucket == nil {
				t.Errorf("TestForEachBucket: %d: Bucket is nil", i)
				return nil
			}

			value := bucket.Get(testKey)
			if !bytes.Equal(testValue, value) {
				t.Errorf("TestForEachBucket: %d: Expected value: %s, but got: %s",
					i, testValue, value)
				return nil
			}

			i++
			return nil
		})

		return err
	})
	if err != nil {
		t.Fatalf("TestForEachBucket: Error running actual tests: %s", err)
	}
}

// TestStoreBlockErrors tests all error-cases in *tx.StoreBlock().
// The non-error-cases are tested in the more general tests.
func TestStoreBlockErrors(t *testing.T) {
	testBlock := util.NewBlock(wire.NewMsgBlock(wire.NewBlockHeader(1, []daghash.Hash{}, &daghash.Hash{}, &daghash.Hash{}, 0, 0)))

	tests := []struct {
		name        string
		target      interface{}
		replacement interface{}
		isWritable  bool
		isClosed    bool
		expectedErr database.ErrorCode
	}{
		{"transaction is closed", nil, nil, true, true, database.ErrTxClosed},
		{"transaction is not writable", nil, nil, false, false, database.ErrTxNotWritable},
		{"block exists", (*transaction).hasBlock,
			func(*transaction, *daghash.Hash) bool { return true },
			true, false, database.ErrBlockExists},
		{"error in block.Bytes", (*util.Block).Bytes,
			func(*util.Block) ([]byte, error) { return nil, errors.New("Error in block.Bytes()") },
			true, false, database.ErrDriverSpecific},
	}

	for _, test := range tests {
		func() {
			pdb := newTestDb("TestStoreBlockErrors", t)
			defer pdb.Close()

			if test.target != nil && test.replacement != nil {
				patch := monkey.Patch(test.target, test.replacement)
				defer patch.Unpatch()
			}

			tx, err := pdb.Begin(test.isWritable)
			defer tx.Commit()
			if err != nil {
				t.Fatalf("TestStoreBlockErrors: %s: error from pdb.Begin: %s", test.name, err)
			}
			if test.isClosed {
				err = tx.Commit()
				if err != nil {
					t.Fatalf("TestStoreBlockErrors: %s: error from tx.Commit: %s", test.name, err)
				}
			}

			err = tx.StoreBlock(testBlock)
			if !database.IsErrorCode(err, test.expectedErr) {
				t.Errorf("TestStoreBlockErrors: %s: Expected error of type %d "+
					"but got '%v'", test.name, test.expectedErr, err)
			}

		}()
	}
}

// TestDeleteDoubleNestedBucket tests what happens when bucket.DeleteBucket()
// is invoked on a bucket that contains a nested bucket.
func TestDeleteDoubleNestedBucket(t *testing.T) {
	pdb := newTestDb("TestDeleteDoubleNestedBucket", t)
	defer pdb.Close()

	firstKey := []byte("first")
	secondKey := []byte("second")
	key := []byte("key")
	value := []byte("value")
	var rawKey, rawSecondKey []byte

	// Test setup
	err := pdb.Update(func(dbTx database.Tx) error {
		metadata := dbTx.Metadata()
		firstBucket, err := metadata.CreateBucket(firstKey)
		if err != nil {
			return fmt.Errorf("Error creating first bucket: %s", err)
		}
		secondBucket, err := firstBucket.CreateBucket(secondKey)
		if err != nil {
			return fmt.Errorf("Error creating second bucket: %s", err)
		}
		secondBucket.Put(key, value)

		// extract rawKey from cursor and make sure it's in raw database
		c := secondBucket.Cursor()
		for ok := c.First(); ok && !bytes.Equal(c.Key(), key); ok = c.Next() {
		}
		if !bytes.Equal(c.Key(), key) {
			return fmt.Errorf("Couldn't find key to extract rawKey")
		}
		rawKey = c.(*cursor).rawKey()
		if dbTx.(*transaction).fetchKey(rawKey) == nil {
			return fmt.Errorf("rawKey not found")
		}

		// extract rawSecondKey from cursor and make sure it's in raw database
		c = firstBucket.Cursor()
		for ok := c.First(); ok && !bytes.Equal(c.Key(), secondKey); ok = c.Next() {
		}
		if !bytes.Equal(c.Key(), secondKey) {
			return fmt.Errorf("Couldn't find secondKey to extract rawSecondKey")
		}
		rawSecondKey = c.(*cursor).rawKey()
		if dbTx.(*transaction).fetchKey(rawSecondKey) == nil {
			return fmt.Errorf("rawSecondKey not found for some reason")
		}

		return nil
	})
	if err != nil {
		t.Fatalf("TestDeleteDoubleNestedBucket: Error in test setup pdb.Update: %s", err)
	}

	// Actual test
	err = pdb.Update(func(dbTx database.Tx) error {
		metadata := dbTx.Metadata()
		err := metadata.DeleteBucket(firstKey)
		if err != nil {
			return err
		}

		if dbTx.(*transaction).fetchKey(rawSecondKey) != nil {
			t.Error("TestDeleteDoubleNestedBucket: secondBucket was not deleted")
		}

		if dbTx.(*transaction).fetchKey(rawKey) != nil {
			t.Error("TestDeleteDoubleNestedBucket: value inside secondBucket was not deleted")
		}

		return nil
	})
	if err != nil {
		t.Fatalf("TestDeleteDoubleNestedBucket: Error in actual test pdb.Update: %s", err)
	}
}

// TestWritePendingAndCommitErrors tests some error-cases in *tx.writePendingAndCommit().
// The non-error-cases are tested in the more general tests.
func TestWritePendingAndCommitErrors(t *testing.T) {
	putPatch := monkey.Patch((*bucket).Put,
		func(_ *bucket, _, _ []byte) error { return errors.New("Error in bucket.Put") })
	defer putPatch.Unpatch()

	rollbackCalled := false
	var rollbackPatch *monkey.PatchGuard
	rollbackPatch = monkey.Patch((*blockStore).handleRollback,
		func(s *blockStore, oldBlockFileNum, oldBlockOffset uint32) {
			rollbackPatch.Unpatch()
			defer rollbackPatch.Restore()

			rollbackCalled = true
			s.handleRollback(oldBlockFileNum, oldBlockOffset)
		})
	defer rollbackPatch.Unpatch()

	pdb := newTestDb("TestWritePendingAndCommitErrors", t)
	defer pdb.Close()

	err := pdb.Update(func(dbTx database.Tx) error { return nil })
	if err == nil {
		t.Errorf("No error returned when metaBucket.Put() should have returned an error")
	}
	if !rollbackCalled {
		t.Errorf("No rollback called when metaBucket.Put() have returned an error")
	}

	rollbackCalled = false
	err = pdb.Update(func(dbTx database.Tx) error {
		return dbTx.StoreBlock(util.NewBlock(wire.NewMsgBlock(
			wire.NewBlockHeader(1, []daghash.Hash{}, &daghash.Hash{}, &daghash.Hash{}, 0, 0))))
	})
	if err == nil {
		t.Errorf("No error returned when blockIdx.Put() should have returned an error")
	}
	if !rollbackCalled {
		t.Errorf("No rollback called when blockIdx.Put() have returned an error")
	}
}
