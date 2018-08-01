package ffldb

import (
	"os"
	"testing"

	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/wire"
	"github.com/daglabs/btcutil"
)

func TestDeleteFile(t *testing.T) {
	testBlock := btcutil.NewBlock(wire.NewMsgBlock(
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

			err := pdb.Update(func(tx database.Tx) error {
				tx.StoreBlock(testBlock)
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
