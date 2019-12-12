// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"compress/bzip2"
	"encoding/binary"
	"github.com/pkg/errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/kaspanet/kaspad/dagconfig"
	_ "github.com/kaspanet/kaspad/database/ffldb"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

func loadBlocksWithLog(t *testing.T, filename string) ([]*util.Block, error) {
	blocks, err := LoadBlocks(filename)
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

// TestSetCoinbaseMaturity makes the ability to set the coinbase maturity
// available when running tests.
func (dag *BlockDAG) TestSetCoinbaseMaturity(maturity uint64) {
	dag.dagParams.BlockCoinbaseMaturity = maturity
}

// newTestDAG returns a DAG that is usable for syntetic tests. It is
// important to note that this DAG has no database associated with it, so
// it is not usable with all functions and the tests must take care when making
// use of it.
func newTestDAG(params *dagconfig.Params) *BlockDAG {
	// Create a genesis block node and block index index populated with it
	// for use when creating the fake DAG below.
	node := newBlockNode(&params.GenesisBlock.Header, newSet(), params.K)
	index := newBlockIndex(nil, params)
	index.AddNode(node)

	targetTimePerBlock := int64(params.TargetTimePerBlock / time.Second)
	return &BlockDAG{
		dagParams:                      params,
		timeSource:                     NewMedianTime(),
		targetTimePerBlock:             targetTimePerBlock,
		difficultyAdjustmentWindowSize: params.DifficultyAdjustmentWindowSize,
		TimestampDeviationTolerance:    params.TimestampDeviationTolerance,
		powMaxBits:                     util.BigToCompact(params.PowMax),
		index:                          index,
		virtual:                        newVirtualBlock(setFromSlice(node), params.K),
		genesis:                        index.LookupNode(params.GenesisHash),
		warningCaches:                  newThresholdCaches(vbNumBits),
		deploymentCaches:               newThresholdCaches(dagconfig.DefinedDeployments),
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
		return errors.Errorf("wrong error - got %T (%[1]v), want %T",
			gotErr, wantErr)
	}
	if gotErr == nil {
		return nil
	}

	// Ensure the want error type is a script error.
	werr, ok := wantErr.(RuleError)
	if !ok {
		return errors.Errorf("unexpected test error type %T", wantErr)
	}

	// Ensure the error codes match. It's safe to use a raw type assert
	// here since the code above already proved they are the same type and
	// the want error is a script error.
	gotErrorCode := gotErr.(RuleError).ErrorCode
	if gotErrorCode != werr.ErrorCode {
		return errors.Errorf("mismatched error code - got %v (%v), want %v",
			gotErrorCode, gotErr, werr.ErrorCode)
	}

	return nil
}
