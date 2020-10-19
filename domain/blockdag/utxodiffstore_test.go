package blockdag

import (
	"github.com/kaspanet/kaspad/domain/blocknode"
	"github.com/kaspanet/kaspad/domain/utxo"
	"reflect"
	"testing"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
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
	createNode := func() *blocknode.Node {
		nodeCounter++
		node := &blocknode.Node{Hash: &daghash.Hash{nodeCounter}}
		dag.Index.AddNode(node)
		return node
	}

	// Check that an error is returned when asking for non existing node
	nonExistingNode := createNode()
	_, err = dag.UTXODiffStore.DiffByNode(nonExistingNode)
	if !dbaccess.IsNotFoundError(err) {
		if err != nil {
			t.Errorf("DiffByNode: %s", err)
		} else {
			t.Errorf("DiffByNode: unexpectedly found diff data")
		}
	}

	// Add node's diff data to the utxoDiffStore and check if it's checked correctly.
	node := createNode()
	diff := utxo.NewDiff()
	txOut1 := &appmessage.TxOut{Value: 1, ScriptPubKey: []byte{0x01}}
	txOut2 := &appmessage.TxOut{Value: 2, ScriptPubKey: []byte{0x02}}
	diff.ToAdd.Add(appmessage.Outpoint{TxID: daghash.TxID{0x01}, Index: 0}, utxo.NewEntry(txOut1, false, 0))
	diff.ToRemove.Add(appmessage.Outpoint{TxID: daghash.TxID{0x02}, Index: 0}, utxo.NewEntry(txOut2, false, 0))
	if err := dag.UTXODiffStore.SetBlockDiff(node, diff); err != nil {
		t.Fatalf("SetBlockDiff: unexpected error: %s", err)
	}
	diffChild := createNode()
	if err := dag.UTXODiffStore.SetBlockDiffChild(node, diffChild); err != nil {
		t.Fatalf("SetBlockDiffChild: unexpected error: %s", err)
	}

	if storeDiff, err := dag.UTXODiffStore.DiffByNode(node); err != nil {
		t.Fatalf("DiffByNode: unexpected error: %s", err)
	} else if !reflect.DeepEqual(storeDiff, diff) {
		t.Errorf("Expected diff and storeDiff to be equal")
	}

	if storeDiffChild, err := dag.UTXODiffStore.DiffChildByNode(node); err != nil {
		t.Fatalf("DiffByNode: unexpected error: %s", err)
	} else if !reflect.DeepEqual(storeDiffChild, diffChild) {
		t.Errorf("Expected diff and storeDiff to be equal")
	}

	// Flush changes to db, delete them from the dag.utxoDiffStore.loaded
	// map, and check if the diff data is re-fetched from the database.
	dbTx, err := dag.DatabaseContext.NewTx()
	if err != nil {
		t.Fatalf("Failed to open database transaction: %s", err)
	}
	defer dbTx.RollbackUnlessClosed()
	err = dag.UTXODiffStore.FlushToDB(dbTx)
	if err != nil {
		t.Fatalf("Error flushing utxoDiffStore data to DB: %s", err)
	}
	err = dbTx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit database transaction: %s", err)
	}
	delete(dag.UTXODiffStore.Loaded, node)

	if storeDiff, err := dag.UTXODiffStore.DiffByNode(node); err != nil {
		t.Fatalf("DiffByNode: unexpected error: %s", err)
	} else if !reflect.DeepEqual(storeDiff, diff) {
		t.Errorf("Expected diff and storeDiff to be equal")
	}

	// Check if getBlockDiff caches the result in dag.utxoDiffStore.loaded
	if loadedDiffData, ok := dag.UTXODiffStore.Loaded[node]; !ok {
		t.Errorf("the diff data wasn't added to loaded map after requesting it")
	} else if !reflect.DeepEqual(loadedDiffData.Diff, diff) {
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
	currentDifference := utxo.MaxBlueScoreDifferenceToKeepLoaded
	utxo.MaxBlueScoreDifferenceToKeepLoaded = 10
	defer func() { utxo.MaxBlueScoreDifferenceToKeepLoaded = currentDifference }()

	// Add 10 blocks
	blockNodes := make([]*blocknode.Node, 10)
	for i := 0; i < 10; i++ {
		processedBlock := PrepareAndProcessBlockForTest(t, dag, dag.VirtualParentHashes(), nil)

		node, ok := dag.Index.LookupNode(processedBlock.BlockHash())
		if !ok {
			t.Fatalf("TestClearOldEntries: missing blockNode for hash %s", processedBlock.BlockHash())
		}
		blockNodes[i] = node
	}

	// Make sure that all of them exist in the loaded set
	for _, node := range blockNodes {
		_, ok := dag.UTXODiffStore.Loaded[node]
		if !ok {
			t.Fatalf("TestClearOldEntries: diffData for node %s is not in the loaded set", node.Hash)
		}
	}

	// Add 10 more blocks on top of the others
	for i := 0; i < 10; i++ {
		PrepareAndProcessBlockForTest(t, dag, dag.VirtualParentHashes(), nil)
	}

	// Make sure that all the old nodes no longer exist in the loaded set
	for _, node := range blockNodes {
		_, ok := dag.UTXODiffStore.Loaded[node]
		if ok {
			t.Fatalf("TestClearOldEntries: diffData for node %s is in the loaded set", node.Hash)
		}
	}
}
