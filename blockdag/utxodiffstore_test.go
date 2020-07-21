package blockdag

import (
	"reflect"
	"testing"

	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

func TestUTXODiffStore(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	dag, teardownFunc, err := DAGSetup("TestUTXODiffStore", true, Config{
		DAGParams: &dagconfig.SimnetParams,
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
	if !dbaccess.IsNotFoundError(err) {
		if err != nil {
			t.Errorf("diffByNode: %s", err)
		} else {
			t.Errorf("diffByNode: unexpectedly found diff data")
		}
	}

	// Add node's diff data to the utxoDiffStore and check if it's checked correctly.
	node := createNode()
	diff := NewUTXODiff()
	diff.toAdd.add(wire.Outpoint{TxID: daghash.TxID{0x01}, Index: 0}, &UTXOEntry{amount: 1, scriptPubKey: []byte{0x01}})
	diff.toRemove.add(wire.Outpoint{TxID: daghash.TxID{0x02}, Index: 0}, &UTXOEntry{amount: 2, scriptPubKey: []byte{0x02}})
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
	dbTx, err := dag.databaseContext.NewTx()
	if err != nil {
		t.Fatalf("Failed to open database transaction: %s", err)
	}
	defer dbTx.RollbackUnlessClosed()
	err = dag.utxoDiffStore.flushToDB(dbTx)
	if err != nil {
		t.Fatalf("Error flushing utxoDiffStore data to DB: %s", err)
	}
	err = dbTx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit database transaction: %s", err)
	}
	delete(dag.utxoDiffStore.loaded, node)

	if storeDiff, err := dag.utxoDiffStore.diffByNode(node); err != nil {
		t.Fatalf("diffByNode: unexpected error: %s", err)
	} else if !reflect.DeepEqual(storeDiff, diff) {
		t.Errorf("Expected diff and storeDiff to be equal")
	}

	// Check if getBlockDiff caches the result in dag.utxoDiffStore.loaded
	if loadedDiffData, ok := dag.utxoDiffStore.loaded[node]; !ok {
		t.Errorf("the diff data wasn't added to loaded map after requesting it")
	} else if !reflect.DeepEqual(loadedDiffData.diff, diff) {
		t.Errorf("Expected diff and loadedDiff to be equal")
	}
}

func TestClearOldEntries(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	dag, teardownFunc, err := DAGSetup("TestClearOldEntries", true, Config{
		DAGParams: &dagconfig.SimnetParams,
	})
	if err != nil {
		t.Fatalf("TestClearOldEntries: Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	// Set maxBlueScoreDifferenceToKeepLoaded to 10 to make this test fast to run
	currentDifference := maxBlueScoreDifferenceToKeepLoaded
	maxBlueScoreDifferenceToKeepLoaded = 10
	defer func() { maxBlueScoreDifferenceToKeepLoaded = currentDifference }()

	// Add 10 blocks
	blockNodes := make([]*blockNode, 10)
	for i := 0; i < 10; i++ {
		processedBlock := PrepareAndProcessBlockForTest(t, dag, dag.TipHashes(), nil)

		node, ok := dag.index.LookupNode(processedBlock.BlockHash())
		if !ok {
			t.Fatalf("TestClearOldEntries: missing blockNode for hash %s", processedBlock.BlockHash())
		}
		blockNodes[i] = node
	}

	// Make sure that all of them exist in the loaded set
	for _, node := range blockNodes {
		_, ok := dag.utxoDiffStore.loaded[node]
		if !ok {
			t.Fatalf("TestClearOldEntries: diffData for node %s is not in the loaded set", node.hash)
		}
	}

	// Add 10 more blocks on top of the others
	for i := 0; i < 10; i++ {
		PrepareAndProcessBlockForTest(t, dag, dag.TipHashes(), nil)
	}

	// Make sure that all the old nodes no longer exist in the loaded set
	for _, node := range blockNodes {
		_, ok := dag.utxoDiffStore.loaded[node]
		if ok {
			t.Fatalf("TestClearOldEntries: diffData for node %s is in the loaded set", node.hash)
		}
	}

	// Add a block on top of the genesis to force the retrieval of all diffData
	processedBlock := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{dag.genesis.hash}, nil)
	node, ok := dag.index.LookupNode(processedBlock.BlockHash())
	if !ok {
		t.Fatalf("TestClearOldEntries: missing blockNode for hash %s", processedBlock.BlockHash())
	}

	// Make sure that the child-of-genesis node is in the loaded set, since it
	// is a tip.
	_, ok = dag.utxoDiffStore.loaded[node]
	if !ok {
		t.Fatalf("TestClearOldEntries: diffData for node %s is not in the loaded set", node.hash)
	}

	// Make sure that all the old nodes still do not exist in the loaded set
	for _, node := range blockNodes {
		_, ok := dag.utxoDiffStore.loaded[node]
		if ok {
			t.Fatalf("TestClearOldEntries: diffData for node %s is in the loaded set", node.hash)
		}
	}
}
