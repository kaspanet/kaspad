package indexers

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/dbaccess"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"
)

func TestAcceptanceIndexSerializationAndDeserialization(t *testing.T) {
	// Create test data
	hash, _ := daghash.NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	txIn1 := &wire.TxIn{SignatureScript: []byte{1}, PreviousOutpoint: wire.Outpoint{Index: 1}, Sequence: 0}
	txIn2 := &wire.TxIn{SignatureScript: []byte{2}, PreviousOutpoint: wire.Outpoint{Index: 2}, Sequence: 0}
	txOut1 := &wire.TxOut{ScriptPubKey: []byte{1}, Value: 10}
	txOut2 := &wire.TxOut{ScriptPubKey: []byte{2}, Value: 20}
	blockTxsAcceptanceData := blockdag.BlockTxsAcceptanceData{
		BlockHash: *hash,
		TxAcceptanceData: []blockdag.TxAcceptanceData{
			{
				Tx:         util.NewTx(wire.NewNativeMsgTx(wire.TxVersion, []*wire.TxIn{txIn1}, []*wire.TxOut{txOut1})),
				IsAccepted: true,
			},
			{
				Tx:         util.NewTx(wire.NewNativeMsgTx(wire.TxVersion, []*wire.TxIn{txIn2}, []*wire.TxOut{txOut2})),
				IsAccepted: false,
			},
		},
	}
	multiBlockTxsAcceptanceData := blockdag.MultiBlockTxsAcceptanceData{blockTxsAcceptanceData}

	// Serialize
	serializedTxsAcceptanceData, err := serializeMultiBlockTxsAcceptanceData(multiBlockTxsAcceptanceData)
	if err != nil {
		t.Fatalf("TestAcceptanceIndexSerializationAndDeserialization: serialization failed: %s", err)
	}

	// Deserialize
	deserializedTxsAcceptanceData, err := deserializeMultiBlockTxsAcceptanceData(serializedTxsAcceptanceData)
	if err != nil {
		t.Fatalf("TestAcceptanceIndexSerializationAndDeserialization: deserialization failed: %s", err)
	}

	// Check that they're the same
	if !reflect.DeepEqual(multiBlockTxsAcceptanceData, deserializedTxsAcceptanceData) {
		t.Fatalf("TestAcceptanceIndexSerializationAndDeserialization: original data and deseralize data aren't equal")
	}
}

// TestAcceptanceIndexRecover tests the recoverability of the
// acceptance index.
// It does it by following these steps:
// * It creates a DAG with enabled acceptance index (let's call it dag1) and
//   make it process some blocks.
// * It creates a copy of dag1 (let's call it dag2), and disables the acceptance
//   index in it.
// * It processes two more blocks in both dag1 and dag2.
// * A copy of dag2 is created (let's call it dag3) with enabled
//   acceptance index
// * It checks that the two missing blocks are added to dag3 acceptance index by
//   comparing dag1's last block acceptance data and dag3's last block acceptance
//   data.
func TestAcceptanceIndexRecover(t *testing.T) {
	params := &dagconfig.SimnetParams
	params.BlockCoinbaseMaturity = 0

	testFiles := []string{
		"blk_0_to_4.dat",
		"blk_3B.dat",
	}

	var blocks []*util.Block
	for _, file := range testFiles {
		blockTmp, err := blockdag.LoadBlocks(filepath.Join("../testdata/", file))
		if err != nil {
			t.Fatalf("Error loading file: %v\n", err)
		}
		blocks = append(blocks, blockTmp...)
	}

	db1AcceptanceIndex := NewAcceptanceIndex()
	db1IndexManager := NewManager([]Indexer{db1AcceptanceIndex})
	db1Path, err := ioutil.TempDir("", "TestAcceptanceIndexRecover1")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %s", err)
	}
	defer os.RemoveAll(db1Path)

	err = dbaccess.Open(db1Path)
	if err != nil {
		t.Fatalf("error creating db: %s", err)
	}

	db1Config := blockdag.Config{
		IndexManager: db1IndexManager,
		DAGParams:    params,
	}

	db1DAG, teardown, err := blockdag.DAGSetup("", false, db1Config)
	if err != nil {
		t.Fatalf("TestAcceptanceIndexRecover: Failed to setup DAG instance: %v", err)
	}
	if teardown != nil {
		defer teardown()
	}

	for i := 1; i < len(blocks)-2; i++ {
		isOrphan, isDelayed, err := db1DAG.ProcessBlock(blocks[i], blockdag.BFNone)
		if err != nil {
			t.Fatalf("ProcessBlock fail on block %v: %v\n", i, err)
		}
		if isDelayed {
			t.Fatalf("ProcessBlock: block %d "+
				"is too far in the future", i)
		}
		if isOrphan {
			t.Fatalf("ProcessBlock incorrectly returned block %v "+
				"is an orphan\n", i)
		}
	}

	db2Path, err := ioutil.TempDir("", "TestAcceptanceIndexRecover2")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %s", err)
	}
	defer os.RemoveAll(db2Path)

	err = copyDirectory(db1Path, db2Path)
	if err != nil {
		t.Fatalf("copyDirectory: %s", err)
	}

	for i := len(blocks) - 2; i < len(blocks); i++ {
		isOrphan, isDelayed, err := db1DAG.ProcessBlock(blocks[i], blockdag.BFNone)
		if err != nil {
			t.Fatalf("ProcessBlock fail on block %v: %v\n", i, err)
		}
		if isDelayed {
			t.Fatalf("ProcessBlock: block %d "+
				"is too far in the future", i)
		}
		if isOrphan {
			t.Fatalf("ProcessBlock incorrectly returned block %v "+
				"is an orphan\n", i)
		}
	}

	db1LastBlockAcceptanceData, err := db1AcceptanceIndex.TxsAcceptanceData(blocks[len(blocks)-1].Hash())
	if err != nil {
		t.Fatalf("Error fetching acceptance data: %s", err)
	}

	err = dbaccess.Close()
	if err != nil {
		t.Fatalf("Error closing the database: %s", err)
	}
	err = dbaccess.Open(db2Path)
	if err != nil {
		t.Fatalf("error creating db: %s", err)
	}

	db2Config := blockdag.Config{
		DAGParams: params,
	}

	db2DAG, teardown, err := blockdag.DAGSetup("", false, db2Config)
	if err != nil {
		t.Fatalf("TestAcceptanceIndexRecover: Failed to setup DAG instance: %v", err)
	}
	if teardown != nil {
		defer teardown()
	}

	for i := len(blocks) - 2; i < len(blocks); i++ {
		isOrphan, isDelayed, err := db2DAG.ProcessBlock(blocks[i], blockdag.BFNone)
		if err != nil {
			t.Fatalf("ProcessBlock fail on block %v: %v\n", i, err)
		}
		if isDelayed {
			t.Fatalf("ProcessBlock: block %d "+
				"is too far in the future", i)
		}
		if isOrphan {
			t.Fatalf("ProcessBlock incorrectly returned block %v "+
				"is an orphan\n", i)
		}
	}

	db3Path, err := ioutil.TempDir("", "TestAcceptanceIndexRecover3")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %s", err)
	}
	defer os.RemoveAll(db3Path)
	err = copyDirectory(db2Path, db3Path)
	if err != nil {
		t.Fatalf("copyDirectory: %s", err)
	}

	err = dbaccess.Close()
	if err != nil {
		t.Fatalf("Error closing the database: %s", err)
	}
	err = dbaccess.Open(db3Path)
	if err != nil {
		t.Fatalf("error creating db: %s", err)
	}

	db3AcceptanceIndex := NewAcceptanceIndex()
	db3IndexManager := NewManager([]Indexer{db3AcceptanceIndex})
	db3Config := blockdag.Config{
		IndexManager: db3IndexManager,
		DAGParams:    params,
	}

	_, teardown, err = blockdag.DAGSetup("", false, db3Config)
	if err != nil {
		t.Fatalf("TestAcceptanceIndexRecover: Failed to setup DAG instance: %v", err)
	}
	if teardown != nil {
		defer teardown()
	}

	db3LastBlockAcceptanceData, err := db3AcceptanceIndex.TxsAcceptanceData(blocks[len(blocks)-1].Hash())
	if err != nil {
		t.Fatalf("Error fetching acceptance data: %s", err)
	}
	if !reflect.DeepEqual(db1LastBlockAcceptanceData, db3LastBlockAcceptanceData) {
		t.Fatalf("recovery failed")
	}
}

// This function is copied and modified from this stackoverflow answer: https://stackoverflow.com/a/56314145/2413761
func copyDirectory(scrDir, dest string) error {
	entries, err := ioutil.ReadDir(scrDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(scrDir, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		stat, ok := fileInfo.Sys().(*syscall.Stat_t)
		if !ok {
			return errors.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := createIfNotExists(destPath, 0755); err != nil {
				return err
			}
			if err := copyDirectory(sourcePath, destPath); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err := copySymLink(sourcePath, destPath); err != nil {
				return err
			}
		default:
			if err := copyFile(sourcePath, destPath); err != nil {
				return err
			}
		}

		if err := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); err != nil {
			return err
		}

		isSymlink := entry.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, entry.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

// This function is copied and modified from this stackoverflow answer: https://stackoverflow.com/a/56314145/2413761
func copyFile(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer out.Close()

	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer in.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

// This function is copied and modified from this stackoverflow answer: https://stackoverflow.com/a/56314145/2413761
func createIfNotExists(dir string, perm os.FileMode) error {
	if blockdag.FileExists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return errors.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}

// This function is copied and modified from this stackoverflow answer: https://stackoverflow.com/a/56314145/2413761
func copySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return err
	}
	return os.Symlink(link, dest)
}
