// Copyright (c) 2015-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package ffldb

import (
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/util"
)

// TestPersistence ensures that values stored are still valid after closing and
// reopening the database.
func TestPersistence(t *testing.T) {
	t.Parallel()

	// Create a new database to run tests against.
	dbPath := filepath.Join(os.TempDir(), "ffldb-persistencetest")
	_ = os.RemoveAll(dbPath)
	db, err := OpenDB(dbPath, blockDataNet, true)
	if err != nil {
		t.Errorf("Failed to create test database (%s) %v", dbType, err)
		return
	}
	defer os.RemoveAll(dbPath)
	defer db.Close()

	// Create a bucket, put some values into it, and store a block so they
	// can be tested for existence on re-open.
	bucket1Key := []byte("bucket1")
	storeValues := map[string]string{
		"b1key1": "foo1",
		"b1key2": "foo2",
		"b1key3": "foo3",
	}
	genesisBlock := util.NewBlock(dagconfig.MainnetParams.GenesisBlock)
	genesisHash := dagconfig.MainnetParams.GenesisHash
	err = db.Update(func(dbTx *Transaction) error {
		metadataBucket := dbTx.Metadata()
		if metadataBucket == nil {
			return errors.Errorf("Metadata: unexpected nil bucket")
		}

		bucket1, err := metadataBucket.CreateBucket(bucket1Key)
		if err != nil {
			return errors.Errorf("CreateBucket: unexpected error: %v",
				err)
		}

		for k, v := range storeValues {
			err := bucket1.Put([]byte(k), []byte(v))
			if err != nil {
				return errors.Errorf("Put: unexpected error: %v",
					err)
			}
		}

		if err := dbTx.StoreBlock(genesisBlock); err != nil {
			return errors.Errorf("StoreBlock: unexpected error: %v",
				err)
		}

		return nil
	})
	if err != nil {
		t.Errorf("Update: unexpected error: %v", err)
		return
	}

	// Close and reopen the database to ensure the values persist.
	db.Close()
	db, err = OpenDB(dbPath, blockDataNet, false)
	if err != nil {
		t.Errorf("Failed to open test database (%s) %v", dbType, err)
		return
	}
	defer db.Close()

	// Ensure the values previously stored in the 3rd namespace still exist
	// and are correct.
	err = db.View(func(dbTx *Transaction) error {
		metadataBucket := dbTx.Metadata()
		if metadataBucket == nil {
			return errors.Errorf("Metadata: unexpected nil bucket")
		}

		bucket1 := metadataBucket.Bucket(bucket1Key)
		if bucket1 == nil {
			return errors.Errorf("Bucket1: unexpected nil bucket")
		}

		for k, v := range storeValues {
			gotVal := bucket1.Get([]byte(k))
			if !reflect.DeepEqual(gotVal, []byte(v)) {
				return errors.Errorf("Get: key '%s' does not "+
					"match expected value - got %s, want %s",
					k, gotVal, v)
			}
		}

		genesisBlockBytes, _ := genesisBlock.Bytes()
		gotBytes, err := dbTx.FetchBlock(genesisHash)
		if err != nil {
			return errors.Errorf("FetchBlock: unexpected error: %v",
				err)
		}
		if !reflect.DeepEqual(gotBytes, genesisBlockBytes) {
			return errors.Errorf("FetchBlock: stored block mismatch")
		}

		return nil
	})
	if err != nil {
		t.Errorf("View: unexpected error: %v", err)
		return
	}
}

// TestInterface performs all interfaces tests for this database driver.
func TestInterface(t *testing.T) {
	t.Parallel()

	// Create a new database to run tests against.
	dbPath := filepath.Join(os.TempDir(), "ffldb-interfacetest")
	_ = os.RemoveAll(dbPath)
	db, err := OpenDB(dbPath, blockDataNet, true)
	if err != nil {
		t.Errorf("Failed to create test database (%s) %v", dbType, err)
		return
	}
	defer os.RemoveAll(dbPath)
	defer db.Close()

	// Ensure the driver type is the expected value.
	gotDbType := db.Type()
	if gotDbType != dbType {
		t.Errorf("Type: unepxected driver type - got %v, want %v",
			gotDbType, dbType)
		return
	}

	// Run all of the interface tests against the database.
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Change the maximum file size to a small value to force multiple flat
	// files with the test data set.
	// Change maximum open files to small value to force shifts in the LRU
	// mechanism
	TstRunWithMaxBlockFileSizeAndMaxOpenFiles(db, 2048, 10, func() {
		testInterface(t, db)
	})
}
