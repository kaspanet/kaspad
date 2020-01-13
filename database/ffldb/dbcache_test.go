package ffldb

import (
	"bytes"
	"testing"

	ldbutil "github.com/btcsuite/goleveldb/leveldb/util"
	"github.com/kaspanet/kaspad/database"
)

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

	err = pdb.cache.flush()
	if err != nil {
		t.Fatalf("Error flushing cache: %s", err)
	}

	// test skips
	err = pdb.Update(func(dbTx database.Tx) error {
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
