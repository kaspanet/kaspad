// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"bou.ke/monkey"

	"math/rand"

	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/util/subnetworkid"
	"github.com/daglabs/btcd/wire"
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
	dag, teardownFunc, err := DAGSetup("TestBlockCount", Config{
		DAGParams: &dagconfig.SimNetParams,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	// Since we're not dealing with the real block DAG, set the coinbase
	// maturity to 0.
	dag.TestSetCoinbaseMaturity(0)

	for i := 1; i < len(blocks); i++ {
		isOrphan, delay, err := dag.ProcessBlock(blocks[i], BFNone)
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

	expectedBlockCount := uint64(6)
	if dag.BlockCount() != expectedBlockCount {
		t.Errorf("TestBlockCount: BlockCount expected to return %v but got %v", expectedBlockCount, dag.BlockCount())
	}
}

// TestHaveBlock tests the HaveBlock API to ensure proper functionality.
func TestHaveBlock(t *testing.T) {
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
	dag, teardownFunc, err := DAGSetup("haveblock", Config{
		DAGParams: &dagconfig.SimNetParams,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	// Since we're not dealing with the real block DAG, set the coinbase
	// maturity to 0.
	dag.TestSetCoinbaseMaturity(0)

	for i := 1; i < len(blocks); i++ {
		isOrphan, delay, err := dag.ProcessBlock(blocks[i], BFNone)
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
	isOrphan, delay, err := dag.ProcessBlock(blocks[6], BFNone)

	// Block 3C should fail to connect since its parents are related. (It points to 1 and 2, and 1 is the parent of 2)
	if err == nil {
		t.Fatalf("ProcessBlock for block 3C has no error when expected to have an error\n")
	}
	if delay != 0 {
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
	isOrphan, delay, err = dag.ProcessBlock(blocks[7], BFNone)

	// Block 3D should fail to connect since it has a transaction with the same input twice
	if err == nil {
		t.Fatalf("ProcessBlock for block 3D has no error when expected to have an error\n")
	}
	rErr, ok := err.(RuleError)
	if !ok {
		t.Fatalf("ProcessBlock for block 3D expected a RuleError, but got %v\n", err)
	}
	if !ok || rErr.ErrorCode != ErrDuplicateTxInputs {
		t.Fatalf("ProcessBlock for block 3D expected error code %s but got %s\n", ErrDuplicateTxInputs, rErr.ErrorCode)
	}
	if delay != 0 {
		t.Fatalf("ProcessBlock: block 3D " +
			"is too far in the future")
	}
	if isOrphan {
		t.Fatalf("ProcessBlock incorrectly returned block 3D " +
			"is an orphan\n")
	}

	// Insert an orphan block.
	isOrphan, delay, err = dag.ProcessBlock(util.NewBlock(&Block100000), BFNoPoWCheck)
	if err != nil {
		t.Fatalf("Unable to process block 100000: %v", err)
	}
	if delay != 0 {
		t.Fatalf("ProcessBlock incorrectly returned that block 100000 "+
			"has a %s delay", delay)
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
		{hash: dagconfig.SimNetParams.GenesisHash.String(), want: true},

		// Block 3b should be present (as a second child of Block 2).
		{hash: "6ffe9704c50b3f1892ce9e667337304ec0e9eb50a23673bc8ff7aaa20745ee4a", want: true},

		// Block 100000 should be present (as an orphan).
		{hash: "65b20b048a074793ebfd1196e49341c8d194dabfc6b44a4fd0c607406e122baf", want: true},

		// Random hashes should not be available.
		{hash: "123", want: false},
	}

	for i, test := range tests {
		hash, err := daghash.NewHashFromStr(test.hash)
		if err != nil {
			t.Fatalf("NewHashFromStr: %v", err)
		}

		result := dag.HaveBlock(hash)
		if result != test.want {
			t.Fatalf("HaveBlock #%d got %v want %v", i, result,
				test.want)
		}
	}
}

// TestCalcSequenceLock tests the LockTimeToSequence function, and the
// CalcSequenceLock method of a DAG instance. The tests exercise several
// combinations of inputs to the CalcSequenceLock function in order to ensure
// the returned SequenceLocks are correct for each test instance.
func TestCalcSequenceLock(t *testing.T) {
	netParams := &dagconfig.SimNetParams

	blockVersion := int32(0x10000000)

	// Generate enough synthetic blocks for the rest of the test
	dag := newTestDAG(netParams)
	node := dag.selectedTip()
	blockTime := node.Header().Timestamp
	numBlocksToGenerate := 5
	for i := 0; i < numBlocksToGenerate; i++ {
		blockTime = blockTime.Add(time.Second)
		node = newTestNode(setFromSlice(node), blockVersion, 0, blockTime, netParams.K)
		dag.index.AddNode(node)
		dag.virtual.SetTips(setFromSlice(node))
	}

	// Create a utxo view with a fake utxo for the inputs used in the
	// transactions created below.  This utxo is added such that it has an
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
	// transactions created in the tests.  It has an age of 4 blocks.  Note
	// that the sequence lock heights are always calculated from the same
	// point of view that they were originally calculated from for a given
	// utxo.  That is to say, the height prior to it.
	utxo := wire.Outpoint{
		TxID:  *targetTx.ID(),
		Index: 0,
	}
	prevUtxoChainHeight := uint64(numBlocksToGenerate) - 4

	// Obtain the past median time from the PoV of the input created above.
	// The past median time for the input is the past median time from the PoV
	// of the block *prior* to the one that included it.
	medianTime := node.RelativeAncestor(5).PastMedianTime(dag).Unix()

	// The median time calculated from the PoV of the best block in the
	// test chain.  For unconfirmed inputs, this value will be used since
	// the MTP will be calculated from the PoV of the yet-to-be-mined
	// block.
	nextMedianTime := node.PastMedianTime(dag).Unix()
	nextBlockChainHeight := int32(numBlocksToGenerate) + 1

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
				Seconds:        -1,
				BlockBlueScore: -1,
			},
		},
		// A transaction with a single input whose lock time is
		// expressed in seconds.  However, the specified lock time is
		// below the required floor for time based lock times since
		// they have time granularity of 512 seconds.  As a result, the
		// seconds lock-time should be just before the median time of
		// the targeted block.
		{
			name:    "single input, seconds lock time below time granularity",
			tx:      wire.NewNativeMsgTx(1, []*wire.TxIn{{PreviousOutpoint: utxo, Sequence: LockTimeToSequence(true, 2)}}, nil),
			utxoSet: utxoSet,
			want: &SequenceLock{
				Seconds:        medianTime - 1,
				BlockBlueScore: -1,
			},
		},
		// A transaction with a single input whose lock time is
		// expressed in seconds.  The number of seconds should be 1023
		// seconds after the median past time of the last block in the
		// chain.
		{
			name:    "single input, 1023 seconds after median time",
			tx:      wire.NewNativeMsgTx(1, []*wire.TxIn{{PreviousOutpoint: utxo, Sequence: LockTimeToSequence(true, 1024)}}, nil),
			utxoSet: utxoSet,
			want: &SequenceLock{
				Seconds:        medianTime + 1023,
				BlockBlueScore: -1,
			},
		},
		// A transaction with multiple inputs.  The first input has a
		// lock time expressed in seconds.  The second input has a
		// sequence lock in blocks with a value of 4.  The last input
		// has a sequence number with a value of 5, but has the disable
		// bit set.  So the first lock should be selected as it's the
		// latest lock that isn't disabled.
		{
			name: "multiple varied inputs",
			tx: wire.NewNativeMsgTx(1,
				[]*wire.TxIn{{
					PreviousOutpoint: utxo,
					Sequence:         LockTimeToSequence(true, 2560),
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
				Seconds:        medianTime + (5 << wire.SequenceLockTimeGranularity) - 1,
				BlockBlueScore: int64(prevUtxoChainHeight) + 3,
			},
		},
		// Transaction with a single input.  The input's sequence number
		// encodes a relative lock-time in blocks (3 blocks).  The
		// sequence lock should  have a value of -1 for seconds, but a
		// height of 2 meaning it can be included at height 3.
		{
			name:    "single input, lock-time in blocks",
			tx:      wire.NewNativeMsgTx(1, []*wire.TxIn{{PreviousOutpoint: utxo, Sequence: LockTimeToSequence(false, 3)}}, nil),
			utxoSet: utxoSet,
			want: &SequenceLock{
				Seconds:        -1,
				BlockBlueScore: int64(prevUtxoChainHeight) + 2,
			},
		},
		// A transaction with two inputs with lock times expressed in
		// seconds.  The selected sequence lock value for seconds should
		// be the time further in the future.
		{
			name: "two inputs, lock-times in seconds",
			tx: wire.NewNativeMsgTx(1, []*wire.TxIn{{
				PreviousOutpoint: utxo,
				Sequence:         LockTimeToSequence(true, 5120),
			}, {
				PreviousOutpoint: utxo,
				Sequence:         LockTimeToSequence(true, 2560),
			}}, nil),
			utxoSet: utxoSet,
			want: &SequenceLock{
				Seconds:        medianTime + (10 << wire.SequenceLockTimeGranularity) - 1,
				BlockBlueScore: -1,
			},
		},
		// A transaction with two inputs with lock times expressed in
		// blocks.  The selected sequence lock value for blocks should
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
				Seconds:        -1,
				BlockBlueScore: int64(prevUtxoChainHeight) + 10,
			},
		},
		// A transaction with multiple inputs.  Two inputs are time
		// based, and the other two are block based. The lock lying
		// further into the future for both inputs should be chosen.
		{
			name: "four inputs, two lock-times in time, two lock-times in blocks",
			tx: wire.NewNativeMsgTx(1,
				[]*wire.TxIn{{
					PreviousOutpoint: utxo,
					Sequence:         LockTimeToSequence(true, 2560),
				}, {
					PreviousOutpoint: utxo,
					Sequence:         LockTimeToSequence(true, 6656),
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
				Seconds:        medianTime + (13 << wire.SequenceLockTimeGranularity) - 1,
				BlockBlueScore: int64(prevUtxoChainHeight) + 8,
			},
		},
		// A transaction with a single unconfirmed input.  As the input
		// is confirmed, the height of the input should be interpreted
		// as the height of the *next* block.  So, a 2 block relative
		// lock means the sequence lock should be for 1 block after the
		// *next* block height, indicating it can be included 2 blocks
		// after that.
		{
			name:    "single input, unconfirmed, lock-time in blocks",
			tx:      wire.NewNativeMsgTx(1, []*wire.TxIn{{PreviousOutpoint: unConfUtxo, Sequence: LockTimeToSequence(false, 2)}}, nil),
			utxoSet: utxoSet,
			mempool: true,
			want: &SequenceLock{
				Seconds:        -1,
				BlockBlueScore: int64(nextBlockChainHeight) + 1,
			},
		},
		// A transaction with a single unconfirmed input.  The input has
		// a time based lock, so the lock time should be based off the
		// MTP of the *next* block.
		{
			name:    "single input, unconfirmed, lock-time in seoncds",
			tx:      wire.NewNativeMsgTx(1, []*wire.TxIn{{PreviousOutpoint: unConfUtxo, Sequence: LockTimeToSequence(true, 1024)}}, nil),
			utxoSet: utxoSet,
			mempool: true,
			want: &SequenceLock{
				Seconds:        nextMedianTime + 1023,
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

		if seqLock.Seconds != test.want.Seconds {
			t.Fatalf("test '%s' got %v seconds want %v seconds",
				test.name, seqLock.Seconds, test.want.Seconds)
		}
		if seqLock.BlockBlueScore != test.want.BlockBlueScore {
			t.Fatalf("test '%s' got blue score of %v want blue score of %v ",
				test.name, seqLock.BlockBlueScore, test.want.BlockBlueScore)
		}
	}
}

func TestCalcPastMedianTime(t *testing.T) {
	netParams := &dagconfig.SimNetParams

	blockVersion := int32(0x10000000)

	dag := newTestDAG(netParams)
	numBlocks := uint32(300)
	nodes := make([]*blockNode, numBlocks)
	nodes[0] = dag.genesis
	blockTime := dag.genesis.Header().Timestamp
	for i := uint32(1); i < numBlocks; i++ {
		blockTime = blockTime.Add(time.Second)
		nodes[i] = newTestNode(setFromSlice(nodes[i-1]), blockVersion, 0, blockTime, netParams.K)
		dag.index.AddNode(nodes[i])
	}

	tests := []struct {
		blockNumber                 uint32
		expectedSecondsSinceGenesis int64
	}{
		{
			blockNumber:                 262,
			expectedSecondsSinceGenesis: 130,
		},
		{
			blockNumber:                 270,
			expectedSecondsSinceGenesis: 138,
		},
		{
			blockNumber:                 240,
			expectedSecondsSinceGenesis: 108,
		},
		{
			blockNumber:                 5,
			expectedSecondsSinceGenesis: 0,
		},
	}

	for _, test := range tests {
		secondsSinceGenesis := nodes[test.blockNumber].PastMedianTime(dag).Unix() - dag.genesis.Header().Timestamp.Unix()
		if secondsSinceGenesis != test.expectedSecondsSinceGenesis {
			t.Errorf("TestCalcPastMedianTime: expected past median time of block %v to be %v seconds from genesis but got %v", test.blockNumber, test.expectedSecondsSinceGenesis, secondsSinceGenesis)
		}
	}

}

// nodeHashes is a convenience function that returns the hashes for all of the
// passed indexes of the provided nodes.  It is used to construct expected hash
// slices in the tests.
func nodeHashes(nodes []*blockNode, indexes ...int) []*daghash.Hash {
	hashes := make([]*daghash.Hash, 0, len(indexes))
	for _, idx := range indexes {
		hashes = append(hashes, nodes[idx].hash)
	}
	return hashes
}

// testNoncePrng provides a deterministic prng for the nonce in generated fake
// nodes.  The ensures that the node have unique hashes.
var testNoncePrng = rand.New(rand.NewSource(0))

// chainedNodes returns the specified number of nodes constructed such that each
// subsequent node points to the previous one to create a chain.  The first node
// will point to the passed parent which can be nil if desired.
func chainedNodes(parents blockSet, numNodes int) []*blockNode {
	nodes := make([]*blockNode, numNodes)
	tips := parents
	for i := 0; i < numNodes; i++ {
		// This is invalid, but all that is needed is enough to get the
		// synthetic tests to work.
		header := wire.BlockHeader{
			Nonce:                testNoncePrng.Uint64(),
			HashMerkleRoot:       &daghash.ZeroHash,
			AcceptedIDMerkleRoot: &daghash.ZeroHash,
			UTXOCommitment:       &daghash.ZeroHash,
		}
		header.ParentHashes = tips.hashes()
		nodes[i] = newBlockNode(&header, tips, dagconfig.SimNetParams.K)
		tips = setFromSlice(nodes[i])
	}
	return nodes
}

// testTip is a convenience function to grab the tip of a chain of block nodes
// created via chainedNodes.
func testTip(nodes []*blockNode) *blockNode {
	return nodes[len(nodes)-1]
}

// TestHeightToHashRange ensures that fetching a range of block hashes by start
// height and end hash works as expected.
func TestHeightToHashRange(t *testing.T) {
	// Construct a synthetic block chain with a block index consisting of
	// the following structure.
	// 	genesis -> 1 -> 2 -> ... -> 15 -> 16  -> 17  -> 18
	// 	                              \-> 16a -> 17a -> 18a (unvalidated)
	tip := testTip
	blockDAG := newTestDAG(&dagconfig.SimNetParams)
	branch0Nodes := chainedNodes(setFromSlice(blockDAG.genesis), 18)
	branch1Nodes := chainedNodes(setFromSlice(branch0Nodes[14]), 3)
	for _, node := range branch0Nodes {
		blockDAG.index.SetStatusFlags(node, statusValid)
		blockDAG.index.AddNode(node)
	}
	for _, node := range branch1Nodes {
		if node.height < 18 {
			blockDAG.index.SetStatusFlags(node, statusValid)
		}
		blockDAG.index.AddNode(node)
	}
	blockDAG.virtual.SetTips(setFromSlice(tip(branch0Nodes)))

	tests := []struct {
		name        string
		startHeight uint64          // locator for requested inventory
		endHash     *daghash.Hash   // stop hash for locator
		maxResults  int             // max to locate, 0 = wire const
		hashes      []*daghash.Hash // expected located hashes
		expectError bool
	}{
		{
			name:        "blocks below tip",
			startHeight: 11,
			endHash:     branch0Nodes[14].hash,
			maxResults:  10,
			hashes:      nodeHashes(branch0Nodes, 10, 11, 12, 13, 14),
		},
		{
			name:        "blocks on main chain",
			startHeight: 15,
			endHash:     branch0Nodes[17].hash,
			maxResults:  10,
			hashes:      nodeHashes(branch0Nodes, 14, 15, 16, 17),
		},
		{
			name:        "blocks on stale chain",
			startHeight: 15,
			endHash:     branch1Nodes[1].hash,
			maxResults:  10,
			hashes: append(nodeHashes(branch0Nodes, 14),
				nodeHashes(branch1Nodes, 0, 1)...),
		},
		{
			name:        "invalid start height",
			startHeight: 19,
			endHash:     branch0Nodes[17].hash,
			maxResults:  10,
			expectError: true,
		},
		{
			name:        "too many results",
			startHeight: 1,
			endHash:     branch0Nodes[17].hash,
			maxResults:  10,
			expectError: true,
		},
		{
			name:        "unvalidated block",
			startHeight: 15,
			endHash:     branch1Nodes[2].hash,
			maxResults:  10,
			expectError: true,
		},
	}
	for _, test := range tests {
		hashes, err := blockDAG.HeightToHashRange(test.startHeight, test.endHash,
			test.maxResults)
		if err != nil {
			if !test.expectError {
				t.Errorf("%s: unexpected error: %v", test.name, err)
			}
			continue
		}

		if !reflect.DeepEqual(hashes, test.hashes) {
			t.Errorf("%s: unxpected hashes -- got %v, want %v",
				test.name, hashes, test.hashes)
		}
	}
}

// TestIntervalBlockHashes ensures that fetching block hashes at specified
// intervals by end hash works as expected.
func TestIntervalBlockHashes(t *testing.T) {
	// Construct a synthetic block chain with a block index consisting of
	// the following structure.
	// 	genesis -> 1 -> 2 -> ... -> 15 -> 16  -> 17  -> 18
	// 	                              \-> 16a -> 17a -> 18a (unvalidated)
	tip := testTip
	dag := newTestDAG(&dagconfig.SimNetParams)
	branch0Nodes := chainedNodes(setFromSlice(dag.genesis), 18)
	branch1Nodes := chainedNodes(setFromSlice(branch0Nodes[14]), 3)
	for _, node := range branch0Nodes {
		dag.index.SetStatusFlags(node, statusValid)
		dag.index.AddNode(node)
	}
	for _, node := range branch1Nodes {
		if node.height < 18 {
			dag.index.SetStatusFlags(node, statusValid)
		}
		dag.index.AddNode(node)
	}
	dag.virtual.SetTips(setFromSlice(tip(branch0Nodes)))

	tests := []struct {
		name        string
		endHash     *daghash.Hash
		interval    uint64
		hashes      []*daghash.Hash
		expectError bool
	}{
		{
			name:     "blocks on main chain",
			endHash:  branch0Nodes[17].hash,
			interval: 8,
			hashes:   nodeHashes(branch0Nodes, 7, 15),
		},
		{
			name:     "blocks on stale chain",
			endHash:  branch1Nodes[1].hash,
			interval: 8,
			hashes: append(nodeHashes(branch0Nodes, 7),
				nodeHashes(branch1Nodes, 0)...),
		},
		{
			name:     "no results",
			endHash:  branch0Nodes[17].hash,
			interval: 20,
			hashes:   []*daghash.Hash{},
		},
		{
			name:        "unvalidated block",
			endHash:     branch1Nodes[2].hash,
			interval:    8,
			expectError: true,
		},
	}
	for _, test := range tests {
		hashes, err := dag.IntervalBlockHashes(test.endHash, test.interval)
		if err != nil {
			if !test.expectError {
				t.Errorf("%s: unexpected error: %v", test.name, err)
			}
			continue
		}

		if !reflect.DeepEqual(hashes, test.hashes) {
			t.Errorf("%s: unxpected hashes -- got %v, want %v",
				test.name, hashes, test.hashes)
		}
	}
}

// TestApplyUTXOChangesErrors tests that
// dag.applyUTXOChanges panics when unexpected
// error occurs
func TestApplyUTXOChangesPanic(t *testing.T) {
	targetErrorMessage := "updateParents error"
	defer func() {
		if recover() == nil {
			t.Errorf("Got no panic on past UTXO error, while expected panic")
		}
	}()
	testErrorThroughPatching(
		t,
		targetErrorMessage,
		(*blockNode).updateParents,
		func(_ *blockNode, _ *virtualBlock, _ UTXOSet) error {
			return errors.New(targetErrorMessage)
		},
	)
}

// TestRestoreUTXOErrors tests all error-cases in restoreUTXO.
// The non-error-cases are tested in the more general tests.
func TestRestoreUTXOErrors(t *testing.T) {
	targetErrorMessage := "WithDiff error"
	testErrorThroughPatching(
		t,
		targetErrorMessage,
		(*FullUTXOSet).WithDiff,
		func(fus *FullUTXOSet, other *UTXODiff) (UTXOSet, error) {
			return nil, errors.New(targetErrorMessage)
		},
	)
}

func testErrorThroughPatching(t *testing.T, expectedErrorMessage string, targetFunction interface{}, replacementFunction interface{}) {
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

	// Create a new database and dag instance to run tests against.
	dag, teardownFunc, err := DAGSetup("testErrorThroughPatching", Config{
		DAGParams: &dagconfig.SimNetParams,
	})
	if err != nil {
		t.Fatalf("Failed to setup dag instance: %v", err)
	}
	defer teardownFunc()

	// Since we're not dealing with the real block DAG, set the coinbase
	// maturity to 0.
	dag.TestSetCoinbaseMaturity(0)

	guard := monkey.Patch(targetFunction, replacementFunction)
	defer guard.Unpatch()

	err = nil
	for i := 1; i < len(blocks); i++ {
		var isOrphan bool
		var delay time.Duration
		isOrphan, delay, err = dag.ProcessBlock(blocks[i], BFNone)
		if delay != 0 {
			t.Fatalf("ProcessBlock: block %d "+
				"is too far in the future", i)
		}
		if isOrphan {
			t.Fatalf("ProcessBlock incorrectly returned block %v "+
				"is an orphan\n", i)
		}
		if err != nil {
			break
		}
	}
	if err == nil {
		t.Errorf("ProcessBlock unexpectedly succeeded. "+
			"Expected: %s", expectedErrorMessage)
	}
	if !strings.Contains(err.Error(), expectedErrorMessage) {
		t.Errorf("ProcessBlock returned wrong error. "+
			"Want: %s, got: %s", expectedErrorMessage, err)
	}
}

func TestNew(t *testing.T) {
	// Create the root directory for test databases.
	if !FileExists(testDbRoot) {
		if err := os.MkdirAll(testDbRoot, 0700); err != nil {
			t.Fatalf("unable to create test db "+
				"root: %s", err)
		}
	}

	dbPath := filepath.Join(testDbRoot, "TestNew")
	_ = os.RemoveAll(dbPath)
	db, err := database.Create(testDbType, dbPath, blockDataNet)
	if err != nil {
		t.Fatalf("error creating db: %s", err)
	}
	defer func() {
		db.Close()
		os.RemoveAll(dbPath)
		os.RemoveAll(testDbRoot)
	}()
	config := &Config{
		DAGParams:  &dagconfig.SimNetParams,
		DB:         db,
		TimeSource: NewMedianTime(),
		SigCache:   txscript.NewSigCache(1000),
	}
	_, err = New(config)
	if err != nil {
		t.Fatalf("failed to create dag instance: %s", err)
	}

	config.SubnetworkID = &subnetworkid.SubnetworkID{0xff}
	_, err = New(config)
	expectedErrorMessage := fmt.Sprintf("Cannot start btcd with subnetwork ID %s because"+
		" its database is already built with subnetwork ID <nil>. If you"+
		" want to switch to a new database, please reset the"+
		" database by starting btcd with --reset-db flag", config.SubnetworkID)
	if err.Error() != expectedErrorMessage {
		t.Errorf("Unexpected error. Expected error '%s' but got '%s'", expectedErrorMessage, err)
	}
}

// TestAcceptingInInit makes sure that blocks that were stored but not
// yet fully processed do get correctly processed on DAG init. This may
// occur when the node shuts down improperly while a block is being
// validated.
func TestAcceptingInInit(t *testing.T) {
	// Create the root directory for test databases.
	if !FileExists(testDbRoot) {
		if err := os.MkdirAll(testDbRoot, 0700); err != nil {
			t.Fatalf("unable to create test db "+
				"root: %s", err)
		}
	}

	// Create a test database
	dbPath := filepath.Join(testDbRoot, "TestAcceptingInInit")
	_ = os.RemoveAll(dbPath)
	db, err := database.Create(testDbType, dbPath, blockDataNet)
	if err != nil {
		t.Fatalf("error creating db: %s", err)
	}
	defer func() {
		db.Close()
		os.RemoveAll(dbPath)
		os.RemoveAll(testDbRoot)
	}()

	// Create a DAG to add the test block into
	config := &Config{
		DAGParams:  &dagconfig.SimNetParams,
		DB:         db,
		TimeSource: NewMedianTime(),
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
	genesisNode := dag.index.LookupNode(genesisBlock.Hash())
	testNode := newBlockNode(&testBlock.MsgBlock().Header, setFromSlice(genesisNode), dag.dagParams.K)
	testNode.status = statusDataStored

	// Manually add the test block to the database
	err = db.Update(func(dbTx database.Tx) error {
		err := dbStoreBlock(dbTx, testBlock)
		if err != nil {
			return err
		}
		return dbStoreBlockNode(dbTx, testNode)
	})
	if err != nil {
		t.Fatalf("Failed to update block index: %s", err)
	}

	// Create a new DAG. We expect this DAG to process the
	// test node
	dag, err = New(config)
	if err != nil {
		t.Fatalf("failed to create dag instance: %s", err)
	}

	// Make sure that the test node's status is valid
	testNode = dag.index.LookupNode(testBlock.Hash())
	if testNode.status&statusValid == 0 {
		t.Fatalf("testNode is unexpectedly invalid")
	}
}

func TestConfirmations(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	params := dagconfig.SimNetParams
	params.K = 1
	dag, teardownFunc, err := DAGSetup("TestConfirmations", Config{
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
	if genesisConfirmations != 1 {
		t.Fatalf("TestConfirmations: unexpected confirmations for genesis block. Want: 1, got: %d", genesisConfirmations)
	}

	// Add a chain of blocks
	chainBlocks := make([]*blockNode, 5)
	chainBlocks[0] = dag.genesis
	buildNode := buildNodeGenerator(dag.dagParams.K, true)
	for i := uint32(1); i < 5; i++ {
		chainBlocks[i] = buildNode(setFromSlice(chainBlocks[i-1]))
		dag.virtual.AddTip(chainBlocks[i])
	}

	// Make sure that each one of the chain blocks has the expected confirmations number
	for i, node := range chainBlocks {
		confirmations, err := dag.blockConfirmations(node)
		if err != nil {
			t.Fatalf("TestConfirmations: confirmations for node 1 unexpectedly failed: %s", err)
		}

		expectedConfirmations := uint64(len(chainBlocks) - i)
		if confirmations != expectedConfirmations {
			t.Fatalf("TestConfirmations: unexpected confirmations for node 1. "+
				"Want: %d, got: %d", expectedConfirmations, confirmations)
		}
	}

	branchingBlocks := make([]*blockNode, 2)
	// Add two branching blocks
	branchingBlocks[0] = buildNode(setFromSlice(chainBlocks[1]))
	dag.virtual.AddTip(branchingBlocks[0])
	branchingBlocks[1] = buildNode(setFromSlice(branchingBlocks[0]))
	dag.virtual.AddTip(branchingBlocks[1])

	// Check that the genesis has a confirmations number == len(chainBlocks)
	genesisConfirmations, err = dag.blockConfirmations(dag.genesis)
	if err != nil {
		t.Fatalf("TestConfirmations: confirmations for genesis block unexpectedly failed: %s", err)
	}
	expectedGenesisConfirmations := uint64(len(chainBlocks))
	if genesisConfirmations != expectedGenesisConfirmations {
		t.Fatalf("TestConfirmations: unexpected confirmations for genesis block. "+
			"Want: %d, got: %d", expectedGenesisConfirmations, genesisConfirmations)
	}

	// Check that each of the blue tips has a confirmation number of 1, and each of the red tips has 0 confirmations.
	tips := dag.virtual.tips()
	for _, tip := range tips {
		tipConfirmations, err := dag.blockConfirmations(tip)
		if err != nil {
			t.Fatalf("TestConfirmations: confirmations for tip unexpectedly failed: %s", err)
		}
		expectedConfirmations := uint64(0)
		if tip == dag.selectedTip() {
			expectedConfirmations = 1
		}
		if tipConfirmations != expectedConfirmations {
			t.Fatalf("TestConfirmations: unexpected confirmations for tip. "+
				"Want: %d, got: %d", expectedConfirmations, tipConfirmations)
		}
	}

	// Generate 100 blocks to force the "main" chain to become red
	branchingChainTip := branchingBlocks[1]
	for i := uint32(0); i < 100; i++ {
		nextBranchingChainTip := buildNode(setFromSlice(branchingChainTip))
		dag.virtual.AddTip(nextBranchingChainTip)
		branchingChainTip = nextBranchingChainTip
	}

	// Make sure that a red block has confirmation number = 0
	redChainBlock := chainBlocks[3]
	redChainBlockConfirmations, err := dag.blockConfirmations(redChainBlock)
	if err != nil {
		t.Fatalf("TestConfirmations: confirmations for red chain block unexpectedly failed: %s", err)
	}
	if redChainBlockConfirmations != 0 {
		t.Fatalf("TestConfirmations: unexpected confirmations for red chain block. "+
			"Want: 0, got: %d", redChainBlockConfirmations)
	}

	// Make sure that the red tip has confirmation number = 0
	redChainTip := chainBlocks[len(chainBlocks)-1]
	redChainTipConfirmations, err := dag.blockConfirmations(redChainTip)
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
	params := dagconfig.SimNetParams
	params.K = 3
	dag, teardownFunc, err := DAGSetup("TestAcceptingBlock", Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()
	dag.TestSetCoinbaseMaturity(0)

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
	chainBlocks := make([]*blockNode, numChainBlocks)
	chainBlocks[0] = dag.genesis
	buildNode := buildNodeGenerator(dag.dagParams.K, true)
	for i := uint32(1); i <= numChainBlocks-1; i++ {
		chainBlocks[i] = buildNode(setFromSlice(chainBlocks[i-1]))
		dag.virtual.AddTip(chainBlocks[i])
	}

	// Make sure that each chain block (including the genesis) is accepted by its child
	for i, chainBlockNode := range chainBlocks[:len(chainBlocks)-1] {
		expectedAcceptingBlockNode := chainBlocks[i+1]
		chainAcceptingBlockNode, err := dag.acceptingBlock(chainBlockNode)
		if err != nil {
			t.Fatalf("TestAcceptingBlock: acceptingBlock for chain block %d unexpectedly failed: %s", i, err)
		}
		if expectedAcceptingBlockNode != chainAcceptingBlockNode {
			t.Fatalf("TestAcceptingBlock: unexpected acceptingBlock for chain block. "+
				"Want: %s, got: %s", expectedAcceptingBlockNode.hash, chainAcceptingBlockNode.hash)
		}
	}

	// Make sure that the selected tip doesn't have an accepting
	tipAcceptingBlock, err := dag.acceptingBlock(chainBlocks[len(chainBlocks)-1])
	if err != nil {
		t.Fatalf("TestAcceptingBlock: acceptingBlock for tip unexpectedly failed: %s", err)
	}
	if tipAcceptingBlock != nil {
		t.Fatalf("TestAcceptingBlock: unexpected acceptingBlock for tip. "+
			"Want: nil, got: %s", tipAcceptingBlock.hash)
	}

	// Generate side-chain of dag.dagParams.K + 1 blocks so its tip
	// will be in the blues of the virtual but in the anticone of
	// the selected tip.
	branchingChainTip := chainBlocks[len(chainBlocks)-2]
	for i := uint32(0); i < dag.dagParams.K+1; i++ {
		nextBranchingChainTip := buildNode(setFromSlice(branchingChainTip))
		dag.virtual.AddTip(nextBranchingChainTip)
		branchingChainTip = nextBranchingChainTip
	}

	// Make sure that branchingChainTip is not in the selected parent chain
	if dag.IsInSelectedParentChain(branchingChainTip.hash) {
		t.Fatalf("TestAcceptingBlock: branchingChainTip wasn't expected to be in the selected parent chain")
	}

	// Make sure that branchingChainTip is in the virtual blues
	isVirtualBlue := false
	for _, virtualBlue := range dag.virtual.blues {
		if branchingChainTip == virtualBlue {
			isVirtualBlue = true
			break
		}
	}
	if !isVirtualBlue {
		t.Fatalf("TestAcceptingBlock: redChainBlock was expected to be a virtual blue")
	}

	// Make sure that a block that is in the anticone of the selected tip and
	// in the blues of the virtual doesn't have an accepting block
	branchingChainTipAcceptionBlock, err := dag.acceptingBlock(branchingChainTip)
	if err != nil {
		t.Fatalf("TestAcceptingBlock: acceptingBlock for red chain block unexpectedly failed: %s", err)
	}
	if branchingChainTipAcceptionBlock != nil {
		t.Fatalf("TestAcceptingBlock: unexpected acceptingBlock for branchingChainTipAcceptionBlock. "+
			"Want: nil, got: %s", branchingChainTipAcceptionBlock.hash)
	}

	// Add K + 1 branching blocks
	intersectionBlock := chainBlocks[1]
	sideChainTip := buildNode(setFromSlice(intersectionBlock))
	i := uint32(0)
	for ; i < dag.dagParams.K+1; sideChainTip = buildNode(setFromSlice(sideChainTip)) {
		dag.virtual.AddTip(sideChainTip)
		i++
	}

	// Make sure that the accepting block of the parent of the branching block didn't change
	expectedAcceptingBlock := chainBlocks[2]
	intersectionAcceptingBlock, err := dag.acceptingBlock(intersectionBlock)
	if err != nil {
		t.Fatalf("TestAcceptingBlock: acceptingBlock for intersection block unexpectedly failed: %s", err)
	}
	if expectedAcceptingBlock != intersectionAcceptingBlock {
		t.Fatalf("TestAcceptingBlock: unexpected acceptingBlock for intersection block. "+
			"Want: %s, got: %s", expectedAcceptingBlock.hash, intersectionAcceptingBlock.hash)
	}

	// Make sure that a block that is found in the red set of the selected tip
	// doesn't have an accepting block
	newTip := buildNode(setFromSlice(sideChainTip, chainBlocks[len(chainBlocks)-1]))
	dag.virtual.AddTip(newTip)

	sideChainTipAcceptingBlock, err := dag.acceptingBlock(sideChainTip)
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
	params := dagconfig.SimNetParams
	params.K = 1
	dag, teardownFunc, err := DAGSetup("testFinalizeNodesBelowFinalityPoint", Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	blockVersion := int32(0x10000000)
	blockTime := dag.genesis.Header().Timestamp

	flushUTXODiffStore := func() {
		err := dag.db.Update(func(dbTx database.Tx) error {
			return dag.utxoDiffStore.flushToDB(dbTx)
		})
		if err != nil {
			t.Fatalf("Error flushing utxoDiffStore data to DB: %s", err)
		}
		dag.utxoDiffStore.clearDirtyEntries()
	}

	addNode := func(parent *blockNode) *blockNode {
		blockTime = blockTime.Add(time.Second)
		node := newTestNode(setFromSlice(parent), blockVersion, 0, blockTime, dag.dagParams.K)
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
	for i := 0; i <= finalityInterval*2; i++ {
		currentNode = addNode(currentNode)
		nodes = append(nodes, currentNode)
	}

	// Manually set the last finality point
	dag.lastFinalityPoint = nodes[finalityInterval-1]

	dag.finalizeNodesBelowFinalityPoint(deleteDiffData)
	flushUTXODiffStore()

	for _, node := range nodes[:finalityInterval-1] {
		if !node.isFinalized {
			t.Errorf("Node with blue score %d expected to be finalized", node.blueScore)
		}
		if _, ok := dag.utxoDiffStore.loaded[*node.hash]; deleteDiffData && ok {
			t.Errorf("The diff data of node with blue score %d should have been unloaded if deleteDiffData is %T", node.blueScore, deleteDiffData)
		} else if !deleteDiffData && !ok {
			t.Errorf("The diff data of node with blue score %d shouldn't have been unloaded if deleteDiffData is %T", node.blueScore, deleteDiffData)
		}
		if diffData, err := dag.utxoDiffStore.diffDataFromDB(node.hash); err != nil {
			t.Errorf("diffDataFromDB: %s", err)
		} else if deleteDiffData && diffData != nil {
			t.Errorf("The diff data of node with blue score %d should have been deleted from the database if deleteDiffData is %T", node.blueScore, deleteDiffData)
		} else if !deleteDiffData && diffData == nil {
			t.Errorf("The diff data of node with blue score %d shouldn't have been deleted from the database if deleteDiffData is %T", node.blueScore, deleteDiffData)
		}
	}

	for _, node := range nodes[finalityInterval-1:] {
		if node.isFinalized {
			t.Errorf("Node with blue score %d wasn't expected to be finalized", node.blueScore)
		}
		if _, ok := dag.utxoDiffStore.loaded[*node.hash]; !ok {
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
	params := dagconfig.SimNetParams
	dag, teardownFunc, err := DAGSetup("TestDAGIndexFailedStatus", Config{
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
	isOrphan, delay, err := dag.ProcessBlock(invalidBlock, BFNoPoWCheck)

	if _, ok := err.(RuleError); !ok {
		t.Fatalf("ProcessBlock: expected a rule error but got %s instead", err)
	}
	if delay != 0 {
		t.Fatalf("ProcessBlock: invalidBlock " +
			"is too far in the future")
	}
	if isOrphan {
		t.Fatalf("ProcessBlock incorrectly returned invalidBlock " +
			"is an orphan\n")
	}

	invalidBlockNode := dag.index.LookupNode(invalidBlock.Hash())
	if invalidBlockNode == nil {
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

	isOrphan, delay, err = dag.ProcessBlock(invalidBlockChild, BFNoPoWCheck)
	if rErr, ok := err.(RuleError); !ok || rErr.ErrorCode != ErrInvalidAncestorBlock {
		t.Fatalf("ProcessBlock: expected a rule error but got %s instead", err)
	}
	if delay != 0 {
		t.Fatalf("ProcessBlock: invalidBlockChild " +
			"is too far in the future")
	}
	if isOrphan {
		t.Fatalf("ProcessBlock incorrectly returned invalidBlockChild " +
			"is an orphan\n")
	}
	invalidBlockChildNode := dag.index.LookupNode(invalidBlockChild.Hash())
	if invalidBlockChildNode == nil {
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

	isOrphan, delay, err = dag.ProcessBlock(invalidBlockGrandChild, BFNoPoWCheck)
	if rErr, ok := err.(RuleError); !ok || rErr.ErrorCode != ErrInvalidAncestorBlock {
		t.Fatalf("ProcessBlock: expected a rule error but got %s instead", err)
	}
	if delay != 0 {
		t.Fatalf("ProcessBlock: invalidBlockGrandChild " +
			"is too far in the future")
	}
	if isOrphan {
		t.Fatalf("ProcessBlock incorrectly returned invalidBlockGrandChild " +
			"is an orphan\n")
	}
	invalidBlockGrandChildNode := dag.index.LookupNode(invalidBlockGrandChild.Hash())
	if invalidBlockGrandChildNode == nil {
		t.Fatalf("invalidBlockGrandChild wasn't added to the block index as expected")
	}
	if invalidBlockGrandChildNode.status&statusInvalidAncestor != statusInvalidAncestor {
		t.Fatalf("invalidBlockGrandChildNode status to have %b flags raised (got %b)", statusInvalidAncestor, invalidBlockGrandChildNode.status)
	}
}
