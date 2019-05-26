// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"errors"
	"fmt"
	"math"
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
	"github.com/daglabs/btcd/util/txsort"
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
		blockTmp, err := loadBlocks(file)
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

	// Since we're not dealing with the real block DAG, set the block reward
	// maturity to 1.
	dag.TestSetBlockRewardMaturity(1)

	for i := 1; i < len(blocks); i++ {
		isOrphan, err := dag.ProcessBlock(blocks[i], BFNone)
		if err != nil {
			t.Fatalf("ProcessBlock fail on block %v: %v\n", i, err)
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
		blockTmp, err := loadBlocks(file)
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

	// Since we're not dealing with the real block DAG, set the block reward
	// maturity to 1.
	dag.TestSetBlockRewardMaturity(1)

	for i := 1; i < len(blocks); i++ {
		isOrphan, err := dag.ProcessBlock(blocks[i], BFNone)
		if err != nil {
			t.Fatalf("ProcessBlock fail on block %v: %v\n", i, err)
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
		blockTmp, err := loadBlocks(file)
		if err != nil {
			t.Fatalf("Error loading file: %v\n", err)
		}
		blocks = append(blocks, blockTmp...)
	}
	isOrphan, err := dag.ProcessBlock(blocks[6], BFNone)

	// Block 3C should fail to connect since its parents are related. (It points to 1 and 2, and 1 is the parent of 2)
	if err == nil {
		t.Fatalf("ProcessBlock for block 3C has no error when expected to have an error\n")
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
		blockTmp, err := loadBlocks(file)
		if err != nil {
			t.Fatalf("Error loading file: %v\n", err)
		}
		blocks = append(blocks, blockTmp...)
	}
	isOrphan, err = dag.ProcessBlock(blocks[7], BFNone)

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
	if isOrphan {
		t.Fatalf("ProcessBlock incorrectly returned block 3D " +
			"is an orphan\n")
	}

	// Insert an orphan block.
	isOrphan, err = dag.ProcessBlock(util.NewBlock(&Block100000), BFNoPoWCheck)
	if err != nil {
		t.Fatalf("Unable to process block: %v", err)
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
		{hash: "3f2ded16b7115e69a48cee5f4be743ff23ad8d41da16d059c38cc83d14459863", want: true},

		// Block 100000 should be present (as an orphan).
		{hash: "4e530ee9f967de3b2cd47ac5cd00109bb9ed7b0e30a60485c94badad29ecb4ce", want: true},

		// Random hashes should not be available.
		{hash: "123", want: false},
	}

	for i, test := range tests {
		hash, err := daghash.NewHashFromStr(test.hash)
		if err != nil {
			t.Fatalf("NewHashFromStr: %v", err)
		}

		result, err := dag.HaveBlock(hash)
		if err != nil {
			t.Fatalf("HaveBlock #%d unexpected error: %v", i, err)
		}
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
	msgTx := wire.NewNativeMsgTx(wire.TxVersion, nil, []*wire.TxOut{{PkScript: nil, Value: 10}})
	targetTx := util.NewTx(msgTx)
	utxoSet := NewFullUTXOSet()
	utxoSet.AddTx(targetTx.MsgTx(), uint64(numBlocksToGenerate)-4)

	// Create a utxo that spends the fake utxo created above for use in the
	// transactions created in the tests.  It has an age of 4 blocks.  Note
	// that the sequence lock heights are always calculated from the same
	// point of view that they were originally calculated from for a given
	// utxo.  That is to say, the height prior to it.
	utxo := wire.OutPoint{
		TxID:  *targetTx.ID(),
		Index: 0,
	}
	prevUtxoChainHeight := uint64(numBlocksToGenerate) - 4

	// Obtain the median time past from the PoV of the input created above.
	// The MTP for the input is the MTP from the PoV of the block *prior*
	// to the one that included it.
	medianTime := node.RelativeAncestor(5).PastMedianTime().Unix()

	// The median time calculated from the PoV of the best block in the
	// test chain.  For unconfirmed inputs, this value will be used since
	// the MTP will be calculated from the PoV of the yet-to-be-mined
	// block.
	nextMedianTime := node.PastMedianTime().Unix()
	nextBlockChainHeight := int32(numBlocksToGenerate) + 1

	// Add an additional transaction which will serve as our unconfirmed
	// output.
	unConfTx := wire.NewNativeMsgTx(wire.TxVersion, nil, []*wire.TxOut{{PkScript: nil, Value: 5}})
	unConfUtxo := wire.OutPoint{
		TxID:  *unConfTx.TxID(),
		Index: 0,
	}

	utxoSet.AddTx(unConfTx, UnminedChainHeight)

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
			tx:      wire.NewNativeMsgTx(1, []*wire.TxIn{{PreviousOutPoint: utxo, Sequence: wire.MaxTxInSequenceNum}}, nil),
			utxoSet: utxoSet,
			want: &SequenceLock{
				Seconds:          -1,
				BlockChainHeight: -1,
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
			tx:      wire.NewNativeMsgTx(1, []*wire.TxIn{{PreviousOutPoint: utxo, Sequence: LockTimeToSequence(true, 2)}}, nil),
			utxoSet: utxoSet,
			want: &SequenceLock{
				Seconds:          medianTime - 1,
				BlockChainHeight: -1,
			},
		},
		// A transaction with a single input whose lock time is
		// expressed in seconds.  The number of seconds should be 1023
		// seconds after the median past time of the last block in the
		// chain.
		{
			name:    "single input, 1023 seconds after median time",
			tx:      wire.NewNativeMsgTx(1, []*wire.TxIn{{PreviousOutPoint: utxo, Sequence: LockTimeToSequence(true, 1024)}}, nil),
			utxoSet: utxoSet,
			want: &SequenceLock{
				Seconds:          medianTime + 1023,
				BlockChainHeight: -1,
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
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(true, 2560),
				}, {
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(false, 4),
				}, {
					PreviousOutPoint: utxo,
					Sequence: LockTimeToSequence(false, 5) |
						wire.SequenceLockTimeDisabled,
				}},
				nil),
			utxoSet: utxoSet,
			want: &SequenceLock{
				Seconds:          medianTime + (5 << wire.SequenceLockTimeGranularity) - 1,
				BlockChainHeight: int64(prevUtxoChainHeight) + 3,
			},
		},
		// Transaction with a single input.  The input's sequence number
		// encodes a relative lock-time in blocks (3 blocks).  The
		// sequence lock should  have a value of -1 for seconds, but a
		// height of 2 meaning it can be included at height 3.
		{
			name:    "single input, lock-time in blocks",
			tx:      wire.NewNativeMsgTx(1, []*wire.TxIn{{PreviousOutPoint: utxo, Sequence: LockTimeToSequence(false, 3)}}, nil),
			utxoSet: utxoSet,
			want: &SequenceLock{
				Seconds:          -1,
				BlockChainHeight: int64(prevUtxoChainHeight) + 2,
			},
		},
		// A transaction with two inputs with lock times expressed in
		// seconds.  The selected sequence lock value for seconds should
		// be the time further in the future.
		{
			name: "two inputs, lock-times in seconds",
			tx: wire.NewNativeMsgTx(1, []*wire.TxIn{{
				PreviousOutPoint: utxo,
				Sequence:         LockTimeToSequence(true, 5120),
			}, {
				PreviousOutPoint: utxo,
				Sequence:         LockTimeToSequence(true, 2560),
			}}, nil),
			utxoSet: utxoSet,
			want: &SequenceLock{
				Seconds:          medianTime + (10 << wire.SequenceLockTimeGranularity) - 1,
				BlockChainHeight: -1,
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
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(false, 1),
				}, {
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(false, 11),
				}},
				nil),
			utxoSet: utxoSet,
			want: &SequenceLock{
				Seconds:          -1,
				BlockChainHeight: int64(prevUtxoChainHeight) + 10,
			},
		},
		// A transaction with multiple inputs.  Two inputs are time
		// based, and the other two are block based. The lock lying
		// further into the future for both inputs should be chosen.
		{
			name: "four inputs, two lock-times in time, two lock-times in blocks",
			tx: wire.NewNativeMsgTx(1,
				[]*wire.TxIn{{
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(true, 2560),
				}, {
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(true, 6656),
				}, {
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(false, 3),
				}, {
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(false, 9),
				}},
				nil),
			utxoSet: utxoSet,
			want: &SequenceLock{
				Seconds:          medianTime + (13 << wire.SequenceLockTimeGranularity) - 1,
				BlockChainHeight: int64(prevUtxoChainHeight) + 8,
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
			tx:      wire.NewNativeMsgTx(1, []*wire.TxIn{{PreviousOutPoint: unConfUtxo, Sequence: LockTimeToSequence(false, 2)}}, nil),
			utxoSet: utxoSet,
			mempool: true,
			want: &SequenceLock{
				Seconds:          -1,
				BlockChainHeight: int64(nextBlockChainHeight) + 1,
			},
		},
		// A transaction with a single unconfirmed input.  The input has
		// a time based lock, so the lock time should be based off the
		// MTP of the *next* block.
		{
			name:    "single input, unconfirmed, lock-time in seoncds",
			tx:      wire.NewNativeMsgTx(1, []*wire.TxIn{{PreviousOutPoint: unConfUtxo, Sequence: LockTimeToSequence(true, 1024)}}, nil),
			utxoSet: utxoSet,
			mempool: true,
			want: &SequenceLock{
				Seconds:          nextMedianTime + 1023,
				BlockChainHeight: -1,
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
		if seqLock.BlockChainHeight != test.want.BlockChainHeight {
			t.Fatalf("test '%s' got chain-height of %v want chain-height of %v ",
				test.name, seqLock.BlockChainHeight, test.want.BlockChainHeight)
		}
	}
}

func TestCalcPastMedianTime(t *testing.T) {
	netParams := &dagconfig.SimNetParams

	blockVersion := int32(0x10000000)

	dag := newTestDAG(netParams)
	numBlocks := uint32(60)
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
			blockNumber:                 50,
			expectedSecondsSinceGenesis: 25,
		},
		{
			blockNumber:                 59,
			expectedSecondsSinceGenesis: 34,
		},
		{
			blockNumber:                 40,
			expectedSecondsSinceGenesis: 15,
		},
		{
			blockNumber:                 5,
			expectedSecondsSinceGenesis: 0,
		},
	}

	for _, test := range tests {
		secondsSinceGenesis := nodes[test.blockNumber].PastMedianTime().Unix() - dag.genesis.Header().Timestamp.Unix()
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
		blockTmp, err := loadBlocks(file)
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

	// Since we're not dealing with the real block chain, set the block reward
	// maturity to 1.
	dag.TestSetBlockRewardMaturity(1)

	guard := monkey.Patch(targetFunction, replacementFunction)
	defer guard.Unpatch()

	err = nil
	for i := 1; i < len(blocks); i++ {
		var isOrphan bool
		isOrphan, err = dag.ProcessBlock(blocks[i], BFNone)
		if isOrphan {
			t.Fatalf("ProcessBlock incorrectly returned block %v "+
				"is an orphan\n", i)
		}
		if err != nil {
			fmt.Printf("ERROR %v\n", err)
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
	if !fileExists(testDbRoot) {
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

func TestValidateFeeTransaction(t *testing.T) {
	params := dagconfig.SimNetParams
	params.K = 1
	dag, teardownFunc, err := DAGSetup("TestValidateFeeTransaction", Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()
	extraNonce := int64(0)
	createCoinbase := func(pkScript []byte) *wire.MsgTx {
		extraNonce++
		cbTx, err := createCoinbaseTxForTest(dag.Height()+1, 2, extraNonce, &params)
		if err != nil {
			t.Fatalf("createCoinbaseTxForTest: %v", err)
		}
		if pkScript != nil {
			cbTx.TxOut[0].PkScript = pkScript
		}
		return cbTx
	}

	var flags BehaviorFlags
	flags |= BFFastAdd | BFNoPoWCheck

	buildBlock := func(blockName string, parentHashes []*daghash.Hash, transactions []*wire.MsgTx, expectedErrorCode ErrorCode) *wire.MsgBlock {
		utilTxs := make([]*util.Tx, len(transactions))
		for i, tx := range transactions {
			utilTxs[i] = util.NewTx(tx)
		}

		newVirtual, err := GetVirtualFromParentsForTest(dag, parentHashes)
		if err != nil {
			t.Fatalf("block %v: unexpected error when setting virtual for test: %v", blockName, err)
		}
		oldVirtual := SetVirtualForTest(dag, newVirtual)
		acceptedIDMerkleRoot, err := dag.NextAcceptedIDMerkleRoot()
		if err != nil {
			t.Fatalf("block %v: unexpected error when getting next acceptIDMerkleRoot: %v", blockName, err)
		}
		SetVirtualForTest(dag, oldVirtual)

		daghash.Sort(parentHashes)
		msgBlock := &wire.MsgBlock{
			Header: wire.BlockHeader{
				Bits:                 dag.genesis.Header().Bits,
				ParentHashes:         parentHashes,
				HashMerkleRoot:       BuildHashMerkleTreeStore(utilTxs).Root(),
				AcceptedIDMerkleRoot: acceptedIDMerkleRoot,
				UTXOCommitment:       &daghash.ZeroHash,
			},
			Transactions: transactions,
		}
		block := util.NewBlock(msgBlock)
		isOrphan, err := dag.ProcessBlock(block, flags)
		if expectedErrorCode != 0 {
			checkResult := checkRuleError(err, RuleError{
				ErrorCode: expectedErrorCode,
			})
			if checkResult != nil {
				t.Errorf("block %v: unexpected error code: %v", blockName, checkResult)
			}
		} else {
			if err != nil {
				t.Fatalf("block %v: unexpected error: %v", blockName, err)
			}
			if isOrphan {
				t.Errorf("block %v unexpectely got orphaned", blockName)
			}
		}
		return msgBlock
	}

	cb1 := createCoinbase(nil)
	blockWithExtraFeeTxTransactions := []*wire.MsgTx{
		cb1,
		{ // Fee Transaction
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						TxID:  daghash.TxID(*dag.genesis.hash),
						Index: math.MaxUint32,
					},
					Sequence: wire.MaxTxInSequenceNum,
				},
			},
			SubnetworkID: *subnetworkid.SubnetworkIDNative,
		},
		{ // Extra Fee Transaction
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						TxID:  daghash.TxID(*dag.genesis.hash),
						Index: math.MaxUint32,
					},
					Sequence: wire.MaxTxInSequenceNum,
				},
			},
			SubnetworkID: *subnetworkid.SubnetworkIDNative,
		},
	}
	buildBlock("blockWithExtraFeeTx", []*daghash.Hash{dag.genesis.hash}, blockWithExtraFeeTxTransactions, ErrMultipleFeeTransactions)

	block1Txs := []*wire.MsgTx{
		cb1,
		{ // Fee Transaction
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						TxID:  daghash.TxID(*dag.genesis.hash),
						Index: math.MaxUint32,
					},
					Sequence: wire.MaxTxInSequenceNum,
				},
			},
			SubnetworkID: *subnetworkid.SubnetworkIDNative,
		},
	}
	block1 := buildBlock("block1", []*daghash.Hash{dag.genesis.hash}, block1Txs, 0)

	cb1A := createCoinbase(nil)
	block1ATxs := []*wire.MsgTx{
		cb1A,
		{ // Fee Transaction
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						TxID:  daghash.TxID(*dag.genesis.hash),
						Index: math.MaxUint32,
					},
					Sequence: wire.MaxTxInSequenceNum,
				},
			},
			SubnetworkID: *subnetworkid.SubnetworkIDNative,
		},
	}
	block1A := buildBlock("block1A", []*daghash.Hash{dag.genesis.hash}, block1ATxs, 0)

	block1AChildCbPkScript, err := payToPubKeyHashScript((&[20]byte{0x1A, 0xC0})[:])
	if err != nil {
		t.Fatalf("payToPubKeyHashScript: %v", err)
	}
	cb1AChild := createCoinbase(block1AChildCbPkScript)
	block1AChildTxs := []*wire.MsgTx{
		cb1AChild,
		{ // Fee Transaction
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						TxID:  daghash.TxID(*block1A.BlockHash()),
						Index: math.MaxUint32,
					},
					Sequence: wire.MaxTxInSequenceNum,
				},
			},
			SubnetworkID: *subnetworkid.SubnetworkIDNative,
		},
		{
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						TxID:  *cb1A.TxID(),
						Index: 0,
					},
					Sequence: wire.MaxTxInSequenceNum,
				},
			},
			TxOut: []*wire.TxOut{
				{
					PkScript: OpTrueScript,
					Value:    1,
				},
			},
			SubnetworkID: *subnetworkid.SubnetworkIDNative,
		},
	}
	block1AChild := buildBlock("block1AChild", []*daghash.Hash{block1A.BlockHash()}, block1AChildTxs, 0)

	cb2 := createCoinbase(nil)
	block2Txs := []*wire.MsgTx{
		cb2,
		{ // Fee Transaction
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						TxID:  daghash.TxID(*block1.BlockHash()),
						Index: math.MaxUint32,
					},
					Sequence: wire.MaxTxInSequenceNum,
				},
			},
			SubnetworkID: *subnetworkid.SubnetworkIDNative,
		},
	}
	block2 := buildBlock("block2", []*daghash.Hash{block1.BlockHash()}, block2Txs, 0)

	cb3 := createCoinbase(nil)
	block3Txs := []*wire.MsgTx{
		cb3,
		{ // Fee Transaction
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						TxID:  daghash.TxID(*block2.BlockHash()),
						Index: math.MaxUint32,
					},
					Sequence: wire.MaxTxInSequenceNum,
				},
			},
			SubnetworkID: *subnetworkid.SubnetworkIDNative,
		},
	}
	block3 := buildBlock("block3", []*daghash.Hash{block2.BlockHash()}, block3Txs, 0)

	block4CbPkScript, err := payToPubKeyHashScript((&[20]byte{0x40})[:])
	if err != nil {
		t.Fatalf("payToPubKeyHashScript: %v", err)
	}

	cb4 := createCoinbase(block4CbPkScript)
	block4Txs := []*wire.MsgTx{
		cb4,
		{ // Fee Transaction
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						TxID:  daghash.TxID(*block3.BlockHash()),
						Index: math.MaxUint32,
					},
					Sequence: wire.MaxTxInSequenceNum,
				},
			},
			SubnetworkID: *subnetworkid.SubnetworkIDNative,
		},
		{
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						TxID:  *cb3.TxID(),
						Index: 0,
					},
					Sequence: wire.MaxTxInSequenceNum,
				},
			},
			TxOut: []*wire.TxOut{
				{
					PkScript: OpTrueScript,
					Value:    1,
				},
			},
			SubnetworkID: *subnetworkid.SubnetworkIDNative,
		},
	}
	block4 := buildBlock("block4", []*daghash.Hash{block3.BlockHash()}, block4Txs, 0)

	block4ACbPkScript, err := payToPubKeyHashScript((&[20]byte{0x4A})[:])
	if err != nil {
		t.Fatalf("payToPubKeyHashScript: %v", err)
	}
	cb4A := createCoinbase(block4ACbPkScript)
	block4ATxs := []*wire.MsgTx{
		cb4A,
		{ // Fee Transaction
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						TxID:  daghash.TxID(*block3.BlockHash()),
						Index: math.MaxUint32,
					},
					Sequence: wire.MaxTxInSequenceNum,
				},
			},
			SubnetworkID: *subnetworkid.SubnetworkIDNative,
		},
		{
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						TxID:  *cb3.TxID(),
						Index: 1,
					},
					Sequence: wire.MaxTxInSequenceNum,
				},
			},
			TxOut: []*wire.TxOut{
				{
					PkScript: OpTrueScript,
					Value:    1,
				},
			},
			SubnetworkID: *subnetworkid.SubnetworkIDNative,
		},
	}
	block4A := buildBlock("block4A", []*daghash.Hash{block3.BlockHash()}, block4ATxs, 0)

	cb5 := createCoinbase(nil)
	feeInOuts := map[daghash.Hash]*struct {
		txIn  *wire.TxIn
		txOut *wire.TxOut
	}{
		*block4.BlockHash(): {
			txIn: &wire.TxIn{
				PreviousOutPoint: wire.OutPoint{
					TxID:  daghash.TxID(*block4.BlockHash()),
					Index: math.MaxUint32,
				},
				Sequence: wire.MaxTxInSequenceNum,
			},
			txOut: &wire.TxOut{
				PkScript: block4CbPkScript,
				Value:    cb3.TxOut[0].Value - 1,
			},
		},
		*block4A.BlockHash(): {
			txIn: &wire.TxIn{
				PreviousOutPoint: wire.OutPoint{
					TxID:  daghash.TxID(*block4A.BlockHash()),
					Index: math.MaxUint32,
				},
				Sequence: wire.MaxTxInSequenceNum,
			},
			txOut: &wire.TxOut{
				PkScript: block4ACbPkScript,
				Value:    cb3.TxOut[1].Value - 1,
			},
		},
	}

	txIns := []*wire.TxIn{}
	txOuts := []*wire.TxOut{}
	for hash := range feeInOuts {
		txIns = append(txIns, feeInOuts[hash].txIn)
		txOuts = append(txOuts, feeInOuts[hash].txOut)
	}
	block5FeeTx := wire.NewNativeMsgTx(1, txIns, txOuts)
	sortedBlock5FeeTx := txsort.Sort(block5FeeTx)

	block5Txs := []*wire.MsgTx{cb5, sortedBlock5FeeTx}

	block5ParentHashes := []*daghash.Hash{block4.BlockHash(), block4A.BlockHash()}
	buildBlock("block5", block5ParentHashes, block5Txs, 0)

	sortedBlock5FeeTx.TxIn[0], sortedBlock5FeeTx.TxIn[1] = sortedBlock5FeeTx.TxIn[1], sortedBlock5FeeTx.TxIn[0]
	buildBlock("block5WrongOrder", block5ParentHashes, block5Txs, ErrBadFeeTransaction)

	block5FeeTxWith1Achild := block5FeeTx.Copy()

	block5FeeTxWith1Achild.AddTxIn(&wire.TxIn{
		PreviousOutPoint: wire.OutPoint{
			TxID:  daghash.TxID(*block1AChild.BlockHash()),
			Index: math.MaxUint32,
		},
		Sequence: wire.MaxTxInSequenceNum,
	})
	block5FeeTxWith1Achild.AddTxOut(&wire.TxOut{
		PkScript: block1AChildCbPkScript,
		Value:    cb1AChild.TxOut[0].Value - 1,
	})

	sortedBlock5FeeTxWith1Achild := txsort.Sort(block5FeeTxWith1Achild)

	block5Txs[1] = sortedBlock5FeeTxWith1Achild
	buildBlock("block5WithRedBlockFees", block5ParentHashes, block5Txs, ErrBadFeeTransaction)

	block5FeeTxWithWrongFees := block5FeeTx.Copy()
	block5FeeTxWithWrongFees.TxOut[0].Value--
	sortedBlock5FeeTxWithWrongFees := txsort.Sort(block5FeeTxWithWrongFees)
	block5Txs[1] = sortedBlock5FeeTxWithWrongFees
	buildBlock("block5WithRedBlockFees", block5ParentHashes, block5Txs, ErrBadFeeTransaction)
}

func TestConfirmations(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	dag, teardownFunc, err := DAGSetup("TestBlockCount", Config{
		DAGParams: &dagconfig.SimNetParams,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()
	dag.TestSetBlockRewardMaturity(1)

	// Check that the genesis block of a DAG with only the genesis block in it has confirmations = 1.
	genesisConfirmations, err := dag.blockConfirmations(dag.genesis)
	if err != nil {
		t.Fatalf("TestConfirmations: confirmations for genesis block unexpectedly failed: %s", err)
	}
	if genesisConfirmations != 1 {
		t.Fatalf("TestConfirmations: unexpected confirmations for genesis block. Want: 1, got: %d", genesisConfirmations)
	}

	processBlocks := func(blocks []*util.Block) {
		for _, block := range blocks {
			isOrphan, err := dag.ProcessBlock(block, BFNone)
			if err != nil {
				t.Fatalf("ProcessBlock fail on block %s: %v\n", block.Hash(), err)
			}
			if isOrphan {
				t.Fatalf("ProcessBlock incorrectly returned block %s is an orphan\n", block.Hash())
			}
		}
	}

	// Add a chain of blocks
	loadedBlocks, err := loadBlocks("blk_0_to_4.dat")
	if err != nil {
		t.Fatalf("Error loading file: %v\n", err)
	}
	chainBlocks := loadedBlocks[1:]
	processBlocks(chainBlocks)

	// Make sure that each one of the chain blocks has the expected confirmations number
	for i, block := range chainBlocks {
		node := dag.index.LookupNode(block.Hash())
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

	// Add a branching block
	loadedBlocks, err = loadBlocks("blk_3B.dat")
	if err != nil {
		t.Fatalf("Error loading file: %v\n", err)
	}
	processBlocks(loadedBlocks)

	// Check that the genesis has a confirmations number == blockCount
	genesisConfirmations, err = dag.blockConfirmations(dag.genesis)
	if err != nil {
		t.Fatalf("TestConfirmations: confirmations for genesis block unexpectedly failed: %s", err)
	}
	expectedGenesisConfirmations := dag.blockCount
	if genesisConfirmations != expectedGenesisConfirmations {
		t.Fatalf("TestConfirmations: unexpected confirmations for genesis block. "+
			"Want: %d, got: %d", expectedGenesisConfirmations, genesisConfirmations)
	}

	// Check that each of the tips had a confirmation number of 1.
	tips := dag.virtual.tips()
	for _, tip := range tips {
		tipConfirmations, err := dag.blockConfirmations(tip)
		if err != nil {
			t.Fatalf("TestConfirmations: confirmations for tip unexpectedly failed: %s", err)
		}
		if tipConfirmations != 1 {
			t.Fatalf("TestConfirmations: unexpected confirmations for tip. "+
				"Want: 1, got: %d", tipConfirmations)
		}
	}

	// Generate K blocks to force the "main" chain to become red
	nodeGenerator := buildNodeGenerator(dag.dagParams.K, false)
	branchingChainTip := dag.index.LookupNode(loadedBlocks[0].Hash())
	for i := uint32(0); i < dag.dagParams.K; i++ {
		nextBranchingChainTip := nodeGenerator(setFromSlice(branchingChainTip))
		dag.virtual.AddTip(nextBranchingChainTip)
		branchingChainTip = nextBranchingChainTip
	}

	// Make sure that a red block has confirmation number = 0
	redChainBlock := dag.index.LookupNode(chainBlocks[3].Hash())
	redChainBlockConfirmations, err := dag.blockConfirmations(redChainBlock)
	if err != nil {
		t.Fatalf("TestConfirmations: confirmations for red chain block unexpectedly failed: %s", err)
	}
	if redChainBlockConfirmations != 0 {
		t.Fatalf("TestConfirmations: unexpected confirmations for red chain block. "+
			"Want: 0, got: %d", redChainBlockConfirmations)
	}

	// Make sure that the red tip has confirmation number = 0
	redChainTip := dag.index.LookupNode(chainBlocks[len(chainBlocks)-1].Hash())
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
	dag, teardownFunc, err := DAGSetup("TestAcceptingBlock", Config{
		DAGParams: &dagconfig.SimNetParams,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()
	dag.TestSetBlockRewardMaturity(1)

	// Check that the genesis block of a DAG with only the genesis block in it is accepted by the virtual.
	genesisAcceptingBlock, err := dag.acceptingBlock(dag.genesis)
	if err != nil {
		t.Fatalf("TestAcceptingBlock: acceptingBlock for genesis block unexpectedly failed: %s", err)
	}
	if genesisAcceptingBlock != &dag.virtual.blockNode {
		t.Fatalf("TestAcceptingBlock: unexpected acceptingBlock for genesis block. "+
			"Want: virtual, got: %s", genesisAcceptingBlock.hash)
	}

	processBlocks := func(blocks []*util.Block) {
		for _, block := range blocks {
			isOrphan, err := dag.ProcessBlock(block, BFNone)
			if err != nil {
				t.Fatalf("ProcessBlock fail on block %s: %v\n", block.Hash(), err)
			}
			if isOrphan {
				t.Fatalf("ProcessBlock incorrectly returned block %s is an orphan\n", block.Hash())
			}
		}
	}

	// Add a chain of blocks
	chainBlocks, err := loadBlocks("blk_0_to_4.dat")
	if err != nil {
		t.Fatalf("Error loading file: %v\n", err)
	}
	processBlocks(chainBlocks[1:])

	// Make sure that each chain block (including the genesis) is accepted by its child
	for i, chainBlock := range chainBlocks[:1] {
		expectedAcceptingBlock := chainBlocks[i+1]
		expectedAcceptingBlockNode := dag.index.LookupNode(expectedAcceptingBlock.Hash())

		chainBlockNode := dag.index.LookupNode(chainBlock.Hash())
		chainAcceptingBlockNode, err := dag.acceptingBlock(chainBlockNode)
		if err != nil {
			t.Fatalf("TestAcceptingBlock: acceptingBlock for chain block unexpectedly failed: %s", err)
		}
		if expectedAcceptingBlockNode != chainAcceptingBlockNode {
			t.Fatalf("TestAcceptingBlock: unexpected acceptingBlock for chain block. "+
				"Want: %s, got: %s", expectedAcceptingBlockNode.hash, chainAcceptingBlockNode.hash)
		}
	}

	// Add a branching block
	branchingBlock, err := loadBlocks("blk_3B.dat")
	if err != nil {
		t.Fatalf("Error loading file: %v\n", err)
	}
	processBlocks(branchingBlock)

	// Make sure that the accepting block of the parent of the branching block didn't change
	expectedAcceptingBlock := dag.index.LookupNode(chainBlocks[3].Hash())
	intersectionBlock := dag.index.LookupNode(chainBlocks[2].Hash())
	intersectionAcceptingBlock, err := dag.acceptingBlock(intersectionBlock)
	if err != nil {
		t.Fatalf("TestAcceptingBlock: acceptingBlock for intersection block unexpectedly failed: %s", err)
	}
	if expectedAcceptingBlock != intersectionAcceptingBlock {
		t.Fatalf("TestAcceptingBlock: unexpected acceptingBlock for intersection block. "+
			"Want: %s, got: %s", expectedAcceptingBlock.hash, intersectionAcceptingBlock.hash)
	}

	// Make sure that the accepting block of all the tips in the virtual block
	for _, tip := range dag.virtual.tips() {
		tipAcceptingBlock, err := dag.acceptingBlock(tip)
		if err != nil {
			t.Fatalf("TestAcceptingBlock: acceptingBlock for tip unexpectedly failed: %s", err)
		}
		if tipAcceptingBlock != &dag.virtual.blockNode {
			t.Fatalf("TestAcceptingBlock: unexpected acceptingBlock for tip. "+
				"Want: Virtual, got: %s", tipAcceptingBlock.hash)
		}
	}

	// Generate K blocks to force the "main" chain to become red
	nodeGenerator := buildNodeGenerator(dag.dagParams.K, false)
	branchingChainTip := dag.index.LookupNode(branchingBlock[0].Hash())
	for i := uint32(0); i < dag.dagParams.K; i++ {
		nextBranchingChainTip := nodeGenerator(setFromSlice(branchingChainTip))
		dag.virtual.AddTip(nextBranchingChainTip)
		branchingChainTip = nextBranchingChainTip
	}

	// Make sure that a red block returns nil
	redChainBlock := dag.index.LookupNode(chainBlocks[3].Hash())
	redChainBlockAcceptionBlock, err := dag.acceptingBlock(redChainBlock)
	if err != nil {
		t.Fatalf("TestAcceptingBlock: acceptingBlock for red chain block unexpectedly failed: %s", err)
	}
	if redChainBlockAcceptionBlock != nil {
		t.Fatalf("TestAcceptingBlock: unexpected acceptingBlock for red chain block. "+
			"Want: nil, got: %s", redChainBlockAcceptionBlock.hash)
	}

	// Make sure that a red tip returns nil
	redChainTip := dag.index.LookupNode(chainBlocks[len(chainBlocks)-1].Hash())
	redChainTipAcceptingBlock, err := dag.acceptingBlock(redChainTip)
	if err != nil {
		t.Fatalf("TestAcceptingBlock: acceptingBlock for red chain tip unexpectedly failed: %s", err)
	}
	if redChainTipAcceptingBlock != nil {
		t.Fatalf("TestAcceptingBlock: unexpected acceptingBlock for red tip block. "+
			"Want: nil, got: %s", redChainTipAcceptingBlock.hash)
	}
}

// payToPubKeyHashScript creates a new script to pay a transaction
// output to a 20-byte pubkey hash. It is expected that the input is a valid
// hash.
func payToPubKeyHashScript(pubKeyHash []byte) ([]byte, error) {
	return txscript.NewScriptBuilder().
		AddOp(txscript.OpDup).
		AddOp(txscript.OpHash160).
		AddData(pubKeyHash).
		AddOp(txscript.OpEqualVerify).
		AddOp(txscript.OpCheckSig).
		Script()
}
