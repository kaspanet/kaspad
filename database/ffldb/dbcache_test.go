package ffldb

import (
	"bytes"
	"errors"
	"testing"

	"bou.ke/monkey"
	"github.com/btcsuite/goleveldb/leveldb"
	"github.com/btcsuite/goleveldb/leveldb/opt"
	ldbutil "github.com/btcsuite/goleveldb/leveldb/util"
	"github.com/daglabs/btcd/database"
)

// TestDBCacheCloseErrors tests all error-cases in *dbCache.Close().
// The non-error-cases are tested in the more general tests.
func TestDBCacheCloseErrors(t *testing.T) {
	cache := newTestDb("TestDBCacheCloseErrors", t).cache
	defer cache.Close()

	closeCalled := false
	closePatch := monkey.Patch((*leveldb.DB).Close, func(*leveldb.DB) error { closeCalled = true; return nil })
	defer closePatch.Unpatch()

	expectedErr := errors.New("error on flush")

	flushPatch := monkey.Patch((*dbCache).flush, func(*dbCache) error { return expectedErr })
	defer flushPatch.Unpatch()

	err := cache.Close()
	if err != expectedErr {
		t.Errorf("TestDBCacheCloseErrors: Expected error on bad flush is %s but got %s", expectedErr, err)
	}
	if !closeCalled {
		t.Errorf("TestDBCacheCloseErrors: ldb.Close was not called when error flushing")
	}
}

// TestUpdateDBErrors tests all error-cases in *dbCache.UpdateDB().
// The non-error-cases are tested in the more general tests.
func TestUpdateDBErrors(t *testing.T) {
	// Test when ldb.OpenTransaction returns error
	func() {
		cache := newTestDb("TestDBCacheCloseErrors", t).cache
		defer cache.Close()

		patch := monkey.Patch((*leveldb.DB).OpenTransaction,
			func(*leveldb.DB) (*leveldb.Transaction, error) { return nil, errors.New("error in OpenTransaction") })
		defer patch.Unpatch()

		err := cache.updateDB(func(ldbTx *leveldb.Transaction) error { return nil })
		if err == nil {
			t.Errorf("No error in updateDB when ldb.OpenTransaction returns an error")
		}
	}()

	// Test when ldbTx.Commit returns an error
	func() {
		cache := newTestDb("TestDBCacheCloseErrors", t).cache
		defer cache.Close()

		patch := monkey.Patch((*leveldb.Transaction).Commit,
			func(*leveldb.Transaction) error { return errors.New("error in Commit") })
		defer patch.Unpatch()

		err := cache.updateDB(func(ldbTx *leveldb.Transaction) error { return nil })
		if err == nil {
			t.Errorf("No error in updateDB when ldbTx.Commit returns an error")
		}
	}()

	cache := newTestDb("TestDBCacheCloseErrors", t).cache
	defer cache.Close()

	// Test when function passed to updateDB returns an error
	err := cache.updateDB(func(ldbTx *leveldb.Transaction) error { return errors.New("Error in fn") })
	if err == nil {
		t.Errorf("No error in updateDB when passed function returns an error")
	}
}

// TestCommitTxFlushNeeded test the *dbCache.commitTx function when flush is needed,
// including error-cases.
// When flush is not needed is tested in the more general tests.
func TestCommitTxFlushNeeded(t *testing.T) {
	tests := []struct {
		name          string
		target        interface{}
		replacement   interface{}
		expectedError bool
	}{
		{"No errors", nil, nil, false},
		{"Error in flush", (*dbCache).flush, func(*dbCache) error { return errors.New("error") }, true},
		{"Error in commitTreaps", (*dbCache).commitTreaps,
			func(*dbCache, TreapForEacher, TreapForEacher) error { return errors.New("error") }, true},
	}

	for _, test := range tests {
		func() {
			db := newTestDb("TestDBCacheCloseErrors", t)
			defer db.Close()
			cache := db.cache

			cache.flushInterval = 0 // set flushInterval to 0 so that flush is always required

			if test.target != nil && test.replacement != nil {
				patch := monkey.Patch(test.target, test.replacement)
				defer patch.Unpatch()
			}

			tx, err := db.Begin(true)
			if err != nil {
				t.Fatalf("Error begining transaction: %s", err)
			}
			cache.commitTx(tx.(*transaction))
			db.closeLock.RUnlock()
		}()
	}
}

func TestExhaustedDbCacheIterator(t *testing.T) {
	db := newTestDb("TestExhaustedDbCacheIterator", t)
	defer db.Close()

	snapshot, err := db.cache.Snapshot()
	if err != nil {
		t.Fatalf("TestExhaustedDbCacheIterator: Error creating cache snapshot: %s", err)
	}
	iterator := snapshot.NewIterator(&ldbutil.Range{})

	if next := iterator.Next(); next != false {
		t.Errorf("TestExhaustedDbCacheIterator: Expected .Next() = false, but got %v", next)
	}

	if prev := iterator.Prev(); prev != false {
		t.Errorf("TestExhaustedDbCacheIterator: Expected .Prev() = false, but got %v", prev)
	}

	if key := iterator.Key(); key != nil {
		t.Errorf("TestExhaustedDbCacheIterator: Expected .Key() = nil, but got %v", key)
	}

	if value := iterator.Value(); value != nil {
		t.Errorf("TestExhaustedDbCacheIterator: Expected .Value() = nil, but got %v", value)
	}
}

// TestLDBIteratorImplPlaceholders hits functions that are there to implement leveldb iterator.Iterator interface,
// but surve no other purpose.
func TestLDBIteratorImplPlaceholders(t *testing.T) {
	db := newTestDb("TestIteratorImplPlaceholders", t)
	defer db.Close()

	snapshot, err := db.cache.Snapshot()
	if err != nil {
		t.Fatalf("TestLDBIteratorImplPlaceholders: Error creating cache snapshot: %s", err)
	}
	iterator := newLdbCacheIter(snapshot, &ldbutil.Range{})

	if err = iterator.Error(); err != nil {
		t.Errorf("TestLDBIteratorImplPlaceholders: Expected .Error() = nil, but got %v", err)
	}

	// Call SetReleaser to achieve coverage of it. Actually does nothing
	iterator.SetReleaser(nil)
}

func TestSkipPendingUpdatesCache(t *testing.T) {
	pdb := newTestDb("TestSkipPendingUpdatesCache", t)
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

	err = pdb.cache.flush()
	if err != nil {
		t.Fatalf("Error flushing cache: %s", err)
	}

	// test skips
	err = pdb.Update(func(tx database.Tx) error {
		snapshot, err := pdb.cache.Snapshot()
		if err != nil {
			t.Fatalf("TestSkipPendingUpdatesCache: Error getting snapshot: %s", err)
		}

		iterator := snapshot.NewIterator(&ldbutil.Range{})
		snapshot.pendingRemove = snapshot.pendingRemove.Put(bucketizedKey(metadataBucketID, toDeleteKey), value)
		snapshot.pendingKeys = snapshot.pendingKeys.Put(bucketizedKey(metadataBucketID, toUpdateKey), value)

		// Check that first is ok
		iterator.First()
		expectedKey := bucketizedKey(metadataBucketID, firstKey)
		actualKey := iterator.Key()
		if !bytes.Equal(actualKey, expectedKey) {
			t.Errorf("TestSkipPendingUpdatesCache: 1: key expected to be %v but is %v", expectedKey, actualKey)
		}

		// Go to the next key, which is second, toDelete and toUpdate will be skipped
		iterator.Next()
		expectedKey = bucketizedKey(metadataBucketID, secondKey)
		actualKey = iterator.Key()
		if !bytes.Equal(actualKey, expectedKey) {
			t.Errorf("TestSkipPendingUpdatesCache: 2: key expected to be %s but is %s", expectedKey, actualKey)
		}

		// now traverse backwards - should get first, toUpdate and toDelete will be skipped
		iterator.Prev()
		expectedKey = bucketizedKey(metadataBucketID, firstKey)
		actualKey = iterator.Key()
		if !bytes.Equal(actualKey, expectedKey) {
			t.Errorf("TestSkipPendingUpdatesCache: 4: key expected to be %s but is %s", expectedKey, actualKey)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("TestSkipPendingUpdatesCache: Error running main part of test: %s", err)
	}
}

// TestFlushCommitTreapsErrors tests error-cases in *dbCache.flush() when commitTreaps returns error.
// The non-error-cases are tested in the more general tests.
func TestFlushCommitTreapsErrors(t *testing.T) {
	pdb := newTestDb("TestFlushCommitTreapsErrors", t)
	defer pdb.Close()

	key := []byte("key")
	value := []byte("value")

	// Before setting flush interval to zero - put some data so that there's something to flush
	err := pdb.Update(func(tx database.Tx) error {
		metadata := tx.Metadata()
		metadata.Put(key, value)

		return nil
	})
	if err != nil {
		t.Fatalf("TestFlushCommitTreapsErrors: Error putting some data to flush: %s", err)
	}

	cache := pdb.cache
	cache.flushInterval = 0 // set flushInterval to 0 so that flush is always required

	// Test for correctness when encountered error on Put
	func() {
		patch := monkey.Patch((*leveldb.Transaction).Put,
			func(*leveldb.Transaction, []byte, []byte, *opt.WriteOptions) error { return errors.New("error") })
		defer patch.Unpatch()

		err := pdb.Update(func(tx database.Tx) error {
			metadata := tx.Metadata()
			metadata.Put(key, value)

			return nil
		})

		if err == nil {
			t.Errorf("TestFlushCommitTreapsErrors: No error from pdb.Update when ldbTx.Put returned error")
		}
	}()

	// Test for correctness when encountered error on Delete

	// First put some data we can later "fail" to delete
	err = pdb.Update(func(tx database.Tx) error {
		metadata := tx.Metadata()
		metadata.Put(key, value)

		return nil
	})
	if err != nil {
		t.Fatalf("TestFlushCommitTreapsErrors: Error putting some data to delete: %s", err)
	}

	// Now "fail" to delete it
	func() {
		patch := monkey.Patch((*leveldb.Transaction).Delete,
			func(*leveldb.Transaction, []byte, *opt.WriteOptions) error { return errors.New("error") })
		defer patch.Unpatch()

		err := pdb.Update(func(tx database.Tx) error {
			metadata := tx.Metadata()
			metadata.Delete(key)

			return nil
		})

		if err == nil {
			t.Errorf("TestFlushCommitTreapsErrors: No error from pdb.Update when ldbTx.Delete returned error")
		}
	}()
}
