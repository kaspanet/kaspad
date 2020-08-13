package ffldb

import (
	"bytes"
	"github.com/kaspanet/kaspad/infrastructure/database"
	"strings"
	"testing"
)

func TestTransactionCommitForLevelDBMethods(t *testing.T) {
	db, teardownFunc := prepareDatabaseForTest(t, "TestTransactionCommitForLevelDBMethods")
	defer teardownFunc()

	// Put a value into the database
	key1 := database.MakeBucket().Key([]byte("key1"))
	value1 := []byte("value1")
	err := db.Put(key1, value1)
	if err != nil {
		t.Fatalf("TestTransactionCommitForLevelDBMethods: Put "+
			"unexpectedly failed: %s", err)
	}

	// Begin a new transaction
	dbTx, err := db.Begin()
	if err != nil {
		t.Fatalf("TestTransactionCommitForLevelDBMethods: Begin "+
			"unexpectedly failed: %s", err)
	}
	defer func() {
		err := dbTx.RollbackUnlessClosed()
		if err != nil {
			t.Fatalf("TestTransactionCommitForLevelDBMethods: RollbackUnlessClosed "+
				"unexpectedly failed: %s", err)
		}
	}()

	// Make sure that Has returns that the original value exists
	exists, err := dbTx.Has(key1)
	if err != nil {
		t.Fatalf("TestTransactionCommitForLevelDBMethods: Has "+
			"unexpectedly failed: %s", err)
	}
	if !exists {
		t.Fatalf("TestTransactionCommitForLevelDBMethods: Has " +
			"unexpectedly returned that the value does not exist")
	}

	// Get the existing value and make sure it's equal to the original
	existingValue, err := dbTx.Get(key1)
	if err != nil {
		t.Fatalf("TestTransactionCommitForLevelDBMethods: Get "+
			"unexpectedly failed: %s", err)
	}
	if !bytes.Equal(existingValue, value1) {
		t.Fatalf("TestTransactionCommitForLevelDBMethods: Get "+
			"returned unexpected value. Want: %s, got: %s",
			string(value1), string(existingValue))
	}

	// Delete the existing value
	err = dbTx.Delete(key1)
	if err != nil {
		t.Fatalf("TestTransactionCommitForLevelDBMethods: Delete "+
			"unexpectedly failed: %s", err)
	}

	// Try to get a value that does not exist and make sure it returns ErrNotFound
	_, err = dbTx.Get(database.MakeBucket().Key([]byte("doesn't exist")))
	if err == nil {
		t.Fatalf("TestTransactionCommitForLevelDBMethods: Get " +
			"unexpectedly succeeded")
	}
	if !database.IsNotFoundError(err) {
		t.Fatalf("TestTransactionCommitForLevelDBMethods: Get "+
			"returned unexpected error: %s", err)
	}

	// Put a new value
	key2 := database.MakeBucket().Key([]byte("key2"))
	value2 := []byte("value2")
	err = dbTx.Put(key2, value2)
	if err != nil {
		t.Fatalf("TestTransactionCommitForLevelDBMethods: Put "+
			"unexpectedly failed: %s", err)
	}

	// Commit the transaction
	err = dbTx.Commit()
	if err != nil {
		t.Fatalf("TestTransactionCommitForLevelDBMethods: Commit "+
			"unexpectedly failed: %s", err)
	}

	// Make sure that Has returns that the original value does NOT exist
	exists, err = db.Has(key1)
	if err != nil {
		t.Fatalf("TestTransactionCommitForLevelDBMethods: Has "+
			"unexpectedly failed: %s", err)
	}
	if exists {
		t.Fatalf("TestTransactionCommitForLevelDBMethods: Has " +
			"unexpectedly returned that the value exists")
	}

	// Try to Get the existing value and make sure an ErrNotFound is returned
	_, err = db.Get(key1)
	if err == nil {
		t.Fatalf("TestTransactionCommitForLevelDBMethods: Get " +
			"unexpectedly succeeded")
	}
	if !database.IsNotFoundError(err) {
		t.Fatalf("TestTransactionCommitForLevelDBMethods: Get "+
			"returned unexpected err: %s", err)
	}

	// Make sure that Has returns that the new value exists
	exists, err = db.Has(key2)
	if err != nil {
		t.Fatalf("TestTransactionCommitForLevelDBMethods: Has "+
			"unexpectedly failed: %s", err)
	}
	if !exists {
		t.Fatalf("TestTransactionCommitForLevelDBMethods: Has " +
			"unexpectedly returned that the value does not exist")
	}

	// Get the new value and make sure it's equal to the original
	existingValue, err = db.Get(key2)
	if err != nil {
		t.Fatalf("TestTransactionCommitForLevelDBMethods: Get "+
			"unexpectedly failed: %s", err)
	}
	if !bytes.Equal(existingValue, value2) {
		t.Fatalf("TestTransactionCommitForLevelDBMethods: Get "+
			"returned unexpected value. Want: %s, got: %s",
			string(value2), string(existingValue))
	}
}

func TestTransactionRollbackForLevelDBMethods(t *testing.T) {
	db, teardownFunc := prepareDatabaseForTest(t, "TestTransactionRollbackForLevelDBMethods")
	defer teardownFunc()

	// Put a value into the database
	key1 := database.MakeBucket().Key([]byte("key1"))
	value1 := []byte("value1")
	err := db.Put(key1, value1)
	if err != nil {
		t.Fatalf("TestTransactionRollbackForLevelDBMethods: Put "+
			"unexpectedly failed: %s", err)
	}

	// Begin a new transaction
	dbTx, err := db.Begin()
	if err != nil {
		t.Fatalf("TestTransactionRollbackForLevelDBMethods: Begin "+
			"unexpectedly failed: %s", err)
	}
	defer func() {
		err := dbTx.RollbackUnlessClosed()
		if err != nil {
			t.Fatalf("TestTransactionRollbackForLevelDBMethods: RollbackUnlessClosed "+
				"unexpectedly failed: %s", err)
		}
	}()

	// Make sure that Has returns that the original value exists
	exists, err := dbTx.Has(key1)
	if err != nil {
		t.Fatalf("TestTransactionRollbackForLevelDBMethods: Has "+
			"unexpectedly failed: %s", err)
	}
	if !exists {
		t.Fatalf("TestTransactionRollbackForLevelDBMethods: Has " +
			"unexpectedly returned that the value does not exist")
	}

	// Get the existing value and make sure it's equal to the original
	existingValue, err := dbTx.Get(key1)
	if err != nil {
		t.Fatalf("TestTransactionRollbackForLevelDBMethods: Get "+
			"unexpectedly failed: %s", err)
	}
	if !bytes.Equal(existingValue, value1) {
		t.Fatalf("TestTransactionRollbackForLevelDBMethods: Get "+
			"returned unexpected value. Want: %s, got: %s",
			string(value1), string(existingValue))
	}

	// Delete the existing value
	err = dbTx.Delete(key1)
	if err != nil {
		t.Fatalf("TestTransactionRollbackForLevelDBMethods: Delete "+
			"unexpectedly failed: %s", err)
	}

	// Put a new value
	key2 := database.MakeBucket().Key([]byte("key2"))
	value2 := []byte("value2")
	err = dbTx.Put(key2, value2)
	if err != nil {
		t.Fatalf("TestTransactionRollbackForLevelDBMethods: Put "+
			"unexpectedly failed: %s", err)
	}

	// Rollback the transaction
	err = dbTx.Rollback()
	if err != nil {
		t.Fatalf("TestTransactionRollbackForLevelDBMethods: Rollback "+
			"unexpectedly failed: %s", err)
	}

	// Make sure that Has returns that the original value still exists
	exists, err = db.Has(key1)
	if err != nil {
		t.Fatalf("TestTransactionRollbackForLevelDBMethods: Has "+
			"unexpectedly failed: %s", err)
	}
	if !exists {
		t.Fatalf("TestTransactionRollbackForLevelDBMethods: Has " +
			"unexpectedly returned that the value does not exist")
	}

	// Get the existing value and make sure it is still returned
	existingValue, err = db.Get(key1)
	if err != nil {
		t.Fatalf("TestTransactionRollbackForLevelDBMethods: Get "+
			"unexpectedly failed: %s", err)
	}
	if !bytes.Equal(existingValue, value1) {
		t.Fatalf("TestTransactionRollbackForLevelDBMethods: Get "+
			"returned unexpected value. Want: %s, got: %s",
			string(value1), string(existingValue))
	}

	// Make sure that Has returns that the new value does NOT exist
	exists, err = db.Has(key2)
	if err != nil {
		t.Fatalf("TestTransactionRollbackForLevelDBMethods: Has "+
			"unexpectedly failed: %s", err)
	}
	if exists {
		t.Fatalf("TestTransactionRollbackForLevelDBMethods: Has " +
			"unexpectedly returned that the value exists")
	}

	// Try to Get the new value and make sure it returns an ErrNotFound
	_, err = db.Get(key2)
	if err == nil {
		t.Fatalf("TestTransactionRollbackForLevelDBMethods: Get " +
			"unexpectedly succeeded")
	}
	if !database.IsNotFoundError(err) {
		t.Fatalf("TestTransactionRollbackForLevelDBMethods: Get "+
			"returned unexpected error: %s", err)
	}
}

func TestTransactionCloseErrors(t *testing.T) {
	tests := []struct {
		name              string
		function          func(dbTx database.Transaction) error
		shouldReturnError bool
	}{
		{
			name: "Put",
			function: func(dbTx database.Transaction) error {
				return dbTx.Put(database.MakeBucket().Key([]byte("key")), []byte("value"))
			},
			shouldReturnError: true,
		},
		{
			name: "Get",
			function: func(dbTx database.Transaction) error {
				_, err := dbTx.Get(database.MakeBucket().Key([]byte("key")))
				return err
			},
			shouldReturnError: true,
		},
		{
			name: "Has",
			function: func(dbTx database.Transaction) error {
				_, err := dbTx.Has(database.MakeBucket().Key([]byte("key")))
				return err
			},
			shouldReturnError: true,
		},
		{
			name: "Delete",
			function: func(dbTx database.Transaction) error {
				return dbTx.Delete(database.MakeBucket().Key([]byte("key")))
			},
			shouldReturnError: true,
		},
		{
			name: "Cursor",
			function: func(dbTx database.Transaction) error {
				_, err := dbTx.Cursor(database.MakeBucket([]byte("bucket")))
				return err
			},
			shouldReturnError: true,
		},
		{
			name: "AppendToStore",
			function: func(dbTx database.Transaction) error {
				_, err := dbTx.AppendToStore("store", []byte("data"))
				return err
			},
			shouldReturnError: true,
		},
		{
			name: "RetrieveFromStore",
			function: func(dbTx database.Transaction) error {
				_, err := dbTx.RetrieveFromStore("store", []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
				return err
			},
			shouldReturnError: true,
		},
		{
			name: "Rollback",
			function: func(dbTx database.Transaction) error {
				return dbTx.Rollback()
			},
			shouldReturnError: true,
		},
		{
			name: "Commit",
			function: func(dbTx database.Transaction) error {
				return dbTx.Commit()
			},
			shouldReturnError: true,
		},
		{
			name: "RollbackUnlessClosed",
			function: func(dbTx database.Transaction) error {
				return dbTx.RollbackUnlessClosed()
			},
			shouldReturnError: false,
		},
	}

	for _, test := range tests {
		func() {
			db, teardownFunc := prepareDatabaseForTest(t, "TestTransactionCloseErrors")
			defer teardownFunc()

			// Begin a new transaction to test Commit
			commitTx, err := db.Begin()
			if err != nil {
				t.Fatalf("TestTransactionCloseErrors: Begin "+
					"unexpectedly failed: %s", err)
			}
			defer func() {
				err := commitTx.RollbackUnlessClosed()
				if err != nil {
					t.Fatalf("TestTransactionCloseErrors: RollbackUnlessClosed "+
						"unexpectedly failed: %s", err)
				}
			}()

			// Commit the Commit test transaction
			err = commitTx.Commit()
			if err != nil {
				t.Fatalf("TestTransactionCloseErrors: Commit "+
					"unexpectedly failed: %s", err)
			}

			// Begin a new transaction to test Rollback
			rollbackTx, err := db.Begin()
			if err != nil {
				t.Fatalf("TestTransactionCloseErrors: Begin "+
					"unexpectedly failed: %s", err)
			}
			defer func() {
				err := rollbackTx.RollbackUnlessClosed()
				if err != nil {
					t.Fatalf("TestTransactionCloseErrors: RollbackUnlessClosed "+
						"unexpectedly failed: %s", err)
				}
			}()

			// Rollback the Rollback test transaction
			err = rollbackTx.Rollback()
			if err != nil {
				t.Fatalf("TestTransactionCloseErrors: Rollback "+
					"unexpectedly failed: %s", err)
			}

			expectedErrContainsString := "closed transaction"

			// Make sure that the test function returns a "closed transaction" error
			// for both the commitTx and the rollbackTx
			for _, closedTx := range []database.Transaction{commitTx, rollbackTx} {
				err = test.function(closedTx)
				if test.shouldReturnError {
					if err == nil {
						t.Fatalf("TestTransactionCloseErrors: %s "+
							"unexpectedly succeeded", test.name)
					}
					if !strings.Contains(err.Error(), expectedErrContainsString) {
						t.Fatalf("TestTransactionCloseErrors: %s "+
							"returned wrong error. Want: %s, got: %s",
							test.name, expectedErrContainsString, err)
					}
				} else {
					if err != nil {
						t.Fatalf("TestTransactionCloseErrors: %s "+
							"unexpectedly failed: %s", test.name, err)
					}
				}
			}
		}()
	}
}

func TestTransactionRollbackUnlessClosed(t *testing.T) {
	db, teardownFunc := prepareDatabaseForTest(t, "TestTransactionRollbackUnlessClosed")
	defer teardownFunc()

	// Begin a new transaction
	dbTx, err := db.Begin()
	if err != nil {
		t.Fatalf("TestTransactionRollbackUnlessClosed: Begin "+
			"unexpectedly failed: %s", err)
	}

	// Roll it back
	err = dbTx.RollbackUnlessClosed()
	if err != nil {
		t.Fatalf("TestTransactionRollbackUnlessClosed: RollbackUnlessClosed "+
			"unexpectedly failed: %s", err)
	}
}

func TestTransactionCommitForFlatFileMethods(t *testing.T) {
	db, teardownFunc := prepareDatabaseForTest(t, "TestTransactionCommitForFlatFileMethods")
	defer teardownFunc()

	// Put a value into the database
	store := "store"
	value1 := []byte("value1")
	location1, err := db.AppendToStore(store, value1)
	if err != nil {
		t.Fatalf("TestTransactionCommitForFlatFileMethods: AppendToStore "+
			"unexpectedly failed: %s", err)
	}

	// Begin a new transaction
	dbTx, err := db.Begin()
	if err != nil {
		t.Fatalf("TestTransactionCommitForFlatFileMethods: Begin "+
			"unexpectedly failed: %s", err)
	}
	defer func() {
		err := dbTx.RollbackUnlessClosed()
		if err != nil {
			t.Fatalf("TestTransactionCommitForFlatFileMethods: RollbackUnlessClosed "+
				"unexpectedly failed: %s", err)
		}
	}()

	// Retrieve the existing value and make sure it's equal to the original
	existingValue, err := dbTx.RetrieveFromStore(store, location1)
	if err != nil {
		t.Fatalf("TestTransactionCommitForFlatFileMethods: RetrieveFromStore "+
			"unexpectedly failed: %s", err)
	}
	if !bytes.Equal(existingValue, value1) {
		t.Fatalf("TestTransactionCommitForFlatFileMethods: RetrieveFromStore "+
			"returned unexpected value. Want: %s, got: %s",
			string(value1), string(existingValue))
	}

	// Put a new value
	value2 := []byte("value2")
	location2, err := dbTx.AppendToStore(store, value2)
	if err != nil {
		t.Fatalf("TestTransactionCommitForFlatFileMethods: AppendToStore "+
			"unexpectedly failed: %s", err)
	}

	// Commit the transaction
	err = dbTx.Commit()
	if err != nil {
		t.Fatalf("TestTransactionCommitForFlatFileMethods: Commit "+
			"unexpectedly failed: %s", err)
	}

	// Retrieve the new value and make sure it's equal to the original
	newValue, err := db.RetrieveFromStore(store, location2)
	if err != nil {
		t.Fatalf("TestTransactionCommitForFlatFileMethods: RetrieveFromStore "+
			"unexpectedly failed: %s", err)
	}
	if !bytes.Equal(newValue, value2) {
		t.Fatalf("TestTransactionCommitForFlatFileMethods: RetrieveFromStore "+
			"returned unexpected value. Want: %s, got: %s",
			string(value2), string(newValue))
	}
}
