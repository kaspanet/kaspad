package ffldb

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/kaspanet/kaspad/wire"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

func newTestDb(testName string, t *testing.T) *DB {
	dbPath := path.Join(os.TempDir(), "db_test", testName)
	err := os.RemoveAll(dbPath)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("%s: Error deleting database folder before starting: %s", testName, err)
	}

	network := wire.Simnet

	opts := opt.Options{
		ErrorIfExist: true,
		Strict:       opt.DefaultStrict,
		Compression:  opt.NoCompression,
		Filter:       filter.NewBloomFilter(10),
	}
	metadataDbPath := filepath.Join(dbPath, metadataDbName)
	ldb, err := leveldb.OpenFile(metadataDbPath, &opts)
	if err != nil {
		t.Errorf("%s: Error opening metadataDbPath: %s", testName, err)
	}
	err = initDB(ldb)
	if err != nil {
		t.Errorf("%s: Error initializing metadata Db: %s", testName, err)
	}

	store := newBlockStore(dbPath, network)
	cache := newDbCache(ldb, store, defaultCacheSize, defaultFlushSecs)
	return &DB{store: store, cache: cache}
}
