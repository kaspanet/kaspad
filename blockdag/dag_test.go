// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"reflect"
	"testing"
	"time"

	"math/rand"

	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/wire"
)

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
			t.Errorf("Error loading file: %v\n", err)
			return
		}
		blocks = append(blocks, blockTmp...)
	}

	// Create a new database and chain instance to run tests against.
	chain, teardownFunc, err := chainSetup("haveblock",
		&dagconfig.MainNetParams)
	if err != nil {
		t.Errorf("Failed to setup chain instance: %v", err)
		return
	}
	defer teardownFunc()

	// Since we're not dealing with the real block chain, set the coinbase
	// maturity to 1.
	chain.TstSetCoinbaseMaturity(1)

	for i := 1; i < len(blocks); i++ {
		isOrphan, err := chain.ProcessBlock(blocks[i], BFNone)
		if err != nil {
			t.Errorf("ProcessBlock fail on block %v: %v\n", i, err)
			return
		}
		if isOrphan {
			t.Errorf("ProcessBlock incorrectly returned block %v "+
				"is an orphan\n", i)
			return
		}
	}

	testFiles = []string{
		"blk_3C.dat",
	}

	for _, file := range testFiles {
		blockTmp, err := loadBlocks(file)
		if err != nil {
			t.Errorf("Error loading file: %v\n", err)
			return
		}
		blocks = append(blocks, blockTmp...)
	}
	isOrphan, err := chain.ProcessBlock(blocks[6], BFNone)

	// Block 3c should fail to connect since its parents are related. (It points to A and B, and A is the parent of B)
	if err == nil {
		t.Errorf("ProcessBlock for block 3c has no error when expected to have an error\n")
		return
	}
	if isOrphan {
		t.Errorf("ProcessBlock incorrectly returned block 3c " +
			"is an orphan\n")
		return
	}

	// Insert an orphan block.
	isOrphan, err = chain.ProcessBlock(util.NewBlock(&Block100000),
		BFNone)
	if err != nil {
		t.Errorf("Unable to process block: %v", err)
		return
	}
	if !isOrphan {
		t.Errorf("ProcessBlock indicated block is an not orphan when " +
			"it should be\n")
		return
	}

	tests := []struct {
		hash string
		want bool
	}{
		// Genesis block should be present.
		{hash: dagconfig.MainNetParams.GenesisHash.String(), want: true},

		// Block 3b should be present (as a second child of Block 2).
		{hash: "00000093c8f2ab3444502da0754fc8149d738701aef9b2e0f32f32c078039295", want: true},

		// Block 100000 should be present (as an orphan).
		{hash: "000000a805b083e0ef1f516b1153828724c235d6e6f0fabb47b869f6d054ac3f", want: true},

		// Random hashes should not be available.
		{hash: "123", want: false},
	}

	for i, test := range tests {
		hash, err := daghash.NewHashFromStr(test.hash)
		if err != nil {
			t.Errorf("NewHashFromStr: %v", err)
			continue
		}

		result, err := chain.HaveBlock(hash)
		if err != nil {
			t.Errorf("HaveBlock #%d unexpected error: %v", i, err)
			return
		}
		if result != test.want {
			t.Errorf("HaveBlock #%d got %v want %v", i, result,
				test.want)
			continue
		}
	}
}

func TestProcessBlock(t *testing.T) {
	// Block 3c should fail to connect since its parents are related. (It points to A and B, and A is the parent of B)
}

// TestCalcSequenceLock tests the LockTimeToSequence function, and the
// CalcSequenceLock method of a Chain instance. The tests exercise several
// combinations of inputs to the CalcSequenceLock function in order to ensure
// the returned SequenceLocks are correct for each test instance.
func TestCalcSequenceLock(t *testing.T) {
	netParams := &dagconfig.SimNetParams

	blockVersion := int32(0x10000000)

	// Generate enough synthetic blocks for the rest of the test
	chain := newTestDAG(netParams)
	node := chain.virtual.SelectedTip()
	blockTime := node.Header().Timestamp
	numBlocksToGenerate := uint32(5)
	for i := uint32(0); i < numBlocksToGenerate; i++ {
		blockTime = blockTime.Add(time.Second)
		node = newTestNode(setFromSlice(node), blockVersion, 0, blockTime, netParams.K)
		chain.index.AddNode(node)
		chain.virtual.SetTips(setFromSlice(node))
	}

	// Create a utxo view with a fake utxo for the inputs used in the
	// transactions created below.  This utxo is added such that it has an
	// age of 4 blocks.
	targetTx := util.NewTx(&wire.MsgTx{
		TxOut: []*wire.TxOut{{
			PkScript: nil,
			Value:    10,
		}},
	})
	utxoView := NewUtxoViewpoint()
	utxoView.AddTxOuts(targetTx, int32(numBlocksToGenerate)-4)
	utxoView.SetTips(setFromSlice(node))

	// Create a utxo that spends the fake utxo created above for use in the
	// transactions created in the tests.  It has an age of 4 blocks.  Note
	// that the sequence lock heights are always calculated from the same
	// point of view that they were originally calculated from for a given
	// utxo.  That is to say, the height prior to it.
	utxo := wire.OutPoint{
		Hash:  *targetTx.Hash(),
		Index: 0,
	}
	prevUtxoHeight := int32(numBlocksToGenerate) - 4

	// Obtain the median time past from the PoV of the input created above.
	// The MTP for the input is the MTP from the PoV of the block *prior*
	// to the one that included it.
	medianTime := node.RelativeAncestor(5).CalcPastMedianTime().Unix()

	// The median time calculated from the PoV of the best block in the
	// test chain.  For unconfirmed inputs, this value will be used since
	// the MTP will be calculated from the PoV of the yet-to-be-mined
	// block.
	nextMedianTime := node.CalcPastMedianTime().Unix()
	nextBlockHeight := int32(numBlocksToGenerate) + 1

	// Add an additional transaction which will serve as our unconfirmed
	// output.
	unConfTx := &wire.MsgTx{
		TxOut: []*wire.TxOut{{
			PkScript: nil,
			Value:    5,
		}},
	}
	unConfUtxo := wire.OutPoint{
		Hash:  unConfTx.TxHash(),
		Index: 0,
	}

	// Adding a utxo with a height of 0x7fffffff indicates that the output
	// is currently unmined.
	utxoView.AddTxOuts(util.NewTx(unConfTx), 0x7fffffff)

	tests := []struct {
		tx      *wire.MsgTx
		view    *UtxoViewpoint
		mempool bool
		want    *SequenceLock
	}{
		// A transaction with a single input with max sequence number.
		// This sequence number has the high bit set, so sequence locks
		// should be disabled.
		{
			tx: &wire.MsgTx{
				Version: 1,
				TxIn: []*wire.TxIn{{
					PreviousOutPoint: utxo,
					Sequence:         wire.MaxTxInSequenceNum,
				}},
			},
			view: utxoView,
			want: &SequenceLock{
				Seconds:     -1,
				BlockHeight: -1,
			},
		},
		// A transaction with a single input whose lock time is
		// expressed in seconds.  However, the specified lock time is
		// below the required floor for time based lock times since
		// they have time granularity of 512 seconds.  As a result, the
		// seconds lock-time should be just before the median time of
		// the targeted block.
		{
			tx: &wire.MsgTx{
				Version: 1,
				TxIn: []*wire.TxIn{{
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(true, 2),
				}},
			},
			view: utxoView,
			want: &SequenceLock{
				Seconds:     medianTime - 1,
				BlockHeight: -1,
			},
		},
		// A transaction with a single input whose lock time is
		// expressed in seconds.  The number of seconds should be 1023
		// seconds after the median past time of the last block in the
		// chain.
		{
			tx: &wire.MsgTx{
				Version: 1,
				TxIn: []*wire.TxIn{{
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(true, 1024),
				}},
			},
			view: utxoView,
			want: &SequenceLock{
				Seconds:     medianTime + 1023,
				BlockHeight: -1,
			},
		},
		// A transaction with multiple inputs.  The first input has a
		// lock time expressed in seconds.  The second input has a
		// sequence lock in blocks with a value of 4.  The last input
		// has a sequence number with a value of 5, but has the disable
		// bit set.  So the first lock should be selected as it's the
		// latest lock that isn't disabled.
		{
			tx: &wire.MsgTx{
				Version: 1,
				TxIn: []*wire.TxIn{{
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
			},
			view: utxoView,
			want: &SequenceLock{
				Seconds:     medianTime + (5 << wire.SequenceLockTimeGranularity) - 1,
				BlockHeight: prevUtxoHeight + 3,
			},
		},
		// Transaction with a single input.  The input's sequence number
		// encodes a relative lock-time in blocks (3 blocks).  The
		// sequence lock should  have a value of -1 for seconds, but a
		// height of 2 meaning it can be included at height 3.
		{
			tx: &wire.MsgTx{
				Version: 1,
				TxIn: []*wire.TxIn{{
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(false, 3),
				}},
			},
			view: utxoView,
			want: &SequenceLock{
				Seconds:     -1,
				BlockHeight: prevUtxoHeight + 2,
			},
		},
		// A transaction with two inputs with lock times expressed in
		// seconds.  The selected sequence lock value for seconds should
		// be the time further in the future.
		{
			tx: &wire.MsgTx{
				Version: 1,
				TxIn: []*wire.TxIn{{
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(true, 5120),
				}, {
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(true, 2560),
				}},
			},
			view: utxoView,
			want: &SequenceLock{
				Seconds:     medianTime + (10 << wire.SequenceLockTimeGranularity) - 1,
				BlockHeight: -1,
			},
		},
		// A transaction with two inputs with lock times expressed in
		// blocks.  The selected sequence lock value for blocks should
		// be the height further in the future, so a height of 10
		// indicating it can be included at height 11.
		{
			tx: &wire.MsgTx{
				Version: 1,
				TxIn: []*wire.TxIn{{
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(false, 1),
				}, {
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(false, 11),
				}},
			},
			view: utxoView,
			want: &SequenceLock{
				Seconds:     -1,
				BlockHeight: prevUtxoHeight + 10,
			},
		},
		// A transaction with multiple inputs.  Two inputs are time
		// based, and the other two are block based. The lock lying
		// further into the future for both inputs should be chosen.
		{
			tx: &wire.MsgTx{
				Version: 1,
				TxIn: []*wire.TxIn{{
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
			},
			view: utxoView,
			want: &SequenceLock{
				Seconds:     medianTime + (13 << wire.SequenceLockTimeGranularity) - 1,
				BlockHeight: prevUtxoHeight + 8,
			},
		},
		// A transaction with a single unconfirmed input.  As the input
		// is confirmed, the height of the input should be interpreted
		// as the height of the *next* block.  So, a 2 block relative
		// lock means the sequence lock should be for 1 block after the
		// *next* block height, indicating it can be included 2 blocks
		// after that.
		{
			tx: &wire.MsgTx{
				Version: 1,
				TxIn: []*wire.TxIn{{
					PreviousOutPoint: unConfUtxo,
					Sequence:         LockTimeToSequence(false, 2),
				}},
			},
			view:    utxoView,
			mempool: true,
			want: &SequenceLock{
				Seconds:     -1,
				BlockHeight: nextBlockHeight + 1,
			},
		},
		// A transaction with a single unconfirmed input.  The input has
		// a time based lock, so the lock time should be based off the
		// MTP of the *next* block.
		{
			tx: &wire.MsgTx{
				Version: 1,
				TxIn: []*wire.TxIn{{
					PreviousOutPoint: unConfUtxo,
					Sequence:         LockTimeToSequence(true, 1024),
				}},
			},
			view:    utxoView,
			mempool: true,
			want: &SequenceLock{
				Seconds:     nextMedianTime + 1023,
				BlockHeight: -1,
			},
		},
	}

	t.Logf("Running %v SequenceLock tests", len(tests))
	for i, test := range tests {
		utilTx := util.NewTx(test.tx)
		seqLock, err := chain.CalcSequenceLock(utilTx, test.view, test.mempool)
		if err != nil {
			t.Fatalf("test #%d, unable to calc sequence lock: %v", i, err)
		}

		if seqLock.Seconds != test.want.Seconds {
			t.Fatalf("test #%d got %v seconds want %v seconds",
				i, seqLock.Seconds, test.want.Seconds)
		}
		if seqLock.BlockHeight != test.want.BlockHeight {
			t.Fatalf("test #%d got height of %v want height of %v ",
				i, seqLock.BlockHeight, test.want.BlockHeight)
		}
	}
}

// nodeHashes is a convenience function that returns the hashes for all of the
// passed indexes of the provided nodes.  It is used to construct expected hash
// slices in the tests.
func nodeHashes(nodes []*blockNode, indexes ...int) []daghash.Hash {
	hashes := make([]daghash.Hash, 0, len(indexes))
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
		header := wire.BlockHeader{Nonce: testNoncePrng.Uint32()}
		header.PrevBlocks = tips.hashes()
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
	blockDAG := newTestDAG(&dagconfig.MainNetParams)
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
		startHeight int32          // locator for requested inventory
		endHash     daghash.Hash   // stop hash for locator
		maxResults  int            // max to locate, 0 = wire const
		hashes      []daghash.Hash // expected located hashes
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
		hashes, err := blockDAG.HeightToHashRange(test.startHeight, &test.endHash,
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
	chain := newTestDAG(&dagconfig.MainNetParams)
	branch0Nodes := chainedNodes(setFromSlice(chain.genesis), 18)
	branch1Nodes := chainedNodes(setFromSlice(branch0Nodes[14]), 3)
	for _, node := range branch0Nodes {
		chain.index.SetStatusFlags(node, statusValid)
		chain.index.AddNode(node)
	}
	for _, node := range branch1Nodes {
		if node.height < 18 {
			chain.index.SetStatusFlags(node, statusValid)
		}
		chain.index.AddNode(node)
	}
	chain.virtual.SetTips(setFromSlice(tip(branch0Nodes)))

	tests := []struct {
		name        string
		endHash     daghash.Hash
		interval    int
		hashes      []daghash.Hash
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
			hashes:   []daghash.Hash{},
		},
		{
			name:        "unvalidated block",
			endHash:     branch1Nodes[2].hash,
			interval:    8,
			expectError: true,
		},
	}
	for _, test := range tests {
		hashes, err := chain.IntervalBlockHashes(&test.endHash, test.interval)
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
