package blockdag

// This file functions are not considered safe for regular use, and should be used for test purposes only.

import (
	"compress/bzip2"
	"encoding/binary"
	"github.com/kaspanet/kaspad/dbaccess"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/kaspanet/kaspad/util/subnetworkid"

	"github.com/kaspanet/kaspad/txscript"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

// FileExists returns whether or not the named file or directory exists.
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// DAGSetup is used to create a new db and DAG instance with the genesis
// block already inserted. In addition to the new DAG instance, it returns
// a teardown function the caller should invoke when done testing to clean up.
// The openDB parameter instructs DAGSetup whether or not to also open the
// database. Setting it to false is useful in tests that handle database
// opening/closing by themselves.
func DAGSetup(dbName string, openDb bool, config Config) (*BlockDAG, func(), error) {
	var teardown func()

	// To make sure that the teardown function is not called before any goroutines finished to run -
	// overwrite `spawn` to count the number of running goroutines
	spawnWaitGroup := sync.WaitGroup{}
	realSpawn := spawn
	spawn = func(f func()) {
		spawnWaitGroup.Add(1)
		realSpawn(func() {
			f()
			spawnWaitGroup.Done()
		})
	}

	if openDb {
		var err error
		tmpDir, err := ioutil.TempDir("", "DAGSetup")
		if err != nil {
			return nil, nil, errors.Errorf("error creating temp dir: %s", err)
		}

		dbPath := filepath.Join(tmpDir, dbName)
		_ = os.RemoveAll(dbPath)
		err = dbaccess.Open(dbPath)
		if err != nil {
			return nil, nil, errors.Errorf("error creating db: %s", err)
		}

		// Setup a teardown function for cleaning up. This function is
		// returned to the caller to be invoked when it is done testing.
		teardown = func() {
			spawnWaitGroup.Wait()
			spawn = realSpawn
			dbaccess.Close()
			os.RemoveAll(dbPath)
		}
	} else {
		teardown = func() {
			spawnWaitGroup.Wait()
			spawn = realSpawn
		}
	}

	config.TimeSource = NewTimeSource()
	config.SigCache = txscript.NewSigCache(1000)

	// Create the DAG instance.
	dag, err := New(&config)
	if err != nil {
		teardown()
		err := errors.Errorf("failed to create dag instance: %s", err)
		return nil, nil, err
	}
	return dag, teardown, nil
}

// OpTrueScript is script returning TRUE
var OpTrueScript = []byte{txscript.OpTrue}

type txSubnetworkData struct {
	subnetworkID *subnetworkid.SubnetworkID
	Gas          uint64
	Payload      []byte
}

func createTxForTest(numInputs uint32, numOutputs uint32, outputValue uint64, subnetworkData *txSubnetworkData) *wire.MsgTx {
	txIns := []*wire.TxIn{}
	txOuts := []*wire.TxOut{}

	for i := uint32(0); i < numInputs; i++ {
		txIns = append(txIns, &wire.TxIn{
			PreviousOutpoint: *wire.NewOutpoint(&daghash.TxID{}, i),
			SignatureScript:  []byte{},
			Sequence:         wire.MaxTxInSequenceNum,
		})
	}

	for i := uint32(0); i < numOutputs; i++ {
		txOuts = append(txOuts, &wire.TxOut{
			ScriptPubKey: OpTrueScript,
			Value:        outputValue,
		})
	}

	if subnetworkData != nil {
		return wire.NewSubnetworkMsgTx(wire.TxVersion, txIns, txOuts, subnetworkData.subnetworkID, subnetworkData.Gas, subnetworkData.Payload)
	}

	return wire.NewNativeMsgTx(wire.TxVersion, txIns, txOuts)
}

// VirtualForTest is an exported version for virtualBlock, so that it can be returned by exported test_util methods
type VirtualForTest *virtualBlock

// SetVirtualForTest replaces the dag's virtual block. This function is used for test purposes only
func SetVirtualForTest(dag *BlockDAG, virtual VirtualForTest) VirtualForTest {
	oldVirtual := dag.virtual
	dag.virtual = virtual
	return VirtualForTest(oldVirtual)
}

// GetVirtualFromParentsForTest generates a virtual block with the given parents.
func GetVirtualFromParentsForTest(dag *BlockDAG, parentHashes []*daghash.Hash) (VirtualForTest, error) {
	parents := newBlockSet()
	for _, hash := range parentHashes {
		parent := dag.index.LookupNode(hash)
		if parent == nil {
			return nil, errors.Errorf("GetVirtualFromParentsForTest: didn't found node for hash %s", hash)
		}
		parents.add(parent)
	}
	virtual := newVirtualBlock(dag, parents)

	pastUTXO, _, _, err := dag.pastUTXO(&virtual.blockNode)
	if err != nil {
		return nil, err
	}
	diffUTXO := pastUTXO.clone().(*DiffUTXOSet)
	err = diffUTXO.meldToBase()
	if err != nil {
		return nil, err
	}
	virtual.utxoSet = diffUTXO.base

	return VirtualForTest(virtual), nil
}

// LoadBlocks reads files containing kaspa gzipped block data from disk
// and returns them as an array of util.Block.
func LoadBlocks(filename string) (blocks []*util.Block, err error) {
	var network = wire.Mainnet
	var dr io.Reader
	var fi io.ReadCloser

	fi, err = os.Open(filename)
	if err != nil {
		return
	}

	if strings.HasSuffix(filename, ".bz2") {
		dr = bzip2.NewReader(fi)
	} else {
		dr = fi
	}
	defer fi.Close()

	var block *util.Block

	err = nil
	for height := uint64(0); err == nil; height++ {
		var rintbuf uint32
		err = binary.Read(dr, binary.LittleEndian, &rintbuf)
		if err == io.EOF {
			// hit end of file at expected offset: no warning
			height--
			err = nil
			break
		}
		if err != nil {
			break
		}
		if rintbuf != uint32(network) {
			break
		}
		err = binary.Read(dr, binary.LittleEndian, &rintbuf)
		blocklen := rintbuf

		rbytes := make([]byte, blocklen)

		// read block
		dr.Read(rbytes)

		block, err = util.NewBlockFromBytes(rbytes)
		if err != nil {
			return
		}
		blocks = append(blocks, block)
	}

	return
}

// opTrueAddress returns an address pointing to a P2SH anyone-can-spend script
func opTrueAddress(prefix util.Bech32Prefix) (util.Address, error) {
	return util.NewAddressScriptHash(OpTrueScript, prefix)
}

// PrepareBlockForTest generates a block with the proper merkle roots, coinbase transaction etc. This function is used for test purposes only
func PrepareBlockForTest(dag *BlockDAG, parentHashes []*daghash.Hash, transactions []*wire.MsgTx) (*wire.MsgBlock, error) {
	newVirtual, err := GetVirtualFromParentsForTest(dag, parentHashes)
	if err != nil {
		return nil, err
	}
	oldVirtual := SetVirtualForTest(dag, newVirtual)
	defer SetVirtualForTest(dag, oldVirtual)

	OpTrueAddr, err := opTrueAddress(dag.dagParams.Prefix)
	if err != nil {
		return nil, err
	}

	blockTransactions := make([]*util.Tx, len(transactions)+1)

	extraNonce := generateDeterministicExtraNonceForTest()
	coinbasePayloadExtraData, err := CoinbasePayloadExtraData(extraNonce, "")
	if err != nil {
		return nil, err
	}

	blockTransactions[0], err = dag.NextCoinbaseFromAddress(OpTrueAddr, coinbasePayloadExtraData)
	if err != nil {
		return nil, err
	}

	for i, tx := range transactions {
		blockTransactions[i+1] = util.NewTx(tx)
	}

	// Sort transactions by subnetwork ID
	sort.Slice(blockTransactions, func(i, j int) bool {
		if blockTransactions[i].MsgTx().SubnetworkID.IsEqual(subnetworkid.SubnetworkIDCoinbase) {
			return true
		}
		if blockTransactions[j].MsgTx().SubnetworkID.IsEqual(subnetworkid.SubnetworkIDCoinbase) {
			return false
		}
		return subnetworkid.Less(&blockTransactions[i].MsgTx().SubnetworkID, &blockTransactions[j].MsgTx().SubnetworkID)
	})

	block, err := dag.BlockForMining(blockTransactions)
	if err != nil {
		return nil, err
	}
	block.Header.Timestamp = dag.NextBlockMinimumTime()
	block.Header.Bits = dag.NextRequiredDifficulty(block.Header.Timestamp)

	return block, nil
}

// generateDeterministicExtraNonceForTest returns a unique deterministic extra nonce for coinbase data, in order to create unique coinbase transactions.
func generateDeterministicExtraNonceForTest() uint64 {
	extraNonceForTest++
	return extraNonceForTest
}

func resetExtraNonceForTest() {
	extraNonceForTest = 0
}

var extraNonceForTest = uint64(0)
