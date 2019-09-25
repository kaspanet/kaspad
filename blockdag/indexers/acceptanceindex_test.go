package indexers

import (
	"fmt"
	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/wire"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"
)

func TestAcceptanceIndexSerializationAnDeserialization(t *testing.T) {
	txsAcceptanceData := blockdag.MultiBlockTxsAcceptanceData{}

	// Create test data
	hash, _ := daghash.NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	txIn1 := &wire.TxIn{SignatureScript: []byte{1}, PreviousOutpoint: wire.Outpoint{Index: 1}, Sequence: 0}
	txIn2 := &wire.TxIn{SignatureScript: []byte{2}, PreviousOutpoint: wire.Outpoint{Index: 2}, Sequence: 0}
	txOut1 := &wire.TxOut{ScriptPubKey: []byte{1}, Value: 10}
	txOut2 := &wire.TxOut{ScriptPubKey: []byte{2}, Value: 20}
	blockTxsAcceptanceData := blockdag.BlockTxsAcceptanceData{
		{
			Tx:         util.NewTx(wire.NewNativeMsgTx(wire.TxVersion, []*wire.TxIn{txIn1}, []*wire.TxOut{txOut1})),
			IsAccepted: true,
		},
		{
			Tx:         util.NewTx(wire.NewNativeMsgTx(wire.TxVersion, []*wire.TxIn{txIn2}, []*wire.TxOut{txOut2})),
			IsAccepted: false,
		},
	}
	txsAcceptanceData[*hash] = blockTxsAcceptanceData

	// Serialize
	serializedTxsAcceptanceData, err := serializeMultiBlockTxsAcceptanceData(txsAcceptanceData)
	if err != nil {
		t.Fatalf("TestAcceptanceIndexSerializationAnDeserialization: serialization failed: %s", err)
	}

	// Deserialize
	deserializedTxsAcceptanceData, err := deserializeMultiBlockTxsAcceptanceData(serializedTxsAcceptanceData)
	if err != nil {
		t.Fatalf("TestAcceptanceIndexSerializationAnDeserialization: deserialization failed: %s", err)
	}

	// Check that they're the same
	if !reflect.DeepEqual(txsAcceptanceData, deserializedTxsAcceptanceData) {
		t.Fatalf("TestAcceptanceIndexSerializationAnDeserialization: original data and deseralize data aren't equal")
	}
}

func TestAcceptanceIndexRecover(t *testing.T) {
	params := &dagconfig.SimNetParams
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

	db1, err := database.Create("ffldb", db1Path, params.Net)
	if err != nil {
		t.Fatalf("error creating db: %s", err)
	}

	db1Config := blockdag.Config{
		IndexManager: db1IndexManager,
		DAGParams:    params,
		DB:           db1,
	}

	db1DAG, teardown, err := blockdag.DAGSetup("", db1Config)
	if err != nil {
		t.Fatalf("TestAcceptanceIndexRecover: Failed to setup DAG instance: %v", err)
	}
	if teardown != nil {
		defer teardown()
	}

	for i := 1; i < len(blocks)-2; i++ {
		isOrphan, delay, err := db1DAG.ProcessBlock(blocks[i], blockdag.BFNone)
		if err != nil {
			t.Fatalf("ProcessBlock fail on block %v: %v\n", i, err)
		}
		if delay != 0 {
			t.Fatalf("ProcessBlock: block %d "+
				"is too far in the future", i)
		}
		if isOrphan {
			t.Fatalf("ProcessBlock incorrectly returned block %v "+
				"is an orphan\n", i)
		}
	}

	err = db1.FlushCache()
	if err != nil {
		t.Fatalf("Error flushing database to disk: %s", err)
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
		isOrphan, delay, err := db1DAG.ProcessBlock(blocks[i], blockdag.BFNone)
		if err != nil {
			t.Fatalf("ProcessBlock fail on block %v: %v\n", i, err)
		}
		if delay != 0 {
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

	db2, err := database.Open("ffldb", db2Path, params.Net)
	if err != nil {
		t.Fatalf("Error opening database: %s", err)
	}

	db2Config := blockdag.Config{
		DAGParams: params,
		DB:        db2,
	}

	db2DAG, teardown, err := blockdag.DAGSetup("", db2Config)
	if err != nil {
		t.Fatalf("TestAcceptanceIndexRecover: Failed to setup DAG instance: %v", err)
	}
	if teardown != nil {
		defer teardown()
	}

	for i := len(blocks) - 2; i < len(blocks); i++ {
		isOrphan, delay, err := db2DAG.ProcessBlock(blocks[i], blockdag.BFNone)
		if err != nil {
			t.Fatalf("ProcessBlock fail on block %v: %v\n", i, err)
		}
		if delay != 0 {
			t.Fatalf("ProcessBlock: block %d "+
				"is too far in the future", i)
		}
		if isOrphan {
			t.Fatalf("ProcessBlock incorrectly returned block %v "+
				"is an orphan\n", i)
		}
	}

	err = db2.FlushCache()
	if err != nil {
		t.Fatalf("Error flushing database to disk: %s", err)
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

	db3, err := database.Open("ffldb", db3Path, params.Net)
	if err != nil {
		t.Fatalf("Error opening database: %s", err)
	}

	db3AcceptanceIndex := NewAcceptanceIndex()
	db3IndexManager := NewManager([]Indexer{db3AcceptanceIndex})
	db3Config := blockdag.Config{
		IndexManager: db3IndexManager,
		DAGParams:    params,
		DB:           db3,
	}

	_, teardown, err = blockdag.DAGSetup("", db3Config)
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
			return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
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

func copyFile(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	defer out.Close()
	if err != nil {
		return err
	}

	in, err := os.Open(srcFile)
	defer in.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

func createIfNotExists(dir string, perm os.FileMode) error {
	if blockdag.FileExists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}

func copySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return err
	}
	return os.Symlink(link, dest)
}
