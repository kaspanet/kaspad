package ffldb

import (
	"errors"
	"testing"

	"github.com/bouk/monkey"
	"github.com/btcsuite/goleveldb/leveldb"
	ldbutil "github.com/btcsuite/goleveldb/leveldb/util"
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
