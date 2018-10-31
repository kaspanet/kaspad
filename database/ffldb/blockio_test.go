package ffldb

import (
	"errors"
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/wire"
)

func TestDeleteFile(t *testing.T) {
	testBlock := util.NewBlock(wire.NewMsgBlock(
		wire.NewBlockHeader(1, []daghash.Hash{}, &daghash.Hash{}, 0, 0)))

	tests := []struct {
		fileNum     uint32
		expectedErr bool
	}{
		{0, false},
		{1, true},
	}

	for _, test := range tests {
		func() {
			pdb := newTestDb("TestDeleteFile", t)
			defer pdb.Close()

			err := pdb.Update(func(dbTx database.Tx) error {
				dbTx.StoreBlock(testBlock)
				return nil
			})
			if err != nil {
				t.Fatalf("TestDeleteFile: Error storing block: %s", err)
			}

			err = pdb.store.deleteFile(test.fileNum)
			if (err != nil) != test.expectedErr {
				t.Errorf("TestDeleteFile: %d: Expected error status: %t, but got: %t",
					test.fileNum, test.expectedErr, (err != nil))
			}
			if err == nil {
				filePath := blockFilePath(pdb.store.basePath, test.fileNum)
				if _, err := os.Stat(filePath); !os.IsNotExist(err) {
					t.Errorf("TestDeleteFile: %d: File %s still exists", test.fileNum, filePath)
				}
			}
		}()
	}
}

// TestHandleRollbackErrors tests all error-cases in *blockStore.handleRollback().
// The non-error-cases are tested in the more general tests.
// Since handleRollback just logs errors, this test simply causes all error-cases to be hit,
// and makes sure no panic occurs, as well as ensures the writeCursor was updated correctly.
func TestHandleRollbackErrors(t *testing.T) {
	testBlock := util.NewBlock(wire.NewMsgBlock(
		wire.NewBlockHeader(1, []daghash.Hash{}, &daghash.Hash{}, 0, 0)))

	testBlockSize := uint32(testBlock.MsgBlock().SerializeSize())
	tests := []struct {
		name        string
		fileNum     uint32
		offset      uint32
		target      interface{}
		replacement interface{}
	}{
		// offset should be size of block + 12 bytes for block network, size and checksum
		{"Nothing to rollback", 1, testBlockSize + 12, nil, nil},
		{"deleteFile fails", 0, 0, (*blockStore).deleteFile,
			func(*blockStore, uint32) error { return errors.New("error in blockstore.deleteFile") }},
		{"openWriteFile fails", 0, 0, (*blockStore).openWriteFile,
			func(*blockStore, uint32) (filer, error) { return nil, errors.New("error in blockstore.openWriteFile") }},
		{"file.Truncate fails", 0, 0, (*os.File).Truncate,
			func(*os.File, int64) error { return errors.New("error in file.Truncate") }},
		{"file.Sync fails", 0, 0, (*os.File).Sync,
			func(*os.File) error { return errors.New("error in file.Sync") }},
	}

	for _, test := range tests {
		func() {
			pdb := newTestDb("TestHandleRollbackErrors", t)
			defer pdb.Close()

			// Set maxBlockFileSize to testBlockSize so that writeCursor.curFileNum increments
			pdb.store.maxBlockFileSize = testBlockSize

			err := pdb.Update(func(dbTx database.Tx) error {
				return dbTx.StoreBlock(testBlock)
			})
			if err != nil {
				t.Fatalf("TestHandleRollbackErrors: %s: Error adding test block to dabase: %s", test.name, err)
			}

			if test.target != nil && test.replacement != nil {
				patch := monkey.Patch(test.target, test.replacement)
				defer patch.Unpatch()
			}

			pdb.store.handleRollback(test.fileNum, test.offset)

			if pdb.store.writeCursor.curFileNum != test.fileNum {
				t.Errorf("TestHandleRollbackErrors: %s: Expected fileNum: %d, but got: %d",
					test.name, test.fileNum, pdb.store.writeCursor.curFileNum)
			}

			if pdb.store.writeCursor.curOffset != test.offset {
				t.Errorf("TestHandleRollbackErrors: %s: offset fileNum: %d, but got: %d",
					test.name, test.offset, pdb.store.writeCursor.curOffset)
			}
		}()
	}
}
