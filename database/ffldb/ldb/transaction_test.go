package ldb

import (
	"github.com/kaspanet/kaspad/database"
	"strings"
	"testing"
)

func TestTransactionCommitErrors(t *testing.T) {
	tests := []struct {
		name string

		// function is the LevelDBCursor function that we're
		// verifying whether it returns an error after the
		// transaction had been closed.
		function          func(dbTx *LevelDBTransaction) error
		shouldReturnError bool
	}{
		{
			name: "Put",
			function: func(dbTx *LevelDBTransaction) error {
				return dbTx.Put(database.MakeBucket().Key([]byte("key")), []byte("value"))
			},
			shouldReturnError: true,
		},
		{
			name: "Get",
			function: func(dbTx *LevelDBTransaction) error {
				_, err := dbTx.Get(database.MakeBucket().Key([]byte("key")))
				return err
			},
			shouldReturnError: true,
		},
		{
			name: "Has",
			function: func(dbTx *LevelDBTransaction) error {
				_, err := dbTx.Has(database.MakeBucket().Key([]byte("key")))
				return err
			},
			shouldReturnError: true,
		},
		{
			name: "Delete",
			function: func(dbTx *LevelDBTransaction) error {
				return dbTx.Delete(database.MakeBucket().Key([]byte("key")))
			},
			shouldReturnError: true,
		},
		{
			name: "Cursor",
			function: func(dbTx *LevelDBTransaction) error {
				_, err := dbTx.Cursor(database.MakeBucket([]byte("bucket")))
				return err
			},
			shouldReturnError: true,
		},
		{
			name:              "Rollback",
			function:          (*LevelDBTransaction).Rollback,
			shouldReturnError: true,
		},
		{
			name:              "Commit",
			function:          (*LevelDBTransaction).Commit,
			shouldReturnError: true,
		},
		{
			name:              "RollbackUnlessClosed",
			function:          (*LevelDBTransaction).RollbackUnlessClosed,
			shouldReturnError: false,
		},
	}

	for _, test := range tests {
		func() {
			ldb, teardownFunc := prepareDatabaseForTest(t, "TestTransactionCommitErrors")
			defer teardownFunc()

			// Begin a new transaction
			dbTx, err := ldb.Begin()
			if err != nil {
				t.Fatalf("TestTransactionCommitErrors: Begin "+
					"unexpectedly failed: %s", err)
			}
			defer func() {
				err := dbTx.RollbackUnlessClosed()
				if err != nil {
					t.Fatalf("TestTransactionCommitErrors: RollbackUnlessClosed "+
						"unexpectedly failed: %s", err)
				}
			}()

			// Commit the transaction
			err = dbTx.Commit()
			if err != nil {
				t.Fatalf("TestTransactionCommitErrors: Commit "+
					"unexpectedly failed: %s", err)
			}

			expectedErrContainsString := "closed transaction"

			// Make sure that the test function returns a "closed transaction" error
			err = test.function(dbTx)
			if test.shouldReturnError {
				if err == nil {
					t.Fatalf("TestTransactionCommitErrors: %s "+
						"unexpectedly succeeded", test.name)
				}
				if !strings.Contains(err.Error(), expectedErrContainsString) {
					t.Fatalf("TestTransactionCommitErrors: %s "+
						"returned wrong error. Want: %s, got: %s",
						test.name, expectedErrContainsString, err)
				}
			} else {
				if err != nil {
					t.Fatalf("TestTransactionCommitErrors: %s "+
						"unexpectedly failed: %s", test.name, err)
				}
			}
		}()
	}
}

func TestTransactionRollbackErrors(t *testing.T) {
	tests := []struct {
		name string

		// function is the LevelDBCursor function that we're
		// verifying whether it returns an error after the
		// transaction had been closed.
		function          func(dbTx *LevelDBTransaction) error
		shouldReturnError bool
	}{
		{
			name: "Put",
			function: func(dbTx *LevelDBTransaction) error {
				return dbTx.Put(database.MakeBucket().Key([]byte("key")), []byte("value"))
			},
			shouldReturnError: true,
		},
		{
			name: "Get",
			function: func(dbTx *LevelDBTransaction) error {
				_, err := dbTx.Get(database.MakeBucket().Key([]byte("key")))
				return err
			},
			shouldReturnError: true,
		},
		{
			name: "Has",
			function: func(dbTx *LevelDBTransaction) error {
				_, err := dbTx.Has(database.MakeBucket().Key([]byte("key")))
				return err
			},
			shouldReturnError: true,
		},
		{
			name: "Delete",
			function: func(dbTx *LevelDBTransaction) error {
				return dbTx.Delete(database.MakeBucket().Key([]byte("key")))
			},
			shouldReturnError: true,
		},
		{
			name: "Cursor",
			function: func(dbTx *LevelDBTransaction) error {
				_, err := dbTx.Cursor(database.MakeBucket([]byte("bucket")))
				return err
			},
			shouldReturnError: true,
		},
		{
			name:              "Rollback",
			function:          (*LevelDBTransaction).Rollback,
			shouldReturnError: true,
		},
		{
			name:              "Commit",
			function:          (*LevelDBTransaction).Commit,
			shouldReturnError: true,
		},
		{
			name:              "RollbackUnlessClosed",
			function:          (*LevelDBTransaction).RollbackUnlessClosed,
			shouldReturnError: false,
		},
	}

	for _, test := range tests {
		func() {
			ldb, teardownFunc := prepareDatabaseForTest(t, "TestTransactionRollbackErrors")
			defer teardownFunc()

			// Begin a new transaction
			dbTx, err := ldb.Begin()
			if err != nil {
				t.Fatalf("TestTransactionRollbackErrors: Begin "+
					"unexpectedly failed: %s", err)
			}
			defer func() {
				err := dbTx.RollbackUnlessClosed()
				if err != nil {
					t.Fatalf("TestTransactionRollbackErrors: RollbackUnlessClosed "+
						"unexpectedly failed: %s", err)
				}
			}()

			// Rollback the transaction
			err = dbTx.Rollback()
			if err != nil {
				t.Fatalf("TestTransactionRollbackErrors: Rollback "+
					"unexpectedly failed: %s", err)
			}

			expectedErrContainsString := "closed transaction"

			// Make sure that the test function returns a "closed transaction" error
			err = test.function(dbTx)
			if test.shouldReturnError {
				if err == nil {
					t.Fatalf("TestTransactionRollbackErrors: %s "+
						"unexpectedly succeeded", test.name)
				}
				if !strings.Contains(err.Error(), expectedErrContainsString) {
					t.Fatalf("TestTransactionRollbackErrors: %s "+
						"returned wrong error. Want: %s, got: %s",
						test.name, expectedErrContainsString, err)
				}
			} else {
				if err != nil {
					t.Fatalf("TestTransactionRollbackErrors: %s "+
						"unexpectedly failed: %s", test.name, err)
				}
			}
		}()
	}
}
