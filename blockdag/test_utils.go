package blockdag

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/daglabs/btcd/util/subnetworkid"

	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/database"
	_ "github.com/daglabs/btcd/database/ffldb" // blank import ffldb so that its init() function runs before tests
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/wire"
)

const (
	// testDbType is the database backend type to use for the tests.
	testDbType = "ffldb"

	// testDbRoot is the root directory used to create all test databases.
	testDbRoot = "testdbs"

	// blockDataNet is the expected network in the test block data.
	blockDataNet = wire.MainNet
)

// isSupportedDbType returns whether or not the passed database type is
// currently supported.
func isSupportedDbType(dbType string) bool {
	supportedDrivers := database.SupportedDrivers()
	for _, driver := range supportedDrivers {
		if dbType == driver {
			return true
		}
	}

	return false
}

// filesExists returns whether or not the named file or directory exists.
func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// DAGSetup is used to create a new db and chain instance with the genesis
// block already inserted.  In addition to the new chain instance, it returns
// a teardown function the caller should invoke when done testing to clean up.
func DAGSetup(dbName string, config Config) (*BlockDAG, func(), error) {
	if !isSupportedDbType(testDbType) {
		return nil, nil, fmt.Errorf("unsupported db type %v", testDbType)
	}

	var teardown func()

	if config.DB == nil {
		// Create the root directory for test databases.
		if !fileExists(testDbRoot) {
			if err := os.MkdirAll(testDbRoot, 0700); err != nil {
				err := fmt.Errorf("unable to create test db "+
					"root: %v", err)
				return nil, nil, err
			}
		}

		dbPath := filepath.Join(testDbRoot, dbName)
		_ = os.RemoveAll(dbPath)
		var err error
		config.DB, err = database.Create(testDbType, dbPath, blockDataNet)
		if err != nil {
			return nil, nil, fmt.Errorf("error creating db: %v", err)
		}

		// Setup a teardown function for cleaning up.  This function is
		// returned to the caller to be invoked when it is done testing.
		teardown = func() {
			config.DB.Close()
			os.RemoveAll(dbPath)
			os.RemoveAll(testDbRoot)
		}
	}

	config.TimeSource = NewMedianTime()
	config.SigCache = txscript.NewSigCache(1000)

	// Create the DAG instance.
	dag, err := New(&config)
	if err != nil {
		teardown()
		err := fmt.Errorf("failed to create dag instance: %v", err)
		return nil, nil, err
	}
	return dag, teardown, nil
}

// OpTrueScript is script returning TRUE
var OpTrueScript = []byte{txscript.OpTrue}

type txSubnetworkData struct {
	subnetworkID subnetworkid.SubNetworkID
	Gas          uint64
	Payload      []byte
}

func createTxForTest(numInputs uint32, numOutputs uint32, outputValue uint64, subnetworkData *txSubnetworkData) *wire.MsgTx {
	tx := wire.NewMsgTx(wire.TxVersion)

	for i := uint32(0); i < numInputs; i++ {
		tx.AddTxIn(&wire.TxIn{
			PreviousOutPoint: *wire.NewOutPoint(&daghash.Hash{}, i),
			SignatureScript:  []byte{},
			Sequence:         wire.MaxTxInSequenceNum,
		})
	}
	for i := uint32(0); i < numOutputs; i++ {
		tx.AddTxOut(&wire.TxOut{
			PkScript: OpTrueScript,
			Value:    outputValue,
		})
	}

	if subnetworkData != nil {
		tx.SubNetworkID = subnetworkData.subnetworkID
		tx.Gas = subnetworkData.Gas
		tx.Payload = subnetworkData.Payload
	} else {
		tx.SubNetworkID = wire.SubNetworkDAGCoin
		tx.Gas = 0
		tx.Payload = []byte{}
	}
	return tx
}

// createCoinbaseTxForTest returns a coinbase transaction with the requested number of
// outputs paying an appropriate subsidy based on the passed block height to the
// address associated with the harness.  It automatically uses a standard
// signature script that starts with the block height
func createCoinbaseTxForTest(blockHeight int32, numOutputs uint32, extraNonce int64, params *dagconfig.Params) (*wire.MsgTx, error) {
	// Create standard coinbase script.
	coinbaseScript, err := txscript.NewScriptBuilder().
		AddInt64(int64(blockHeight)).AddInt64(extraNonce).Script()
	if err != nil {
		return nil, err
	}

	tx := wire.NewMsgTx(wire.TxVersion)
	tx.AddTxIn(&wire.TxIn{
		// Coinbase transactions have no inputs, so previous outpoint is
		// zero hash and max index.
		PreviousOutPoint: *wire.NewOutPoint(&daghash.Hash{},
			wire.MaxPrevOutIndex),
		SignatureScript: coinbaseScript,
		Sequence:        wire.MaxTxInSequenceNum,
	})
	totalInput := CalcBlockSubsidy(blockHeight, params)
	amountPerOutput := totalInput / uint64(numOutputs)
	remainder := totalInput - amountPerOutput*uint64(numOutputs)
	for i := uint32(0); i < numOutputs; i++ {
		// Ensure the final output accounts for any remainder that might
		// be left from splitting the input amount.
		amount := amountPerOutput
		if i == numOutputs-1 {
			amount = amountPerOutput + remainder
		}
		tx.AddTxOut(&wire.TxOut{
			PkScript: OpTrueScript,
			Value:    amount,
		})
	}

	return tx, nil
}

// RegisterSubNetworkForTest is used to register network on DAG with specified gas limit
func RegisterSubNetworkForTest(dag *BlockDAG, gasLimit uint64) (*subnetworkid.SubNetworkID, error) {
	blockTime := time.Unix(dag.selectedTip().timestamp, 0)
	extraNonce := int64(0)

	buildNextBlock := func(parents blockSet, txs []*wire.MsgTx) (*util.Block, error) {
		// We need to change the blockTime to keep all block hashes unique
		blockTime = blockTime.Add(time.Second)

		// We need to change the extraNonce to keep coinbase hashes unique
		extraNonce++

		bh := &wire.BlockHeader{
			Version:      1,
			Bits:         dag.genesis.bits,
			ParentHashes: parents.hashes(),
			Timestamp:    blockTime,
		}
		msgBlock := wire.NewMsgBlock(bh)
		blockHeight := parents.maxHeight() + 1
		coinbaseTx, err := createCoinbaseTxForTest(blockHeight, 1, extraNonce, dag.dagParams)
		if err != nil {
			return nil, err
		}
		_ = msgBlock.AddTransaction(coinbaseTx)

		for _, tx := range txs {
			_ = msgBlock.AddTransaction(tx)
		}

		return util.NewBlock(msgBlock), nil
	}

	addBlockToDAG := func(block *util.Block) (*blockNode, error) {
		dag.dagLock.Lock()
		defer dag.dagLock.Unlock()

		err := dag.maybeAcceptBlock(block, BFNone)
		if err != nil {
			return nil, err
		}

		return dag.index.LookupNode(block.Hash()), nil
	}

	currentNode := dag.selectedTip()

	// Create a block with a valid sub-network registry transaction
	registryTx := wire.NewMsgTx(wire.TxVersion)
	registryTx.SubNetworkID = wire.SubNetworkRegistry
	registryTx.Payload = make([]byte, 8)
	binary.LittleEndian.PutUint64(registryTx.Payload, gasLimit)

	// Add it to the DAG
	registryBlock, err := buildNextBlock(setFromSlice(currentNode), []*wire.MsgTx{registryTx})
	if err != nil {
		return nil, fmt.Errorf("could not build registry block: %s", err)
	}
	currentNode, err = addBlockToDAG(registryBlock)
	if err != nil {
		return nil, fmt.Errorf("could not add registry block to DAG: %s", err)
	}

	// Build a sub-network ID from the registry transaction
	subNetworkID, err := txToSubNetworkID(registryTx)
	if err != nil {
		return nil, fmt.Errorf("could not build sub-network ID: %s", err)
	}
	return subNetworkID, nil
}
