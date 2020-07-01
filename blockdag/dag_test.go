// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"fmt"
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/dbaccess"
	"github.com/pkg/errors"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/txscript"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/subnetworkid"
	"github.com/kaspanet/kaspad/wire"
)

func TestBlockCount(t *testing.T) {
	// Load up blocks such that there is a fork in the DAG.
	// (genesis block) -> 1 -> 2 -> 3 -> 4
	//                          \-> 3b
	testFiles := []string{
		"blk_0_to_4.dat",
		"blk_3B.dat",
	}

	var blocks []*util.Block
	for _, file := range testFiles {
		blockTmp, err := LoadBlocks(filepath.Join("testdata/", file))
		if err != nil {
			t.Fatalf("Error loading file: %v\n", err)
		}
		blocks = append(blocks, blockTmp...)
	}

	// Create a new database and DAG instance to run tests against.
	dag, teardownFunc, err := DAGSetup("TestBlockCount", true, Config{
		DAGParams: &dagconfig.SimnetParams,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	// Since we're not dealing with the real block DAG, set the coinbase
	// maturity to 0.
	dag.TestSetCoinbaseMaturity(0)

	for i := 1; i < len(blocks); i++ {
		isOrphan, isDelayed, err := dag.ProcessBlock(blocks[i], BFNone)
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

	expectedBlockCount := uint64(6)
	if dag.BlockCount() != expectedBlockCount {
		t.Errorf("TestBlockCount: BlockCount expected to return %v but got %v", expectedBlockCount, dag.BlockCount())
	}
}

// TestIsKnownBlock tests the IsKnownBlock API to ensure proper functionality.
func TestIsKnownBlock(t *testing.T) {
	// Load up blocks such that there is a fork in the DAG.
	// (genesis block) -> 1 -> 2 -> 3 -> 4
	//                          \-> 3b
	testFiles := []string{
		"blk_0_to_4.dat",
		"blk_3B.dat",
	}

	var blocks []*util.Block
	for _, file := range testFiles {
		blockTmp, err := LoadBlocks(filepath.Join("testdata/", file))
		if err != nil {
			t.Fatalf("Error loading file: %v\n", err)
		}
		blocks = append(blocks, blockTmp...)
	}

	// Create a new database and DAG instance to run tests against.
	dag, teardownFunc, err := DAGSetup("haveblock", true, Config{
		DAGParams: &dagconfig.SimnetParams,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	// Since we're not dealing with the real block DAG, set the coinbase
	// maturity to 0.
	dag.TestSetCoinbaseMaturity(0)

	for i := 1; i < len(blocks); i++ {
		isOrphan, isDelayed, err := dag.ProcessBlock(blocks[i], BFNone)
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

	// Test a block with related parents
	testFiles = []string{
		"blk_3C.dat",
	}

	for _, file := range testFiles {
		blockTmp, err := LoadBlocks(filepath.Join("testdata/", file))
		if err != nil {
			t.Fatalf("Error loading file: %v\n", err)
		}
		blocks = append(blocks, blockTmp...)
	}
	isOrphan, isDelayed, err := dag.ProcessBlock(blocks[6], BFNone)

	// Block 3C should fail to connect since its parents are related. (It points to 1 and 2, and 1 is the parent of 2)
	if err == nil {
		t.Fatalf("ProcessBlock for block 3C has no error when expected to have an error\n")
	}
	if isDelayed {
		t.Fatalf("ProcessBlock: block 3C " +
			"is too far in the future")
	}
	if isOrphan {
		t.Fatalf("ProcessBlock incorrectly returned block 3C " +
			"is an orphan\n")
	}

	// Test a block with the same input twice
	testFiles = []string{
		"blk_3D.dat",
	}

	for _, file := range testFiles {
		blockTmp, err := LoadBlocks(filepath.Join("testdata/", file))
		if err != nil {
			t.Fatalf("Error loading file: %v\n", err)
		}
		blocks = append(blocks, blockTmp...)
	}
	isOrphan, isDelayed, err = dag.ProcessBlock(blocks[7], BFNone)

	// Block 3D should fail to connect since it has a transaction with the same input twice
	if err == nil {
		t.Fatalf("ProcessBlock for block 3D has no error when expected to have an error\n")
	}
	var ruleErr RuleError
	ok := errors.As(err, &ruleErr)
	if !ok {
		t.Fatalf("ProcessBlock for block 3D expected a RuleError, but got %v\n", err)
	}
	if !ok || ruleErr.ErrorCode != ErrDuplicateTxInputs {
		t.Fatalf("ProcessBlock for block 3D expected error code %s but got %s\n", ErrDuplicateTxInputs, ruleErr.ErrorCode)
	}
	if isDelayed {
		t.Fatalf("ProcessBlock: block 3D " +
			"is too far in the future")
	}
	if isOrphan {
		t.Fatalf("ProcessBlock incorrectly returned block 3D " +
			"is an orphan\n")
	}

	// Insert an orphan block.
	isOrphan, isDelayed, err = dag.ProcessBlock(util.NewBlock(&Block100000), BFNoPoWCheck)
	if err != nil {
		t.Fatalf("Unable to process block 100000: %v", err)
	}
	if isDelayed {
		t.Fatalf("ProcessBlock incorrectly returned that block 100000 " +
			"has a delay")
	}
	if !isOrphan {
		t.Fatalf("ProcessBlock indicated block is an not orphan when " +
			"it should be\n")
	}

	tests := []struct {
		hash string
		want bool
	}{
		// Genesis block should be present.
		{hash: dagconfig.SimnetParams.GenesisHash.String(), want: true},

		// Block 3b should be present (as a second child of Block 2).
		{hash: "46314ca17e117b31b467fe1b26fd36c98ee83e750aa5e3b3c1c32870afbe5984", want: true},

		// Block 100000 should be present (as an orphan).
		{hash: "732c891529619d43b5aeb3df42ba25dea483a8c0aded1cf585751ebabea28f29", want: true},

		// Random hashes should not be available.
		{hash: "123", want: false},
	}

	for i, test := range tests {
		hash, err := daghash.NewHashFromStr(test.hash)
		if err != nil {
			t.Fatalf("NewHashFromStr: %v", err)
		}

		result := dag.IsKnownBlock(hash)
		if result != test.want {
			t.Fatalf("IsKnownBlock #%d got %v want %v", i, result,
				test.want)
		}
	}
}

// TestCalcSequenceLock tests the LockTimeToSequence function, and the
// CalcSequenceLock method of a DAG instance. The tests exercise several
// combinations of inputs to the CalcSequenceLock function in order to ensure
// the returned SequenceLocks are correct for each test instance.
func TestCalcSequenceLock(t *testing.T) {
	netParams := &dagconfig.SimnetParams

	blockVersion := int32(0x10000000)

	// Generate enough synthetic blocks for the rest of the test
	dag := newTestDAG(netParams)
	node := dag.selectedTip()
	blockTime := node.Header().Timestamp
	numBlocksToGenerate := 5
	for i := 0; i < numBlocksToGenerate; i++ {
		blockTime = blockTime.Add(time.Second)
		node = newTestNode(dag, blockSetFromSlice(node), blockVersion, 0, blockTime)
		dag.index.AddNode(node)
		dag.virtual.SetTips(blockSetFromSlice(node))
	}

	// Create a utxo view with a fake utxo for the inputs used in the
	// transactions created below. This utxo is added such that it has an
	// age of 4 blocks.
	msgTx := wire.NewNativeMsgTx(wire.TxVersion, nil, []*wire.TxOut{{ScriptPubKey: nil, Value: 10}})
	targetTx := util.NewTx(msgTx)
	utxoSet := NewFullUTXOSet()
	blueScore := uint64(numBlocksToGenerate) - 4
	if isAccepted, err := utxoSet.AddTx(targetTx.MsgTx(), blueScore); err != nil {
		t.Fatalf("AddTx unexpectedly failed. Error: %s", err)
	} else if !isAccepted {
		t.Fatalf("AddTx unexpectedly didn't add tx %s", targetTx.ID())
	}

	// Create a utxo that spends the fake utxo created above for use in the
	// transactions created in the tests. It has an age of 4 blocks. Note
	// that the sequence lock heights are always calculated from the same
	// point of view that they were originally calculated from for a given
	// utxo. That is to say, the height prior to it.
	utxo := wire.Outpoint{
		TxID:  *targetTx.ID(),
		Index: 0,
	}
	prevUtxoBlueScore := uint64(numBlocksToGenerate) - 4

	// Obtain the past median time from the PoV of the input created above.
	// The past median time for the input is the past median time from the PoV
	// of the block *prior* to the one that included it.
	medianTime := node.RelativeAncestor(5).PastMedianTime(dag).UnixMilliseconds()

	// The median time calculated from the PoV of the best block in the
	// test DAG. For unconfirmed inputs, this value will be used since
	// the MTP will be calculated from the PoV of the yet-to-be-mined
	// block.
	nextMedianTime := node.PastMedianTime(dag).UnixMilliseconds()
	nextBlockBlueScore := int32(numBlocksToGenerate) + 1

	// Add an additional transaction which will serve as our unconfirmed
	// output.
	unConfTx := wire.NewNativeMsgTx(wire.TxVersion, nil, []*wire.TxOut{{ScriptPubKey: nil, Value: 5}})
	unConfUtxo := wire.Outpoint{
		TxID:  *unConfTx.TxID(),
		Index: 0,
	}
	if isAccepted, err := utxoSet.AddTx(unConfTx, UnacceptedBlueScore); err != nil {
		t.Fatalf("AddTx unexpectedly failed. Error: %s", err)
	} else if !isAccepted {
		t.Fatalf("AddTx unexpectedly didn't add tx %s", unConfTx.TxID())
	}

	tests := []struct {
		name    string
		tx      *wire.MsgTx
		utxoSet UTXOSet
		mempool bool
		want    *SequenceLock
	}{
		// A transaction with a single input with max sequence number.
		// This sequence number has the high bit set, so sequence locks
		// should be disabled.
		{
			name:    "single input, max sequence number",
			tx:      wire.NewNativeMsgTx(1, []*wire.TxIn{{PreviousOutpoint: utxo, Sequence: wire.MaxTxInSequenceNum}}, nil),
			utxoSet: utxoSet,
			want: &SequenceLock{
				Milliseconds:   -1,
				BlockBlueScore: -1,
			},
		},
		// A transaction with a single input whose lock time is
		// expressed in seconds. However, the specified lock time is
		// below the required floor for time based lock times since
		// they have time granularity of 524288 milliseconds. As a result, the
		// milliseconds lock-time should be just before the median time of
		// the targeted block.
		{
			name:    "single input, milliseconds lock time below time granularity",
			tx:      wire.NewNativeMsgTx(1, []*wire.TxIn{{PreviousOutpoint: utxo, Sequence: LockTimeToSequence(true, 2)}}, nil),
			utxoSet: utxoSet,
			want: &SequenceLock{
				Milliseconds:   medianTime - 1,
				BlockBlueScore: -1,
			},
		},
		// A transaction with a single input whose lock time is
		// expressed in seconds. The number of seconds should be 1048575
		// milliseconds after the median past time of the DAG.
		{
			name:    "single input, 1048575 milliseconds after median time",
			tx:      wire.NewNativeMsgTx(1, []*wire.TxIn{{PreviousOutpoint: utxo, Sequence: LockTimeToSequence(true, 1048576)}}, nil),
			utxoSet: utxoSet,
			want: &SequenceLock{
				Milliseconds:   medianTime + 1048575,
				BlockBlueScore: -1,
			},
		},
		// A transaction with multiple inputs. The first input has a
		// lock time expressed in seconds. The second input has a
		// sequence lock in blocks with a value of 4. The last input
		// has a sequence number with a value of 5, but has the disable
		// bit set. So the first lock should be selected as it's the
		// latest lock that isn't disabled.
		{
			name: "multiple varied inputs",
			tx: wire.NewNativeMsgTx(1,
				[]*wire.TxIn{{
					PreviousOutpoint: utxo,
					Sequence:         LockTimeToSequence(true, 2621440),
				}, {
					PreviousOutpoint: utxo,
					Sequence:         LockTimeToSequence(false, 4),
				}, {
					PreviousOutpoint: utxo,
					Sequence: LockTimeToSequence(false, 5) |
						wire.SequenceLockTimeDisabled,
				}},
				nil),
			utxoSet: utxoSet,
			want: &SequenceLock{
				Milliseconds:   medianTime + (5 << wire.SequenceLockTimeGranularity) - 1,
				BlockBlueScore: int64(prevUtxoBlueScore) + 3,
			},
		},
		// Transaction with a single input. The input's sequence number
		// encodes a relative lock-time in blocks (3 blocks). The
		// sequence lock should  have a value of -1 for seconds, but a
		// height of 2 meaning it can be included at height 3.
		{
			name:    "single input, lock-time in blocks",
			tx:      wire.NewNativeMsgTx(1, []*wire.TxIn{{PreviousOutpoint: utxo, Sequence: LockTimeToSequence(false, 3)}}, nil),
			utxoSet: utxoSet,
			want: &SequenceLock{
				Milliseconds:   -1,
				BlockBlueScore: int64(prevUtxoBlueScore) + 2,
			},
		},
		// A transaction with two inputs with lock times expressed in
		// seconds. The selected sequence lock value for seconds should
		// be the time further in the future.
		{
			name: "two inputs, lock-times in seconds",
			tx: wire.NewNativeMsgTx(1, []*wire.TxIn{{
				PreviousOutpoint: utxo,
				Sequence:         LockTimeToSequence(true, 5242880),
			}, {
				PreviousOutpoint: utxo,
				Sequence:         LockTimeToSequence(true, 2621440),
			}}, nil),
			utxoSet: utxoSet,
			want: &SequenceLock{
				Milliseconds:   medianTime + (10 << wire.SequenceLockTimeGranularity) - 1,
				BlockBlueScore: -1,
			},
		},
		// A transaction with two inputs with lock times expressed in
		// blocks. The selected sequence lock value for blocks should
		// be the height further in the future, so a height of 10
		// indicating it can be included at height 11.
		{
			name: "two inputs, lock-times in blocks",
			tx: wire.NewNativeMsgTx(1,
				[]*wire.TxIn{{
					PreviousOutpoint: utxo,
					Sequence:         LockTimeToSequence(false, 1),
				}, {
					PreviousOutpoint: utxo,
					Sequence:         LockTimeToSequence(false, 11),
				}},
				nil),
			utxoSet: utxoSet,
			want: &SequenceLock{
				Milliseconds:   -1,
				BlockBlueScore: int64(prevUtxoBlueScore) + 10,
			},
		},
		// A transaction with multiple inputs. Two inputs are time
		// based, and the other two are block based. The lock lying
		// further into the future for both inputs should be chosen.
		{
			name: "four inputs, two lock-times in time, two lock-times in blocks",
			tx: wire.NewNativeMsgTx(1,
				[]*wire.TxIn{{
					PreviousOutpoint: utxo,
					Sequence:         LockTimeToSequence(true, 2621440),
				}, {
					PreviousOutpoint: utxo,
					Sequence:         LockTimeToSequence(true, 6815744),
				}, {
					PreviousOutpoint: utxo,
					Sequence:         LockTimeToSequence(false, 3),
				}, {
					PreviousOutpoint: utxo,
					Sequence:         LockTimeToSequence(false, 9),
				}},
				nil),
			utxoSet: utxoSet,
			want: &SequenceLock{
				Milliseconds:   medianTime + (13 << wire.SequenceLockTimeGranularity) - 1,
				BlockBlueScore: int64(prevUtxoBlueScore) + 8,
			},
		},
		// A transaction with a single unconfirmed input. As the input
		// is confirmed, the height of the input should be interpreted
		// as the height of the *next* block. So, a 2 block relative
		// lock means the sequence lock should be for 1 block after the
		// *next* block height, indicating it can be included 2 blocks
		// after that.
		{
			name:    "single input, unconfirmed, lock-time in blocks",
			tx:      wire.NewNativeMsgTx(1, []*wire.TxIn{{PreviousOutpoint: unConfUtxo, Sequence: LockTimeToSequence(false, 2)}}, nil),
			utxoSet: utxoSet,
			mempool: true,
			want: &SequenceLock{
				Milliseconds:   -1,
				BlockBlueScore: int64(nextBlockBlueScore) + 1,
			},
		},
		// A transaction with a single unconfirmed input. The input has
		// a time based lock, so the lock time should be based off the
		// MTP of the *next* block.
		{
			name:    "single input, unconfirmed, lock-time in milliseoncds",
			tx:      wire.NewNativeMsgTx(1, []*wire.TxIn{{PreviousOutpoint: unConfUtxo, Sequence: LockTimeToSequence(true, 1048576)}}, nil),
			utxoSet: utxoSet,
			mempool: true,
			want: &SequenceLock{
				Milliseconds:   nextMedianTime + 1048575,
				BlockBlueScore: -1,
			},
		},
	}

	t.Logf("Running %v SequenceLock tests", len(tests))
	for _, test := range tests {
		utilTx := util.NewTx(test.tx)
		seqLock, err := dag.CalcSequenceLock(utilTx, utxoSet, test.mempool)
		if err != nil {
			t.Fatalf("test '%s', unable to calc sequence lock: %v", test.name, err)
		}

		if seqLock.Milliseconds != test.want.Milliseconds {
			t.Fatalf("test '%s' got %v milliseconds want %v milliseconds",
				test.name, seqLock.Milliseconds, test.want.Milliseconds)
		}
		if seqLock.BlockBlueScore != test.want.BlockBlueScore {
			t.Fatalf("test '%s' got blue score of %v want blue score of %v ",
				test.name, seqLock.BlockBlueScore, test.want.BlockBlueScore)
		}
	}
}

func TestCalcPastMedianTime(t *testing.T) {
	netParams := &dagconfig.SimnetParams

	blockVersion := int32(0x10000000)

	dag := newTestDAG(netParams)
	numBlocks := uint32(300)
	nodes := make([]*blockNode, numBlocks)
	nodes[0] = dag.genesis
	blockTime := dag.genesis.Header().Timestamp
	for i := uint32(1); i < numBlocks; i++ {
		blockTime = blockTime.Add(time.Second)
		nodes[i] = newTestNode(dag, blockSetFromSlice(nodes[i-1]), blockVersion, 0, blockTime)
		dag.index.AddNode(nodes[i])
	}

	tests := []struct {
		blockNumber                      uint32
		expectedMillisecondsSinceGenesis int64
	}{
		{
			blockNumber:                      262,
			expectedMillisecondsSinceGenesis: 130000,
		},
		{
			blockNumber:                      270,
			expectedMillisecondsSinceGenesis: 138000,
		},
		{
			blockNumber:                      240,
			expectedMillisecondsSinceGenesis: 108000,
		},
		{
			blockNumber:                      5,
			expectedMillisecondsSinceGenesis: 0,
		},
	}

	for _, test := range tests {
		millisecondsSinceGenesis := nodes[test.blockNumber].PastMedianTime(dag).UnixMilliseconds() -
			dag.genesis.Header().Timestamp.UnixMilliseconds()

		if millisecondsSinceGenesis != test.expectedMillisecondsSinceGenesis {
			t.Errorf("TestCalcPastMedianTime: expected past median time of block %v to be %v milliseconds "+
				"from genesis but got %v",
				test.blockNumber, test.expectedMillisecondsSinceGenesis, millisecondsSinceGenesis)
		}
	}
}

func TestNew(t *testing.T) {
	tempDir := os.TempDir()

	dbPath := filepath.Join(tempDir, "TestNew")
	_ = os.RemoveAll(dbPath)
	err := dbaccess.Open(dbPath)
	if err != nil {
		t.Fatalf("error creating db: %s", err)
	}
	defer func() {
		dbaccess.Close()
		os.RemoveAll(dbPath)
	}()
	config := &Config{
		DAGParams:  &dagconfig.SimnetParams,
		TimeSource: NewTimeSource(),
		SigCache:   txscript.NewSigCache(1000),
	}
	_, err = New(config)
	if err != nil {
		t.Fatalf("failed to create dag instance: %s", err)
	}

	config.SubnetworkID = &subnetworkid.SubnetworkID{0xff}
	_, err = New(config)
	expectedErrorMessage := fmt.Sprintf("Cannot start kaspad with subnetwork ID %s because"+
		" its database is already built with subnetwork ID <nil>. If you"+
		" want to switch to a new database, please reset the"+
		" database by starting kaspad with --reset-db flag", config.SubnetworkID)
	if err.Error() != expectedErrorMessage {
		t.Errorf("Unexpected error. Expected error '%s' but got '%s'", expectedErrorMessage, err)
	}
}

// TestAcceptingInInit makes sure that blocks that were stored but not
// yet fully processed do get correctly processed on DAG init. This may
// occur when the node shuts down improperly while a block is being
// validated.
func TestAcceptingInInit(t *testing.T) {
	tempDir := os.TempDir()

	// Create a test database
	dbPath := filepath.Join(tempDir, "TestAcceptingInInit")
	_ = os.RemoveAll(dbPath)
	err := dbaccess.Open(dbPath)
	if err != nil {
		t.Fatalf("error creating db: %s", err)
	}
	defer func() {
		dbaccess.Close()
		os.RemoveAll(dbPath)
	}()

	// Create a DAG to add the test block into
	config := &Config{
		DAGParams:  &dagconfig.SimnetParams,
		TimeSource: NewTimeSource(),
		SigCache:   txscript.NewSigCache(1000),
	}
	dag, err := New(config)
	if err != nil {
		t.Fatalf("failed to create dag instance: %s", err)
	}

	// Load the test block
	blocks, err := LoadBlocks("testdata/blk_0_to_4.dat")
	if err != nil {
		t.Fatalf("Error loading file: %v\n", err)
	}
	genesisBlock := blocks[0]
	testBlock := blocks[1]

	// Create a test blockNode with an unvalidated status
	genesisNode, ok := dag.index.LookupNode(genesisBlock.Hash())
	if !ok {
		t.Fatalf("genesis block does not exist in the DAG")
	}
	testNode, _ := dag.newBlockNode(&testBlock.MsgBlock().Header, blockSetFromSlice(genesisNode))
	testNode.status = statusDataStored

	// Manually add the test block to the database
	dbTx, err := dbaccess.NewTx()
	if err != nil {
		t.Fatalf("Failed to open database "+
			"transaction: %s", err)
	}
	defer dbTx.RollbackUnlessClosed()
	err = storeBlock(dbTx, testBlock)
	if err != nil {
		t.Fatalf("Failed to store block: %s", err)
	}
	dbTestNode, err := serializeBlockNode(testNode)
	if err != nil {
		t.Fatalf("Failed to serialize blockNode: %s", err)
	}
	key := blockIndexKey(testNode.hash, testNode.blueScore)
	err = dbaccess.StoreIndexBlock(dbTx, key, dbTestNode)
	if err != nil {
		t.Fatalf("Failed to update block index: %s", err)
	}
	err = dbTx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit database "+
			"transaction: %s", err)
	}

	// Create a new DAG. We expect this DAG to process the
	// test node
	dag, err = New(config)
	if err != nil {
		t.Fatalf("failed to create dag instance: %s", err)
	}

	// Make sure that the test node's status is valid
	testNode, ok = dag.index.LookupNode(testBlock.Hash())
	if !ok {
		t.Fatalf("block %s does not exist in the DAG", testBlock.Hash())
	}

	if testNode.status&statusValid == 0 {
		t.Fatalf("testNode is unexpectedly invalid")
	}
}

func TestConfirmations(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	params := dagconfig.SimnetParams
	params.K = 1
	dag, teardownFunc, err := DAGSetup("TestConfirmations", true, Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()
	dag.TestSetCoinbaseMaturity(0)

	// Check that the genesis block of a DAG with only the genesis block in it has confirmations = 1.
	genesisConfirmations, err := dag.blockConfirmations(dag.genesis)
	if err != nil {
		t.Fatalf("TestConfirmations: confirmations for genesis block unexpectedly failed: %s", err)
	}
	if genesisConfirmations != 0 {
		t.Fatalf("TestConfirmations: unexpected confirmations for genesis block. Want: 1, got: %d", genesisConfirmations)
	}

	// Add a chain of blocks
	chainBlocks := make([]*wire.MsgBlock, 5)
	chainBlocks[0] = dag.dagParams.GenesisBlock
	for i := uint32(1); i < 5; i++ {
		chainBlocks[i] = prepareAndProcessBlockByParentMsgBlocks(t, dag, chainBlocks[i-1])
	}

	// Make sure that each one of the chain blocks has the expected confirmations number
	for i, block := range chainBlocks {
		confirmations, err := dag.BlockConfirmationsByHash(block.BlockHash())
		if err != nil {
			t.Fatalf("TestConfirmations: confirmations for block unexpectedly failed: %s", err)
		}

		expectedConfirmations := uint64(len(chainBlocks) - i - 1)
		if confirmations != expectedConfirmations {
			t.Fatalf("TestConfirmations: unexpected confirmations for block. "+
				"Want: %d, got: %d", expectedConfirmations, confirmations)
		}
	}

	branchingBlocks := make([]*wire.MsgBlock, 2)
	// Add two branching blocks
	branchingBlocks[0] = prepareAndProcessBlockByParentMsgBlocks(t, dag, chainBlocks[1])
	branchingBlocks[1] = prepareAndProcessBlockByParentMsgBlocks(t, dag, branchingBlocks[0])

	// Check that the genesis has a confirmations number == len(chainBlocks)
	genesisConfirmations, err = dag.blockConfirmations(dag.genesis)
	if err != nil {
		t.Fatalf("TestConfirmations: confirmations for genesis block unexpectedly failed: %s", err)
	}
	expectedGenesisConfirmations := uint64(len(chainBlocks)) - 1
	if genesisConfirmations != expectedGenesisConfirmations {
		t.Fatalf("TestConfirmations: unexpected confirmations for genesis block. "+
			"Want: %d, got: %d", expectedGenesisConfirmations, genesisConfirmations)
	}

	// Check that each of the tips has a 0 confirmations
	tips := dag.virtual.tips()
	for tip := range tips {
		tipConfirmations, err := dag.blockConfirmations(tip)
		if err != nil {
			t.Fatalf("TestConfirmations: confirmations for tip unexpectedly failed: %s", err)
		}
		expectedConfirmations := uint64(0)
		if tipConfirmations != expectedConfirmations {
			t.Fatalf("TestConfirmations: unexpected confirmations for tip. "+
				"Want: %d, got: %d", expectedConfirmations, tipConfirmations)
		}
	}

	// Generate 100 blocks to force the "main" chain to become red
	branchingChainTip := branchingBlocks[1]
	for i := uint32(0); i < 100; i++ {
		nextBranchingChainTip := prepareAndProcessBlockByParentMsgBlocks(t, dag, branchingChainTip)
		branchingChainTip = nextBranchingChainTip
	}

	// Make sure that a red block has confirmation number = 0
	redChainBlock := chainBlocks[3]
	redChainBlockConfirmations, err := dag.BlockConfirmationsByHash(redChainBlock.BlockHash())
	if err != nil {
		t.Fatalf("TestConfirmations: confirmations for red chain block unexpectedly failed: %s", err)
	}
	if redChainBlockConfirmations != 0 {
		t.Fatalf("TestConfirmations: unexpected confirmations for red chain block. "+
			"Want: 0, got: %d", redChainBlockConfirmations)
	}

	// Make sure that the red tip has confirmation number = 0
	redChainTip := chainBlocks[len(chainBlocks)-1]
	redChainTipConfirmations, err := dag.BlockConfirmationsByHash(redChainTip.BlockHash())
	if err != nil {
		t.Fatalf("TestConfirmations: confirmations for red chain tip unexpectedly failed: %s", err)
	}
	if redChainTipConfirmations != 0 {
		t.Fatalf("TestConfirmations: unexpected confirmations for red tip block. "+
			"Want: 0, got: %d", redChainTipConfirmations)
	}
}

func TestAcceptingBlock(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	params := dagconfig.SimnetParams
	params.K = 3
	dag, teardownFunc, err := DAGSetup("TestAcceptingBlock", true, Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()
	dag.TestSetCoinbaseMaturity(0)

	acceptingBlockByMsgBlock := func(block *wire.MsgBlock) (*blockNode, error) {
		node := nodeByMsgBlock(t, dag, block)
		return dag.acceptingBlock(node)
	}

	// Check that the genesis block of a DAG with only the genesis block in it is accepted by the virtual.
	genesisAcceptingBlock, err := dag.acceptingBlock(dag.genesis)
	if err != nil {
		t.Fatalf("TestAcceptingBlock: acceptingBlock for genesis block unexpectedly failed: %s", err)
	}
	if genesisAcceptingBlock != nil {
		t.Fatalf("TestAcceptingBlock: unexpected acceptingBlock for genesis block. "+
			"Want: nil, got: %s", genesisAcceptingBlock.hash)
	}

	numChainBlocks := uint32(10)
	chainBlocks := make([]*wire.MsgBlock, numChainBlocks)
	chainBlocks[0] = dag.dagParams.GenesisBlock
	for i := uint32(1); i <= numChainBlocks-1; i++ {
		chainBlocks[i] = prepareAndProcessBlockByParentMsgBlocks(t, dag, chainBlocks[i-1])
	}

	// Make sure that each chain block (including the genesis) is accepted by its child
	for i, chainBlockNode := range chainBlocks[:len(chainBlocks)-1] {
		expectedAcceptingBlockNode := nodeByMsgBlock(t, dag, chainBlocks[i+1])
		chainAcceptingBlockNode, err := acceptingBlockByMsgBlock(chainBlockNode)
		if err != nil {
			t.Fatalf("TestAcceptingBlock: acceptingBlock for chain block %d unexpectedly failed: %s", i, err)
		}
		if expectedAcceptingBlockNode != chainAcceptingBlockNode {
			t.Fatalf("TestAcceptingBlock: unexpected acceptingBlock for chain block. "+
				"Want: %s, got: %s", expectedAcceptingBlockNode.hash, chainAcceptingBlockNode.hash)
		}
	}

	// Make sure that the selected tip doesn't have an accepting
	tipAcceptingBlock, err := acceptingBlockByMsgBlock(chainBlocks[len(chainBlocks)-1])
	if err != nil {
		t.Fatalf("TestAcceptingBlock: acceptingBlock for tip unexpectedly failed: %s", err)
	}
	if tipAcceptingBlock != nil {
		t.Fatalf("TestAcceptingBlock: unexpected acceptingBlock for tip. "+
			"Want: nil, got: %s", tipAcceptingBlock.hash)
	}

	// Generate a chain tip that will be in the anticone of the selected tip and
	// in dag.virtual.blues.
	branchingChainTip := prepareAndProcessBlockByParentMsgBlocks(t, dag, chainBlocks[len(chainBlocks)-3])

	// Make sure that branchingChainTip is not in the selected parent chain
	isBranchingChainTipInSelectedParentChain, err := dag.IsInSelectedParentChain(branchingChainTip.BlockHash())
	if err != nil {
		t.Fatalf("TestAcceptingBlock: IsInSelectedParentChain unexpectedly failed: %s", err)
	}
	if isBranchingChainTipInSelectedParentChain {
		t.Fatalf("TestAcceptingBlock: branchingChainTip wasn't expected to be in the selected parent chain")
	}

	// Make sure that branchingChainTip is in the virtual blues
	isVirtualBlue := false
	for _, virtualBlue := range dag.virtual.blues {
		if branchingChainTip.BlockHash().IsEqual(virtualBlue.hash) {
			isVirtualBlue = true
			break
		}
	}
	if !isVirtualBlue {
		t.Fatalf("TestAcceptingBlock: redChainBlock was expected to be a virtual blue")
	}

	// Make sure that a block that is in the anticone of the selected tip and
	// in the blues of the virtual doesn't have an accepting block
	branchingChainTipAcceptionBlock, err := acceptingBlockByMsgBlock(branchingChainTip)
	if err != nil {
		t.Fatalf("TestAcceptingBlock: acceptingBlock for red chain block unexpectedly failed: %s", err)
	}
	if branchingChainTipAcceptionBlock != nil {
		t.Fatalf("TestAcceptingBlock: unexpected acceptingBlock for branchingChainTipAcceptionBlock. "+
			"Want: nil, got: %s", branchingChainTipAcceptionBlock.hash)
	}

	// Add shorter side-chain
	intersectionBlock := chainBlocks[1]
	sideChainTip := intersectionBlock
	for i := 0; i < len(chainBlocks)-3; i++ {
		sideChainTip = prepareAndProcessBlockByParentMsgBlocks(t, dag, sideChainTip)
	}

	// Make sure that the accepting block of the parent of the branching block didn't change
	expectedAcceptingBlock := nodeByMsgBlock(t, dag, chainBlocks[2])
	intersectionAcceptingBlock, err := acceptingBlockByMsgBlock(intersectionBlock)
	if err != nil {
		t.Fatalf("TestAcceptingBlock: acceptingBlock for intersection block unexpectedly failed: %s", err)
	}
	if expectedAcceptingBlock != intersectionAcceptingBlock {
		t.Fatalf("TestAcceptingBlock: unexpected acceptingBlock for intersection block. "+
			"Want: %s, got: %s", expectedAcceptingBlock.hash, intersectionAcceptingBlock.hash)
	}

	// Make sure that a block that is found in the red set of the selected tip
	// doesn't have an accepting block
	prepareAndProcessBlockByParentMsgBlocks(t, dag, sideChainTip, chainBlocks[len(chainBlocks)-1])

	sideChainTipAcceptingBlock, err := acceptingBlockByMsgBlock(sideChainTip)
	if err != nil {
		t.Fatalf("TestAcceptingBlock: acceptingBlock for sideChainTip unexpectedly failed: %s", err)
	}
	if sideChainTipAcceptingBlock != nil {
		t.Fatalf("TestAcceptingBlock: unexpected acceptingBlock for sideChainTip. "+
			"Want: nil, got: %s", intersectionAcceptingBlock.hash)
	}
}

func TestFinalizeNodesBelowFinalityPoint(t *testing.T) {
	testFinalizeNodesBelowFinalityPoint(t, true)
	testFinalizeNodesBelowFinalityPoint(t, false)
}

func testFinalizeNodesBelowFinalityPoint(t *testing.T, deleteDiffData bool) {
	params := dagconfig.SimnetParams
	params.K = 1
	dag, teardownFunc, err := DAGSetup("testFinalizeNodesBelowFinalityPoint", true, Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	blockVersion := int32(0x10000000)
	blockTime := dag.genesis.Header().Timestamp

	flushUTXODiffStore := func() {
		dbTx, err := dbaccess.NewTx()
		if err != nil {
			t.Fatalf("Failed to open database transaction: %s", err)
		}
		defer dbTx.RollbackUnlessClosed()
		err = dag.utxoDiffStore.flushToDB(dbTx)
		if err != nil {
			t.Fatalf("Error flushing utxoDiffStore data to DB: %s", err)
		}
		dag.utxoDiffStore.clearDirtyEntries()
		err = dbTx.Commit()
		if err != nil {
			t.Fatalf("Failed to commit database transaction: %s", err)
		}
	}

	addNode := func(parent *blockNode) *blockNode {
		blockTime = blockTime.Add(time.Second)
		node := newTestNode(dag, blockSetFromSlice(parent), blockVersion, 0, blockTime)
		node.updateParentsChildren()
		dag.index.AddNode(node)

		// Put dummy diff data in dag.utxoDiffStore
		err := dag.utxoDiffStore.setBlockDiff(node, NewUTXODiff())
		if err != nil {
			t.Fatalf("setBlockDiff: %s", err)
		}
		flushUTXODiffStore()
		return node
	}
	finalityInterval := dag.dagParams.FinalityInterval
	nodes := make([]*blockNode, 0, finalityInterval)
	currentNode := dag.genesis
	nodes = append(nodes, currentNode)
	for i := uint64(0); i <= finalityInterval*2; i++ {
		currentNode = addNode(currentNode)
		nodes = append(nodes, currentNode)
	}

	// Manually set the last finality point
	dag.lastFinalityPoint = nodes[finalityInterval-1]

	// Don't unload diffData
	currentDifference := maxBlueScoreDifferenceToKeepLoaded
	maxBlueScoreDifferenceToKeepLoaded = math.MaxUint64
	defer func() { maxBlueScoreDifferenceToKeepLoaded = currentDifference }()

	dag.finalizeNodesBelowFinalityPoint(deleteDiffData)
	flushUTXODiffStore()

	for _, node := range nodes[:finalityInterval-1] {
		if !node.isFinalized {
			t.Errorf("Node with blue score %d expected to be finalized", node.blueScore)
		}
		if _, ok := dag.utxoDiffStore.loaded[node]; deleteDiffData && ok {
			t.Errorf("The diff data of node with blue score %d should have been unloaded if deleteDiffData is %T", node.blueScore, deleteDiffData)
		} else if !deleteDiffData && !ok {
			t.Errorf("The diff data of node with blue score %d shouldn't have been unloaded if deleteDiffData is %T", node.blueScore, deleteDiffData)
		}

		_, err := dag.utxoDiffStore.diffDataFromDB(node.hash)
		exists := !dbaccess.IsNotFoundError(err)
		if exists && err != nil {
			t.Errorf("diffDataFromDB: %s", err)
			continue
		}

		if deleteDiffData && exists {
			t.Errorf("The diff data of node with blue score %d should have been deleted from the database if deleteDiffData is %T", node.blueScore, deleteDiffData)
			continue
		}

		if !deleteDiffData && !exists {
			t.Errorf("The diff data of node with blue score %d shouldn't have been deleted from the database if deleteDiffData is %T", node.blueScore, deleteDiffData)
			continue
		}
	}

	for _, node := range nodes[finalityInterval-1:] {
		if node.isFinalized {
			t.Errorf("Node with blue score %d wasn't expected to be finalized", node.blueScore)
		}
		if _, ok := dag.utxoDiffStore.loaded[node]; !ok {
			t.Errorf("The diff data of node with blue score %d shouldn't have been unloaded", node.blueScore)
		}
		if diffData, err := dag.utxoDiffStore.diffDataFromDB(node.hash); err != nil {
			t.Errorf("diffDataFromDB: %s", err)
		} else if diffData == nil {
			t.Errorf("The diff data of node with blue score %d shouldn't have been deleted from the database", node.blueScore)
		}
	}
}

func TestDAGIndexFailedStatus(t *testing.T) {
	params := dagconfig.SimnetParams
	dag, teardownFunc, err := DAGSetup("TestDAGIndexFailedStatus", true, Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	invalidCbTx := wire.NewSubnetworkMsgTx(wire.TxVersion, []*wire.TxIn{}, []*wire.TxOut{}, subnetworkid.SubnetworkIDCoinbase, 0, []byte{})
	txs := []*util.Tx{util.NewTx(invalidCbTx)}
	hashMerkleRoot := BuildHashMerkleTreeStore(txs).Root()
	invalidMsgBlock := wire.NewMsgBlock(
		wire.NewBlockHeader(
			1,
			[]*daghash.Hash{params.GenesisHash}, hashMerkleRoot,
			&daghash.Hash{},
			&daghash.Hash{},
			dag.genesis.bits,
			0),
	)
	invalidMsgBlock.AddTransaction(invalidCbTx)
	invalidBlock := util.NewBlock(invalidMsgBlock)
	isOrphan, isDelayed, err := dag.ProcessBlock(invalidBlock, BFNoPoWCheck)

	if !errors.As(err, &RuleError{}) {
		t.Fatalf("ProcessBlock: expected a rule error but got %s instead", err)
	}
	if isDelayed {
		t.Fatalf("ProcessBlock: invalidBlock " +
			"is too far in the future")
	}
	if isOrphan {
		t.Fatalf("ProcessBlock incorrectly returned invalidBlock " +
			"is an orphan\n")
	}

	invalidBlockNode, ok := dag.index.LookupNode(invalidBlock.Hash())
	if !ok {
		t.Fatalf("invalidBlockNode wasn't added to the block index as expected")
	}
	if invalidBlockNode.status&statusValidateFailed != statusValidateFailed {
		t.Fatalf("invalidBlockNode status to have %b flags raised (got: %b)", statusValidateFailed, invalidBlockNode.status)
	}

	invalidMsgBlockChild := wire.NewMsgBlock(
		wire.NewBlockHeader(1, []*daghash.Hash{
			invalidBlock.Hash(),
		}, hashMerkleRoot, &daghash.Hash{}, &daghash.Hash{}, dag.genesis.bits, 0),
	)
	invalidMsgBlockChild.AddTransaction(invalidCbTx)
	invalidBlockChild := util.NewBlock(invalidMsgBlockChild)

	isOrphan, isDelayed, err = dag.ProcessBlock(invalidBlockChild, BFNoPoWCheck)
	var ruleErr RuleError
	if ok := errors.As(err, &ruleErr); !ok || ruleErr.ErrorCode != ErrInvalidAncestorBlock {
		t.Fatalf("ProcessBlock: expected a rule error but got %s instead", err)
	}
	if isDelayed {
		t.Fatalf("ProcessBlock: invalidBlockChild " +
			"is too far in the future")
	}
	if isOrphan {
		t.Fatalf("ProcessBlock incorrectly returned invalidBlockChild " +
			"is an orphan\n")
	}
	invalidBlockChildNode, ok := dag.index.LookupNode(invalidBlockChild.Hash())
	if !ok {
		t.Fatalf("invalidBlockChild wasn't added to the block index as expected")
	}
	if invalidBlockChildNode.status&statusInvalidAncestor != statusInvalidAncestor {
		t.Fatalf("invalidBlockNode status to have %b flags raised (got %b)", statusInvalidAncestor, invalidBlockChildNode.status)
	}

	invalidMsgBlockGrandChild := wire.NewMsgBlock(
		wire.NewBlockHeader(1, []*daghash.Hash{
			invalidBlockChild.Hash(),
		}, hashMerkleRoot, &daghash.Hash{}, &daghash.Hash{}, dag.genesis.bits, 0),
	)
	invalidMsgBlockGrandChild.AddTransaction(invalidCbTx)
	invalidBlockGrandChild := util.NewBlock(invalidMsgBlockGrandChild)

	isOrphan, isDelayed, err = dag.ProcessBlock(invalidBlockGrandChild, BFNoPoWCheck)
	if ok := errors.As(err, &ruleErr); !ok || ruleErr.ErrorCode != ErrInvalidAncestorBlock {
		t.Fatalf("ProcessBlock: expected a rule error but got %s instead", err)
	}
	if isDelayed {
		t.Fatalf("ProcessBlock: invalidBlockGrandChild " +
			"is too far in the future")
	}
	if isOrphan {
		t.Fatalf("ProcessBlock incorrectly returned invalidBlockGrandChild " +
			"is an orphan\n")
	}
	invalidBlockGrandChildNode, ok := dag.index.LookupNode(invalidBlockGrandChild.Hash())
	if !ok {
		t.Fatalf("invalidBlockGrandChild wasn't added to the block index as expected")
	}
	if invalidBlockGrandChildNode.status&statusInvalidAncestor != statusInvalidAncestor {
		t.Fatalf("invalidBlockGrandChildNode status to have %b flags raised (got %b)", statusInvalidAncestor, invalidBlockGrandChildNode.status)
	}
}

func TestIsDAGCurrentMaxDiff(t *testing.T) {
	netParams := []*dagconfig.Params{
		&dagconfig.MainnetParams,
		&dagconfig.TestnetParams,
		&dagconfig.DevnetParams,
		&dagconfig.RegressionNetParams,
		&dagconfig.SimnetParams,
	}
	for _, params := range netParams {
		if params.TargetTimePerBlock*time.Duration(params.FinalityInterval) < isDAGCurrentMaxDiff {
			t.Errorf("in %s, a DAG can be considered current even if it's below the finality point", params.Name)
		}
	}
}

func testProcessBlockRuleError(t *testing.T, dag *BlockDAG, block *wire.MsgBlock, expectedRuleErr error) {
	isOrphan, isDelayed, err := dag.ProcessBlock(util.NewBlock(block), BFNoPoWCheck)

	err = checkRuleError(err, expectedRuleErr)
	if err != nil {
		t.Errorf("checkRuleError: %s", err)
	}

	if isDelayed {
		t.Fatalf("ProcessBlock: block " +
			"is too far in the future")
	}
	if isOrphan {
		t.Fatalf("ProcessBlock: block got unexpectedly orphaned")
	}
}

func TestDoubleSpends(t *testing.T) {
	params := dagconfig.SimnetParams
	params.BlockCoinbaseMaturity = 0
	// Create a new database and dag instance to run tests against.
	dag, teardownFunc, err := DAGSetup("TestDoubleSpends", true, Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("Failed to setup dag instance: %v", err)
	}
	defer teardownFunc()

	fundingBlock := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{params.GenesisHash}, nil)
	cbTx := fundingBlock.Transactions[0]

	signatureScript, err := txscript.PayToScriptHashSignatureScript(OpTrueScript, nil)
	if err != nil {
		t.Fatalf("Failed to build signature script: %s", err)
	}
	txIn := &wire.TxIn{
		PreviousOutpoint: wire.Outpoint{TxID: *cbTx.TxID(), Index: 0},
		SignatureScript:  signatureScript,
		Sequence:         wire.MaxTxInSequenceNum,
	}
	txOut := &wire.TxOut{
		ScriptPubKey: OpTrueScript,
		Value:        uint64(1),
	}
	tx1 := wire.NewNativeMsgTx(wire.TxVersion, []*wire.TxIn{txIn}, []*wire.TxOut{txOut})

	doubleSpendTxOut := &wire.TxOut{
		ScriptPubKey: OpTrueScript,
		Value:        uint64(2),
	}
	doubleSpendTx1 := wire.NewNativeMsgTx(wire.TxVersion, []*wire.TxIn{txIn}, []*wire.TxOut{doubleSpendTxOut})

	blockWithTx1 := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{fundingBlock.BlockHash()}, []*wire.MsgTx{tx1})

	// Check that a block will be rejected if it has a transaction that already exists in its past.
	anotherBlockWithTx1, err := PrepareBlockForTest(dag, []*daghash.Hash{blockWithTx1.BlockHash()}, nil)
	if err != nil {
		t.Fatalf("PrepareBlockForTest: %v", err)
	}

	// Manually add tx1.
	anotherBlockWithTx1.Transactions = append(anotherBlockWithTx1.Transactions, tx1)
	anotherBlockWithTx1UtilTxs := make([]*util.Tx, len(anotherBlockWithTx1.Transactions))
	for i, tx := range anotherBlockWithTx1.Transactions {
		anotherBlockWithTx1UtilTxs[i] = util.NewTx(tx)
	}
	anotherBlockWithTx1.Header.HashMerkleRoot = BuildHashMerkleTreeStore(anotherBlockWithTx1UtilTxs).Root()

	testProcessBlockRuleError(t, dag, anotherBlockWithTx1, ruleError(ErrOverwriteTx, ""))

	// Check that a block will be rejected if it has a transaction that double spends
	// a transaction from its past.
	blockWithDoubleSpendForTx1, err := PrepareBlockForTest(dag, []*daghash.Hash{blockWithTx1.BlockHash()}, nil)
	if err != nil {
		t.Fatalf("PrepareBlockForTest: %v", err)
	}

	// Manually add a transaction that double spends the block past.
	blockWithDoubleSpendForTx1.Transactions = append(blockWithDoubleSpendForTx1.Transactions, doubleSpendTx1)
	blockWithDoubleSpendForTx1UtilTxs := make([]*util.Tx, len(blockWithDoubleSpendForTx1.Transactions))
	for i, tx := range blockWithDoubleSpendForTx1.Transactions {
		blockWithDoubleSpendForTx1UtilTxs[i] = util.NewTx(tx)
	}
	blockWithDoubleSpendForTx1.Header.HashMerkleRoot = BuildHashMerkleTreeStore(blockWithDoubleSpendForTx1UtilTxs).Root()

	testProcessBlockRuleError(t, dag, blockWithDoubleSpendForTx1, ruleError(ErrMissingTxOut, ""))

	blockInAnticoneOfBlockWithTx1, err := PrepareBlockForTest(dag, []*daghash.Hash{fundingBlock.BlockHash()}, []*wire.MsgTx{doubleSpendTx1})
	if err != nil {
		t.Fatalf("PrepareBlockForTest: %v", err)
	}

	// Check that a block will not get rejected if it has a transaction that double spends
	// a transaction from its anticone.
	testProcessBlockRuleError(t, dag, blockInAnticoneOfBlockWithTx1, nil)

	// Check that a block will be rejected if it has two transactions that spend the same UTXO.
	blockWithDoubleSpendWithItself, err := PrepareBlockForTest(dag, []*daghash.Hash{fundingBlock.BlockHash()}, nil)
	if err != nil {
		t.Fatalf("PrepareBlockForTest: %v", err)
	}

	// Manually add tx1 and doubleSpendTx1.
	blockWithDoubleSpendWithItself.Transactions = append(blockWithDoubleSpendWithItself.Transactions, tx1, doubleSpendTx1)
	blockWithDoubleSpendWithItselfUtilTxs := make([]*util.Tx, len(blockWithDoubleSpendWithItself.Transactions))
	for i, tx := range blockWithDoubleSpendWithItself.Transactions {
		blockWithDoubleSpendWithItselfUtilTxs[i] = util.NewTx(tx)
	}
	blockWithDoubleSpendWithItself.Header.HashMerkleRoot = BuildHashMerkleTreeStore(blockWithDoubleSpendWithItselfUtilTxs).Root()

	testProcessBlockRuleError(t, dag, blockWithDoubleSpendWithItself, ruleError(ErrDoubleSpendInSameBlock, ""))

	// Check that a block will be rejected if it has the same transaction twice.
	blockWithDuplicateTransaction, err := PrepareBlockForTest(dag, []*daghash.Hash{fundingBlock.BlockHash()}, nil)
	if err != nil {
		t.Fatalf("PrepareBlockForTest: %v", err)
	}

	// Manually add tx1 twice.
	blockWithDuplicateTransaction.Transactions = append(blockWithDuplicateTransaction.Transactions, tx1, tx1)
	blockWithDuplicateTransactionUtilTxs := make([]*util.Tx, len(blockWithDuplicateTransaction.Transactions))
	for i, tx := range blockWithDuplicateTransaction.Transactions {
		blockWithDuplicateTransactionUtilTxs[i] = util.NewTx(tx)
	}
	blockWithDuplicateTransaction.Header.HashMerkleRoot = BuildHashMerkleTreeStore(blockWithDuplicateTransactionUtilTxs).Root()
	testProcessBlockRuleError(t, dag, blockWithDuplicateTransaction, ruleError(ErrDuplicateTx, ""))
}

func TestUTXOCommitment(t *testing.T) {
	// Create a new database and dag instance to run tests against.
	params := dagconfig.SimnetParams
	params.BlockCoinbaseMaturity = 0
	dag, teardownFunc, err := DAGSetup("TestUTXOCommitment", true, Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("TestUTXOCommitment: Failed to setup dag instance: %v", err)
	}
	defer teardownFunc()

	resetExtraNonceForTest()

	createTx := func(txToSpend *wire.MsgTx) *wire.MsgTx {
		scriptPubKey, err := txscript.PayToScriptHashScript(OpTrueScript)
		if err != nil {
			t.Fatalf("TestUTXOCommitment: failed to build script pub key: %s", err)
		}
		signatureScript, err := txscript.PayToScriptHashSignatureScript(OpTrueScript, nil)
		if err != nil {
			t.Fatalf("TestUTXOCommitment: failed to build signature script: %s", err)
		}
		txIn := &wire.TxIn{
			PreviousOutpoint: wire.Outpoint{TxID: *txToSpend.TxID(), Index: 0},
			SignatureScript:  signatureScript,
			Sequence:         wire.MaxTxInSequenceNum,
		}
		txOut := &wire.TxOut{
			ScriptPubKey: scriptPubKey,
			Value:        uint64(1),
		}
		return wire.NewNativeMsgTx(wire.TxVersion, []*wire.TxIn{txIn}, []*wire.TxOut{txOut})
	}

	// Build the following DAG:
	// G <- A <- B <- D
	//        <- C <-
	genesis := params.GenesisBlock

	// Block A:
	blockA := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{genesis.BlockHash()}, nil)

	// Block B:
	blockB := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{blockA.BlockHash()}, nil)

	// Block C:
	txSpendBlockACoinbase := createTx(blockA.Transactions[0])
	blockCTxs := []*wire.MsgTx{txSpendBlockACoinbase}
	blockC := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{blockA.BlockHash()}, blockCTxs)

	// Block D:
	txSpendTxInBlockC := createTx(txSpendBlockACoinbase)
	blockDTxs := []*wire.MsgTx{txSpendTxInBlockC}
	blockD := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{blockB.BlockHash(), blockC.BlockHash()}, blockDTxs)

	// Get the pastUTXO of blockD
	blockNodeD, ok := dag.index.LookupNode(blockD.BlockHash())
	if !ok {
		t.Fatalf("TestUTXOCommitment: blockNode for block D not found")
	}
	blockDPastUTXO, _, _, _ := dag.pastUTXO(blockNodeD)
	blockDPastDiffUTXOSet := blockDPastUTXO.(*DiffUTXOSet)

	// Build a Multiset for block D
	multiset := secp256k1.NewMultiset()
	for outpoint, entry := range blockDPastDiffUTXOSet.base.utxoCollection {
		var err error
		multiset, err = addUTXOToMultiset(multiset, entry, &outpoint)
		if err != nil {
			t.Fatalf("TestUTXOCommitment: addUTXOToMultiset unexpectedly failed")
		}
	}
	for outpoint, entry := range blockDPastDiffUTXOSet.UTXODiff.toAdd {
		var err error
		multiset, err = addUTXOToMultiset(multiset, entry, &outpoint)
		if err != nil {
			t.Fatalf("TestUTXOCommitment: addUTXOToMultiset unexpectedly failed")
		}
	}
	for outpoint, entry := range blockDPastDiffUTXOSet.UTXODiff.toRemove {
		var err error
		multiset, err = removeUTXOFromMultiset(multiset, entry, &outpoint)
		if err != nil {
			t.Fatalf("TestUTXOCommitment: removeUTXOFromMultiset unexpectedly failed")
		}
	}

	// Turn the multiset into a UTXO commitment
	utxoCommitment := daghash.Hash(*multiset.Finalize())

	// Make sure that the two commitments are equal
	if !utxoCommitment.IsEqual(blockNodeD.utxoCommitment) {
		t.Fatalf("TestUTXOCommitment: calculated UTXO commitment and "+
			"actual UTXO commitment don't match. Want: %s, got: %s",
			utxoCommitment, blockNodeD.utxoCommitment)
	}
}

func TestPastUTXOMultiSet(t *testing.T) {
	// Create a new database and dag instance to run tests against.
	params := dagconfig.SimnetParams
	dag, teardownFunc, err := DAGSetup("TestPastUTXOMultiSet", true, Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("TestPastUTXOMultiSet: Failed to setup dag instance: %v", err)
	}
	defer teardownFunc()

	// Build a short chain
	genesis := params.GenesisBlock
	blockA := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{genesis.BlockHash()}, nil)
	blockB := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{blockA.BlockHash()}, nil)
	blockC := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{blockB.BlockHash()}, nil)

	// Take blockC's selectedParentMultiset
	blockNodeC, ok := dag.index.LookupNode(blockC.BlockHash())
	if !ok {
		t.Fatalf("TestPastUTXOMultiSet: blockNode for blockC not found")
	}
	blockCSelectedParentMultiset, err := blockNodeC.selectedParentMultiset(dag)
	if err != nil {
		t.Fatalf("TestPastUTXOMultiSet: selectedParentMultiset unexpectedly failed: %s", err)
	}

	// Copy the multiset
	blockCSelectedParentMultisetCopy := *blockCSelectedParentMultiset
	blockCSelectedParentMultiset = &blockCSelectedParentMultisetCopy

	// Add a block on top of blockC
	PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{blockC.BlockHash()}, nil)

	// Get blockC's selectedParentMultiset again
	blockCSelectedParentMultiSetAfterAnotherBlock, err := blockNodeC.selectedParentMultiset(dag)
	if err != nil {
		t.Fatalf("TestPastUTXOMultiSet: selectedParentMultiset unexpectedly failed: %s", err)
	}

	// Make sure that blockC's selectedParentMultiset had not changed
	if !reflect.DeepEqual(blockCSelectedParentMultiset, blockCSelectedParentMultiSetAfterAnotherBlock) {
		t.Fatalf("TestPastUTXOMultiSet: selectedParentMultiset appears to have changed")
	}
}
