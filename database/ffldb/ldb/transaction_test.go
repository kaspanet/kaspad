package ldb

import (
	"github.com/kaspanet/kaspad/database"
	"strings"
	"testing"
)

func TestTransactionCloseErrors(t *testing.T) {
	tests := []struct {
		name              string
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
			name: "Rollback",
			function: func(dbTx *LevelDBTransaction) error {
				return dbTx.Rollback()
			},
			shouldReturnError: true,
		},
		{
			name: "Commit",
			function: func(dbTx *LevelDBTransaction) error {
				return dbTx.Commit()
			},
			shouldReturnError: true,
		},
		{
			name: "RollbackUnlessClosed",
			function: func(dbTx *LevelDBTransaction) error {
				return dbTx.RollbackUnlessClosed()
			},
			shouldReturnError: false,
		},
	}

	for _, test := range tests {
		func() {
			ldb, teardownFunc := prepareDatabaseForTest(t, "TestTransactionCloseErrors")
			defer teardownFunc()

			// Begin a new transaction
			dbTx, err := ldb.Begin()
			if err != nil {
				t.Fatalf("TestTransactionCloseErrors: Begin "+
					"unexpectedly failed: %s", err)
			}
			defer func() {
				err := dbTx.RollbackUnlessClosed()
				if err != nil {
					t.Fatalf("TestTransactionCloseErrors: RollbackUnlessClosed "+
						"unexpectedly failed: %s", err)
				}
			}()

			// Close the transaction
			err = dbTx.Commit()
			if err != nil {
				t.Fatalf("TestTransactionCloseErrors: Commit "+
					"unexpectedly failed: %s", err)
			}

			expectedErrContainsString := "closed transaction"

			// Make sure that the test function returns a "closed transaction" error
			err = test.function(dbTx)
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
		}()
	}
}
