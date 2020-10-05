package ldb

import (
	"strings"
	"testing"

	"github.com/kaspanet/kaspad/infrastructure/db/database"
)

func TestTransactionCloseErrors(t *testing.T) {
	tests := []struct {
		name string

		// function is the LevelDBTransaction function that
		// we're verifying whether it returns an error after
		// the transaction had been closed.
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
			ldb, teardownFunc := prepareDatabaseForTest(t, "TestTransactionCloseErrors")
			defer teardownFunc()

			// Begin a new transaction to test Commit
			commitTx, err := ldb.Begin()
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
			rollbackTx, err := ldb.Begin()
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
				err = test.function(closedTx.(*LevelDBTransaction))
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
