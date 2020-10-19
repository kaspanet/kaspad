package blockdag

// This file functions are not considered safe for regular use, and should be used for test purposes only.

import (
	"compress/bzip2"
	"encoding/binary"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/subnetworkid"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/blocknode"
	"github.com/kaspanet/kaspad/domain/txscript"
	"github.com/kaspanet/kaspad/util/daghash"
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
func DAGSetup(dbName string, openDb bool, dagConfig Config) (*BlockDAG, func(), error) {
	var teardown func()

	// To make sure that the teardown function is not called before any goroutines finished to run -
	// overwrite `spawn` to count the number of running goroutines
	spawnWaitGroup := sync.WaitGroup{}
	realSpawn := spawn
	spawn = func(name string, f func()) {
		spawnWaitGroup.Add(1)
		realSpawn(name, func() {
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

		// We set ldb.Options here to return nil because normally
		// the database is initialized with very large caches that
		// can make opening/closing the database for every test
		// quite heavy.
		originalLDBOptions := ldb.Options
		ldb.Options = func() *opt.Options {
			return nil
		}

		dbPath := filepath.Join(tmpDir, dbName)
		_ = os.RemoveAll(dbPath)
		databaseContext, err := dbaccess.New(dbPath)
		if err != nil {
			return nil, nil, errors.Errorf("error creating db: %s", err)
		}

		dagConfig.DatabaseContext = databaseContext

		// Setup a teardown function for cleaning up. This function is
		// returned to the caller to be invoked when it is done testing.
		teardown = func() {
			spawnWaitGroup.Wait()
			spawn = realSpawn
			databaseContext.Close()
			ldb.Options = originalLDBOptions
			os.RemoveAll(dbPath)
		}
	} else {
		teardown = func() {
			spawnWaitGroup.Wait()
			spawn = realSpawn
		}
	}

	dagConfig.TimeSource = NewTimeSource()
	dagConfig.SigCache = txscript.NewSigCache(1000)
	dagConfig.MaxUTXOCacheSize = config.DefaultConfig().MaxUTXOCacheSize

	// Create the DAG instance.
	dag, err := New(&dagConfig)
	if err != nil {
		teardown()
		err := errors.Wrapf(err, "failed to create dag instance")
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

func createTxForTest(numInputs uint32, numOutputs uint32, outputValue uint64, subnetworkData *txSubnetworkData) *appmessage.MsgTx {
	txIns := []*appmessage.TxIn{}
	txOuts := []*appmessage.TxOut{}

	for i := uint32(0); i < numInputs; i++ {
		txIns = append(txIns, &appmessage.TxIn{
			PreviousOutpoint: *appmessage.NewOutpoint(&daghash.TxID{}, i),
			SignatureScript:  []byte{},
			Sequence:         appmessage.MaxTxInSequenceNum,
		})
	}

	for i := uint32(0); i < numOutputs; i++ {
		txOuts = append(txOuts, &appmessage.TxOut{
			ScriptPubKey: OpTrueScript,
			Value:        outputValue,
		})
	}

	if subnetworkData != nil {
		return appmessage.NewSubnetworkMsgTx(appmessage.TxVersion, txIns, txOuts, subnetworkData.subnetworkID, subnetworkData.Gas, subnetworkData.Payload)
	}

	return appmessage.NewNativeMsgTx(appmessage.TxVersion, txIns, txOuts)
}

// LoadBlocks reads files containing kaspa gzipped block data from disk
// and returns them as an array of util.Block.
func LoadBlocks(filename string) (blocks []*util.Block, err error) {
	var network = appmessage.Mainnet
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
//
// Note: since we need to calculate acceptedIDMerkleRoot and utxoCommitment, we have to resolve selectedParent's utxo-set.
// Therefore, this might skew the test results in a way where blocks that should have been status UTXOPendingVerification have
// some other status.
func PrepareBlockForTest(dag *BlockDAG, parentHashes []*daghash.Hash, transactions []*appmessage.MsgTx) (*appmessage.MsgBlock, error) {
	parents := blocknode.NewSet()
	for _, hash := range parentHashes {
		parent, ok := dag.Index.LookupNode(hash)
		if !ok {
			return nil, errors.Errorf("parent %s was not found", hash)
		}
		parents.Add(parent)
	}
	node, _ := dag.newBlockNode(nil, parents)

	if dag.Index.BlockNodeStatus(node.SelectedParent) == blocknode.StatusUTXOPendingVerification {
		err := resolveNodeStatusForTest(dag, node.SelectedParent)
		if err != nil {
			return nil, err
		}
	}

	_, selectedParentPastUTXO, txsAcceptanceData, err := dag.pastUTXO(node)
	if err != nil {
		return nil, err
	}

	calculatedAccepetedIDMerkleRoot := calculateAcceptedIDMerkleRoot(txsAcceptanceData)

	multiset, err := dag.calcMultiset(node, txsAcceptanceData, selectedParentPastUTXO)
	if err != nil {
		return nil, err
	}

	calculatedMultisetHash := daghash.Hash(*multiset.Finalize())

	OpTrueAddr, err := opTrueAddress(dag.Params.Prefix)
	if err != nil {
		return nil, err
	}

	blockTransactions := make([]*util.Tx, len(transactions)+1)

	extraNonce := generateDeterministicExtraNonceForTest()
	coinbasePayloadExtraData, err := CoinbasePayloadExtraData(extraNonce, "")
	if err != nil {
		return nil, err
	}

	coinbasePayloadScriptPubKey, err := txscript.PayToAddrScript(OpTrueAddr)
	if err != nil {
		return nil, err
	}

	blockTransactions[0], err = dag.expectedCoinbaseTransaction(node,
		txsAcceptanceData, coinbasePayloadScriptPubKey, coinbasePayloadExtraData)
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

	// Create a new block ready to be solved.
	hashMerkleTree := BuildHashMerkleTreeStore(blockTransactions)

	var msgBlock appmessage.MsgBlock
	for _, tx := range blockTransactions {
		msgBlock.AddTransaction(tx.MsgTx())
	}

	timestamp := dag.PastMedianTime(node.Parents.Bluest())
	msgBlock.Header = appmessage.BlockHeader{
		Version: blockVersion,

		// We use parents.hashes() and not parentHashes because parents.hashes() is sorted.
		ParentHashes:         parents.Hashes(),
		HashMerkleRoot:       hashMerkleTree.Root(),
		AcceptedIDMerkleRoot: calculatedAccepetedIDMerkleRoot,
		UTXOCommitment:       &calculatedMultisetHash,
		Timestamp:            timestamp,
		Bits:                 dag.requiredDifficulty(node.Parents.Bluest(), timestamp),
	}

	return &msgBlock, nil
}

func resolveNodeStatusForTest(dag *BlockDAG, node *blocknode.Node) error {
	dbTx, err := dag.DatabaseContext.NewTx()
	if err != nil {
		return err
	}
	defer dbTx.RollbackUnlessClosed()

	err = dag.resolveNodeStatus(node, dbTx)
	if err != nil {
		return err
	}

	err = dbTx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// PrepareAndProcessBlockForTest prepares a block that points to the given parent
// hashes and process it.
func PrepareAndProcessBlockForTest(
	t *testing.T, dag *BlockDAG, parentHashes []*daghash.Hash, transactions []*appmessage.MsgTx) *appmessage.MsgBlock {

	daghash.Sort(parentHashes)
	block, err := PrepareBlockForTest(dag, parentHashes, transactions)
	if err != nil {
		t.Fatalf("error in PrepareBlockForTest: %+v", err)
	}
	utilBlock := util.NewBlock(block)
	isOrphan, isDelayed, err := dag.ProcessBlock(utilBlock, BFNoPoWCheck)
	if err != nil {
		t.Fatalf("unexpected error in ProcessBlock: %+v", err)
	}
	if isDelayed {
		t.Fatalf("block is too far in the future")
	}
	if isOrphan {
		t.Fatalf("block was unexpectedly orphan")
	}
	return block
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
