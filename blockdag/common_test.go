// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"compress/bzip2"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/daglabs/btcd/dagconfig"
	_ "github.com/daglabs/btcd/database/ffldb"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/wire"
)

// loadBlocks reads files containing bitcoin block data (gzipped but otherwise
// in the format bitcoind writes) from disk and returns them as an array of
// util.Block.  This is largely borrowed from the test code in btcdb.
func loadBlocks(filename string) (blocks []*util.Block, err error) {
	filename = filepath.Join("testdata/", filename)

	var network = wire.MainNet
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
		block.SetHeight(height)
		blocks = append(blocks, block)
	}

	return
}

func loadBlocksWithLog(t *testing.T, filename string) ([]*util.Block, error) {
	blocks, err := loadBlocks(filename)
	if err == nil {
		t.Logf("Loaded %d blocks from file %s", len(blocks), filename)
		for i, b := range blocks {
			t.Logf("Block #%d: %s", i, b.Hash())
		}
	}
	return blocks, err
}

// loadUTXOSet returns a utxo view loaded from a file.
func loadUTXOSet(filename string) (UTXOSet, error) {
	// The utxostore file format is:
	// <tx hash><output index><serialized utxo len><serialized utxo>
	//
	// The output index and serialized utxo len are little endian uint32s
	// and the serialized utxo uses the format described in dagio.go.

	filename = filepath.Join("testdata", filename)
	fi, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	// Choose read based on whether the file is compressed or not.
	var r io.Reader
	if strings.HasSuffix(filename, ".bz2") {
		r = bzip2.NewReader(fi)
	} else {
		r = fi
	}
	defer fi.Close()

	utxoSet := NewFullUTXOSet()
	for {
		// Tx ID of the utxo entry.
		var txID daghash.TxID
		_, err := io.ReadAtLeast(r, txID[:], len(txID[:]))
		if err != nil {
			// Expected EOF at the right offset.
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// Output index of the utxo entry.
		var index uint32
		err = binary.Read(r, binary.LittleEndian, &index)
		if err != nil {
			return nil, err
		}

		// Num of serialized utxo entry bytes.
		var numBytes uint32
		err = binary.Read(r, binary.LittleEndian, &numBytes)
		if err != nil {
			return nil, err
		}

		// Serialized utxo entry.
		serialized := make([]byte, numBytes)
		_, err = io.ReadAtLeast(r, serialized, int(numBytes))
		if err != nil {
			return nil, err
		}

		// Deserialize it and add it to the view.
		entry, err := deserializeUTXOEntry(serialized)
		if err != nil {
			return nil, err
		}
		utxoSet.utxoCollection[wire.Outpoint{TxID: txID, Index: index}] = entry
	}

	return utxoSet, nil
}

// TestSetBlockRewardMaturity makes the ability to set the block reward maturity
// available when running tests.
func (dag *BlockDAG) TestSetBlockRewardMaturity(maturity uint64) {
	dag.dagParams.BlockRewardMaturity = maturity
}

// newTestDAG returns a DAG that is usable for syntetic tests.  It is
// important to note that this chain has no database associated with it, so
// it is not usable with all functions and the tests must take care when making
// use of it.
func newTestDAG(params *dagconfig.Params) *BlockDAG {
	// Create a genesis block node and block index index populated with it
	// for use when creating the fake chain below.
	node := newBlockNode(&params.GenesisBlock.Header, newSet(), params.K)
	index := newBlockIndex(nil, params)
	index.AddNode(node)

	targetTimespan := int64(params.TargetTimespan / time.Second)
	targetTimePerBlock := int64(params.TargetTimePerBlock / time.Second)
	adjustmentFactor := params.RetargetAdjustmentFactor
	return &BlockDAG{
		dagParams:           params,
		timeSource:          NewMedianTime(),
		minRetargetTimespan: targetTimespan / adjustmentFactor,
		maxRetargetTimespan: targetTimespan * adjustmentFactor,
		blocksPerRetarget:   uint64(targetTimespan / targetTimePerBlock),
		index:               index,
		virtual:             newVirtualBlock(setFromSlice(node), params.K),
		genesis:             index.LookupNode(params.GenesisHash),
		warningCaches:       newThresholdCaches(vbNumBits),
		deploymentCaches:    newThresholdCaches(dagconfig.DefinedDeployments),
	}
}

// newTestNode creates a block node connected to the passed parent with the
// provided fields populated and fake values for the other fields.
func newTestNode(parents blockSet, blockVersion int32, bits uint32, timestamp time.Time, phantomK uint32) *blockNode {
	// Make up a header and create a block node from it.
	header := &wire.BlockHeader{
		Version:              blockVersion,
		ParentHashes:         parents.hashes(),
		Bits:                 bits,
		Timestamp:            timestamp,
		HashMerkleRoot:       &daghash.ZeroHash,
		AcceptedIDMerkleRoot: &daghash.ZeroHash,
		UTXOCommitment:       &daghash.ZeroHash,
	}
	return newBlockNode(header, parents, phantomK)
}

func addNodeAsChildToParents(node *blockNode) {
	for _, parent := range node.parents {
		parent.children.add(node)
	}
}

func buildNodeGenerator(phantomK uint32, withChildren bool) func(parents blockSet) *blockNode {
	// For the purposes of these tests, we'll create blockNodes whose hashes are a
	// series of numbers from 1 to 255.
	hashCounter := byte(1)
	buildNode := func(parents blockSet) *blockNode {
		block := newBlockNode(nil, parents, phantomK)
		block.hash = &daghash.Hash{hashCounter}
		hashCounter++

		return block
	}
	if withChildren {
		return func(parents blockSet) *blockNode {
			node := buildNode(parents)
			addNodeAsChildToParents(node)
			return node
		}
	}
	return buildNode
}

// checkRuleError ensures the type of the two passed errors are of the
// same type (either both nil or both of type RuleError) and their error codes
// match when not nil.
func checkRuleError(gotErr, wantErr error) error {
	// Ensure the error code is of the expected type and the error
	// code matches the value specified in the test instance.
	if reflect.TypeOf(gotErr) != reflect.TypeOf(wantErr) {
		return fmt.Errorf("wrong error - got %T (%[1]v), want %T",
			gotErr, wantErr)
	}
	if gotErr == nil {
		return nil
	}

	// Ensure the want error type is a script error.
	werr, ok := wantErr.(RuleError)
	if !ok {
		return fmt.Errorf("unexpected test error type %T", wantErr)
	}

	// Ensure the error codes match.  It's safe to use a raw type assert
	// here since the code above already proved they are the same type and
	// the want error is a script error.
	gotErrorCode := gotErr.(RuleError).ErrorCode
	if gotErrorCode != werr.ErrorCode {
		return fmt.Errorf("mismatched error code - got %v (%v), want %v",
			gotErrorCode, gotErr, werr.ErrorCode)
	}

	return nil
}
