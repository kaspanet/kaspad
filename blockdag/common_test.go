// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"compress/bzip2"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kaspanet/kaspad/util/mstime"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

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

		// Deserialize the UTXO entry and add it to the UTXO set.
		entry, err := deserializeUTXOEntry(r)
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
	dag.Params.BlockCoinbaseMaturity = maturity
}

// newTestDAG returns a DAG that is usable for syntetic tests. It is
// important to note that this DAG has no database associated with it, so
// it is not usable with all functions and the tests must take care when making
// use of it.
func newTestDAG(params *dagconfig.Params) *BlockDAG {
	index := newBlockIndex(params)
	dag := &BlockDAG{
		Params:                         params,
		timeSource:                     NewTimeSource(),
		difficultyAdjustmentWindowSize: params.DifficultyAdjustmentWindowSize,
		TimestampDeviationTolerance:    params.TimestampDeviationTolerance,
		powMaxBits:                     util.BigToCompact(params.PowMax),
		index:                          index,
		warningCaches:                  newThresholdCaches(vbNumBits),
		deploymentCaches:               newThresholdCaches(dagconfig.DefinedDeployments),
	}

	// Create a genesis block node and block index index populated with it
	// on the above fake DAG.
	dag.genesis, _ = dag.newBlockNode(&params.GenesisBlock.Header, newBlockSet())
	index.AddNode(dag.genesis)

	dag.virtual = newVirtualBlock(dag, blockSetFromSlice(dag.genesis))
	return dag
}

// newTestNode creates a block node connected to the passed parent with the
// provided fields populated and fake values for the other fields.
func newTestNode(dag *BlockDAG, parents blockSet, blockVersion int32, bits uint32, timestamp mstime.Time) *blockNode {
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
	node, _ := dag.newBlockNode(header, parents)
	return node
}

func addNodeAsChildToParents(node *blockNode) {
	for parent := range node.parents {
		parent.children.add(node)
	}
}

// checkRuleError ensures the type of the two passed errors are of the
// same type (either both nil or both of type RuleError) and their error codes
// match when not nil.
func checkRuleError(gotErr, wantErr error) error {
	if wantErr == nil && gotErr == nil {
		return nil
	}

	var gotRuleErr RuleError
	if ok := errors.As(gotErr, &gotRuleErr); !ok {
		return errors.Errorf("gotErr expected to be RuleError, but got %+v instead", gotErr)
	}

	var wantRuleErr RuleError
	if ok := errors.As(wantErr, &wantRuleErr); !ok {
		return errors.Errorf("wantErr expected to be RuleError, but got %+v instead", wantErr)
	}

	// Ensure the error codes match. It's safe to use a raw type assert
	// here since the code above already proved they are the same type and
	// the want error is a script error.
	if gotRuleErr.ErrorCode != wantRuleErr.ErrorCode {
		return errors.Errorf("mismatched error code - got %v (%v), want %v",
			gotRuleErr.ErrorCode, gotErr, wantRuleErr.ErrorCode)
	}

	return nil
}

func prepareAndProcessBlockByParentMsgBlocks(t *testing.T, dag *BlockDAG, parents ...*wire.MsgBlock) *wire.MsgBlock {
	parentHashes := make([]*daghash.Hash, len(parents))
	for i, parent := range parents {
		parentHashes[i] = parent.BlockHash()
	}
	return PrepareAndProcessBlockForTest(t, dag, parentHashes, nil)
}

func nodeByMsgBlock(t *testing.T, dag *BlockDAG, block *wire.MsgBlock) *blockNode {
	node, ok := dag.index.LookupNode(block.BlockHash())
	if !ok {
		t.Fatalf("couldn't find block node with hash %s", block.BlockHash())
	}
	return node
}

type fakeTimeSource struct {
	time mstime.Time
}

func (fts *fakeTimeSource) Now() mstime.Time {
	return fts.time
}

func newFakeTimeSource(fakeTime mstime.Time) TimeSource {
	return &fakeTimeSource{time: fakeTime}
}
