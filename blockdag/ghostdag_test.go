package blockdag

import (
	"fmt"
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/dbaccess"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"reflect"
	"sort"
	"strings"
	"testing"
)

type testBlockData struct {
	parents                []string
	id                     string // id is a virtual entity that is used only for tests so we can define relations between blocks without knowing their hash
	expectedScore          uint64
	expectedSelectedParent string
	expectedBlues          []string
}

// TestGHOSTDAG iterates over several dag simulations, and checks
// that the blue score, blue set and selected parent of each
// block are calculated as expected.
func TestGHOSTDAG(t *testing.T) {
	dagParams := dagconfig.SimnetParams

	tests := []struct {
		k            dagconfig.KType
		expectedReds []string
		dagData      []*testBlockData
	}{
		{
			k:            3,
			expectedReds: []string{"F", "G", "H", "I", "N", "Q"},
			dagData: []*testBlockData{
				{
					parents:                []string{"A"},
					id:                     "B",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"B"},
					id:                     "C",
					expectedScore:          2,
					expectedSelectedParent: "B",
					expectedBlues:          []string{"B"},
				},
				{
					parents:                []string{"A"},
					id:                     "D",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"C", "D"},
					id:                     "E",
					expectedScore:          4,
					expectedSelectedParent: "C",
					expectedBlues:          []string{"C", "D"},
				},
				{
					parents:                []string{"A"},
					id:                     "F",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"F"},
					id:                     "G",
					expectedScore:          2,
					expectedSelectedParent: "F",
					expectedBlues:          []string{"F"},
				},
				{
					parents:                []string{"A"},
					id:                     "H",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"A"},
					id:                     "I",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"E", "G"},
					id:                     "J",
					expectedScore:          5,
					expectedSelectedParent: "E",
					expectedBlues:          []string{"E"},
				},
				{
					parents:                []string{"J"},
					id:                     "K",
					expectedScore:          6,
					expectedSelectedParent: "J",
					expectedBlues:          []string{"J"},
				},
				{
					parents:                []string{"I", "K"},
					id:                     "L",
					expectedScore:          7,
					expectedSelectedParent: "K",
					expectedBlues:          []string{"K"},
				},
				{
					parents:                []string{"L"},
					id:                     "M",
					expectedScore:          8,
					expectedSelectedParent: "L",
					expectedBlues:          []string{"L"},
				},
				{
					parents:                []string{"M"},
					id:                     "N",
					expectedScore:          9,
					expectedSelectedParent: "M",
					expectedBlues:          []string{"M"},
				},
				{
					parents:                []string{"M"},
					id:                     "O",
					expectedScore:          9,
					expectedSelectedParent: "M",
					expectedBlues:          []string{"M"},
				},
				{
					parents:                []string{"M"},
					id:                     "P",
					expectedScore:          9,
					expectedSelectedParent: "M",
					expectedBlues:          []string{"M"},
				},
				{
					parents:                []string{"M"},
					id:                     "Q",
					expectedScore:          9,
					expectedSelectedParent: "M",
					expectedBlues:          []string{"M"},
				},
				{
					parents:                []string{"M"},
					id:                     "R",
					expectedScore:          9,
					expectedSelectedParent: "M",
					expectedBlues:          []string{"M"},
				},
				{
					parents:                []string{"R"},
					id:                     "S",
					expectedScore:          10,
					expectedSelectedParent: "R",
					expectedBlues:          []string{"R"},
				},
				{
					parents:                []string{"N", "O", "P", "Q", "S"},
					id:                     "T",
					expectedScore:          13,
					expectedSelectedParent: "S",
					expectedBlues:          []string{"S", "O", "P"},
				},
			},
		},
	}

	for i, test := range tests {
		func() {
			resetExtraNonceForTest()
			dagParams.K = test.k
			dag, teardownFunc, err := DAGSetup(fmt.Sprintf("TestGHOSTDAG%d", i), true, Config{
				DAGParams: &dagParams,
			})
			if err != nil {
				t.Fatalf("Failed to setup dag instance: %v", err)
			}
			defer teardownFunc()

			genesisNode := dag.genesis
			blockByIDMap := make(map[string]*blockNode)
			idByBlockMap := make(map[*blockNode]string)
			blockByIDMap["A"] = genesisNode
			idByBlockMap[genesisNode] = "A"

			for _, blockData := range test.dagData {
				parents := blockSet{}
				for _, parentID := range blockData.parents {
					parent := blockByIDMap[parentID]
					parents.add(parent)
				}

				block, err := PrepareBlockForTest(dag, parents.hashes(), nil)
				if err != nil {
					t.Fatalf("TestGHOSTDAG: block %v got unexpected error from PrepareBlockForTest: %v", blockData.id, err)
				}

				utilBlock := util.NewBlock(block)
				isOrphan, isDelayed, err := dag.ProcessBlock(utilBlock, BFNoPoWCheck)
				if err != nil {
					t.Fatalf("TestGHOSTDAG: dag.ProcessBlock got unexpected error for block %v: %v", blockData.id, err)
				}
				if isDelayed {
					t.Fatalf("TestGHOSTDAG: block %s "+
						"is too far in the future", blockData.id)
				}
				if isOrphan {
					t.Fatalf("TestGHOSTDAG: block %v was unexpectedly orphan", blockData.id)
				}

				node := dag.index.LookupNode(utilBlock.Hash())

				blockByIDMap[blockData.id] = node
				idByBlockMap[node] = blockData.id

				bluesIDs := make([]string, 0, len(node.blues))
				for _, blue := range node.blues {
					bluesIDs = append(bluesIDs, idByBlockMap[blue])
				}
				selectedParentID := idByBlockMap[node.selectedParent]
				fullDataStr := fmt.Sprintf("blues: %v, selectedParent: %v, score: %v",
					bluesIDs, selectedParentID, node.blueScore)
				if blockData.expectedScore != node.blueScore {
					t.Errorf("Test %d: Block %v expected to have score %v but got %v (fulldata: %v)",
						i, blockData.id, blockData.expectedScore, node.blueScore, fullDataStr)
				}
				if blockData.expectedSelectedParent != selectedParentID {
					t.Errorf("Test %d: Block %v expected to have selected parent %v but got %v (fulldata: %v)",
						i, blockData.id, blockData.expectedSelectedParent, selectedParentID, fullDataStr)
				}
				if !reflect.DeepEqual(blockData.expectedBlues, bluesIDs) {
					t.Errorf("Test %d: Block %v expected to have blues %v but got %v (fulldata: %v)",
						i, blockData.id, blockData.expectedBlues, bluesIDs, fullDataStr)
				}
			}

			reds := make(map[string]bool)

			for id := range blockByIDMap {
				reds[id] = true
			}

			for tip := &dag.virtual.blockNode; tip.selectedParent != nil; tip = tip.selectedParent {
				tipID := idByBlockMap[tip]
				delete(reds, tipID)
				for _, blue := range tip.blues {
					blueID := idByBlockMap[blue]
					delete(reds, blueID)
				}
			}
			if !checkReds(test.expectedReds, reds) {
				redsIDs := make([]string, 0, len(reds))
				for id := range reds {
					redsIDs = append(redsIDs, id)
				}
				sort.Strings(redsIDs)
				sort.Strings(test.expectedReds)
				t.Errorf("Test %d: Expected reds %v but got %v", i, test.expectedReds, redsIDs)
			}
		}()
	}
}

func checkReds(expectedReds []string, reds map[string]bool) bool {
	if len(expectedReds) != len(reds) {
		return false
	}
	for _, redID := range expectedReds {
		if !reds[redID] {
			return false
		}
	}
	return true
}

func TestBlueAnticoneSizeErrors(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	dag, teardownFunc, err := DAGSetup("TestBlueAnticoneSizeErrors", true, Config{
		DAGParams: &dagconfig.SimnetParams,
	})
	if err != nil {
		t.Fatalf("TestBlueAnticoneSizeErrors: Failed to setup DAG instance: %s", err)
	}
	defer teardownFunc()

	// Prepare a block chain with size K beginning with the genesis block
	currentBlockA := dag.dagParams.GenesisBlock
	for i := dagconfig.KType(0); i < dag.dagParams.K; i++ {
		newBlock := prepareAndProcessBlockByParentMsgBlocks(t, dag, currentBlockA)
		currentBlockA = newBlock
	}

	// Prepare another block chain with size K beginning with the genesis block
	currentBlockB := dag.dagParams.GenesisBlock
	for i := dagconfig.KType(0); i < dag.dagParams.K; i++ {
		newBlock := prepareAndProcessBlockByParentMsgBlocks(t, dag, currentBlockB)
		currentBlockB = newBlock
	}

	// Get references to the tips of the two chains
	blockNodeA := dag.index.LookupNode(currentBlockA.BlockHash())
	blockNodeB := dag.index.LookupNode(currentBlockB.BlockHash())

	// Try getting the blueAnticoneSize between them. Since the two
	// blocks are not in the anticones of eachother, this should fail.
	_, err = dag.blueAnticoneSize(blockNodeA, blockNodeB)
	if err == nil {
		t.Fatalf("TestBlueAnticoneSizeErrors: blueAnticoneSize unexpectedly succeeded")
	}
	expectedErrSubstring := "is not in blue set of"
	if !strings.Contains(err.Error(), expectedErrSubstring) {
		t.Fatalf("TestBlueAnticoneSizeErrors: blueAnticoneSize returned wrong error. "+
			"Want: %s, got: %s", expectedErrSubstring, err)
	}
}

func TestGHOSTDAGErrors(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	dag, teardownFunc, err := DAGSetup("TestGHOSTDAGErrors", true, Config{
		DAGParams: &dagconfig.SimnetParams,
	})
	if err != nil {
		t.Fatalf("TestGHOSTDAGErrors: Failed to setup DAG instance: %s", err)
	}
	defer teardownFunc()

	// Add two child blocks to the genesis
	block1 := prepareAndProcessBlockByParentMsgBlocks(t, dag, dag.dagParams.GenesisBlock)
	block2 := prepareAndProcessBlockByParentMsgBlocks(t, dag, dag.dagParams.GenesisBlock)

	// Add a child block to the previous two blocks
	block3 := prepareAndProcessBlockByParentMsgBlocks(t, dag, block1, block2)

	// Clear the reachability store
	dag.reachabilityStore.loaded = map[daghash.Hash]*reachabilityData{}

	dbTx, err := dbaccess.NewTx()
	if err != nil {
		t.Fatalf("NewTx: %s", err)
	}
	defer dbTx.RollbackUnlessClosed()

	err = dbaccess.ClearReachabilityData(dbTx)
	if err != nil {
		t.Fatalf("ClearReachabilityData: %s", err)
	}

	err = dbTx.Commit()
	if err != nil {
		t.Fatalf("Commit: %s", err)
	}

	// Try to rerun GHOSTDAG on the last block. GHOSTDAG uses
	// reachability data, so we expect it to fail.
	blockNode3 := dag.index.LookupNode(block3.BlockHash())
	_, err = dag.ghostdag(blockNode3)
	if err == nil {
		t.Fatalf("TestGHOSTDAGErrors: ghostdag unexpectedly succeeded")
	}
	expectedErrSubstring := "Couldn't find reachability data"
	if !strings.Contains(err.Error(), expectedErrSubstring) {
		t.Fatalf("TestGHOSTDAGErrors: ghostdag returned wrong error. "+
			"Want: %s, got: %s", expectedErrSubstring, err)
	}
}
