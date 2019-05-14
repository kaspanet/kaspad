package blockdag

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/wire"
)

func TestUTXODiffStore(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	dag, teardownFunc, err := DAGSetup("TestUTXODiffStore", Config{
		DAGParams: &dagconfig.SimNetParams,
	})
	if err != nil {
		t.Fatalf("TestUTXODiffStore: Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	nodeCounter := byte(0)
	createNode := func() *blockNode {
		nodeCounter++
		node := &blockNode{hash: &daghash.Hash{nodeCounter}}
		dag.index.AddNode(node)
		return node
	}

	// Check that an error is returned when asking for non existing node
	nonExistingNode := createNode()
	_, err = dag.utxoDiffStore.diffByNode(nonExistingNode)
	expectedErrString := fmt.Sprintf("Couldn't find diff data for block %s", nonExistingNode.hash)
	if err == nil || err.Error() != expectedErrString {
		t.Errorf("diffByNode: expected error %s but got %s", expectedErrString, err)
	}

	// Add node's diff data to the utxoDiffStore and check if it's checked correctly.
	node := createNode()
	diff := NewUTXODiff()
	diff.toAdd.add(wire.OutPoint{TxID: daghash.TxID{0x01}, Index: 0}, &UTXOEntry{amount: 1, pkScript: []byte{0x01}})
	diff.toRemove.add(wire.OutPoint{TxID: daghash.TxID{0x02}, Index: 0}, &UTXOEntry{amount: 2, pkScript: []byte{0x02}})
	if err := dag.utxoDiffStore.setBlockDiff(node, diff); err != nil {
		t.Fatalf("setBlockDiff: unexpected error: %s", err)
	}
	diffChild := createNode()
	if err := dag.utxoDiffStore.setBlockDiffChild(node, diffChild); err != nil {
		t.Fatalf("setBlockDiffChild: unexpected error: %s", err)
	}

	if storeDiff, err := dag.utxoDiffStore.diffByNode(node); err != nil {
		t.Fatalf("diffByNode: unexpected error: %s", err)
	} else if !reflect.DeepEqual(storeDiff, diff) {
		t.Errorf("Expected diff and storeDiff to be equal")
	}

	if storeDiffChild, err := dag.utxoDiffStore.diffChildByNode(node); err != nil {
		t.Fatalf("diffByNode: unexpected error: %s", err)
	} else if !reflect.DeepEqual(storeDiffChild, diffChild) {
		t.Errorf("Expected diff and storeDiff to be equal")
	}

	// Flush changes to db, delete them from the dag.utxoDiffStore.loaded
	// map, and check if the diff data is re-fetched from the database.
	err = dag.db.Update(func(dbTx database.Tx) error {
		return dag.utxoDiffStore.flushToDB(dbTx)
	})
	if err != nil {
		t.Fatalf("Error flushing utxoDiffStore data to DB: %s", err)
	}
	delete(dag.utxoDiffStore.loaded, *node.hash)

	if storeDiff, err := dag.utxoDiffStore.diffByNode(node); err != nil {
		t.Fatalf("diffByNode: unexpected error: %s", err)
	} else if !reflect.DeepEqual(storeDiff, diff) {
		t.Errorf("Expected diff and storeDiff to be equal")
	}

	// Check if getBlockDiff caches the result in dag.utxoDiffStore.loaded
	if loadedDiffData, ok := dag.utxoDiffStore.loaded[*node.hash]; !ok {
		t.Errorf("the diff data wasn't added to loaded map after requesting it")
	} else if !reflect.DeepEqual(loadedDiffData.diff, diff) {
		t.Errorf("Expected diff and loadedDiff to be equal")
	}
}
