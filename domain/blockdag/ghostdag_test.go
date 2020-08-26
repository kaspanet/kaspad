package blockdag

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
)

type block struct {
	ID                     string // id is a virtual entity that is used only for tests so we can define relations between blocks without knowing their hash
	ExpectedScore          uint64
	ExpectedSelectedParent string
	ExpectedBlues          []string
	Parents                []string
}

type testData struct {
	K            dagconfig.KType
	GenesisID    string
	ExpectedReds []string
	Blocks       []block
}

// TestGHOSTDAG iterates over several dag simulations, and checks
// that the blue score, blue set and selected parent of each
// block are calculated as expected.
func TestGHOSTDAG(t *testing.T) {
	dagParams := dagconfig.SimnetParams
	err := filepath.Walk("./testdata/dags/", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		var test testData
		file, err := os.Open(path)
		if err != nil {
			t.Fatalf("TestGHOSTDAG: failed opening file: %s", path)
		}
		decoder := json.NewDecoder(file)
		decoder.DisallowUnknownFields()
		err = decoder.Decode(&test)
		if err != nil {
			t.Fatalf("TestGHOSTDAG: test: %s, failed decoding json: %v", info.Name(), err)
		}

		func() {
			resetExtraNonceForTest()
			dagParams.K = test.K
			dag, teardownFunc, err := DAGSetup(fmt.Sprintf("TestGHOSTDAG %s", info.Name()), true, Config{
				DAGParams: &dagParams,
			})
			if err != nil {
				t.Fatalf("Failed to setup dag instance: %v", err)
			}
			defer teardownFunc()

			genesisNode := dag.genesis
			blockByIDMap := make(map[string]*blockNode)
			idByBlockMap := make(map[*blockNode]string)
			blockByIDMap[test.GenesisID] = genesisNode
			idByBlockMap[genesisNode] = test.GenesisID

			for _, blockData := range test.Blocks {
				parents := blockSet{}
				for _, parentID := range blockData.Parents {
					parent := blockByIDMap[parentID]
					parents.add(parent)
				}

				block, err := PrepareBlockForTest(dag, parents.hashes(), nil)
				if err != nil {
					t.Fatalf("TestGHOSTDAG: block %s got unexpected error from PrepareBlockForTest: %v", blockData.ID,
						err)
				}

				utilBlock := util.NewBlock(block)
				isOrphan, isDelayed, err := dag.ProcessBlock(utilBlock, BFNoPoWCheck)
				if err != nil {
					t.Fatalf("TestGHOSTDAG: dag.ProcessBlock got unexpected error for block %s: %v", blockData.ID, err)
				}
				if isDelayed {
					t.Fatalf("TestGHOSTDAG: block %s "+
						"is too far in the future", blockData.ID)
				}
				if isOrphan {
					t.Fatalf("TestGHOSTDAG: block %s was unexpectedly orphan", blockData.ID)
				}

				node, ok := dag.index.LookupNode(utilBlock.Hash())
				if !ok {
					t.Fatalf("block %s does not exist in the DAG", utilBlock.Hash())
				}

				blockByIDMap[blockData.ID] = node
				idByBlockMap[node] = blockData.ID

				bluesIDs := make([]string, 0, len(node.blues))
				for _, blue := range node.blues {
					bluesIDs = append(bluesIDs, idByBlockMap[blue])
				}
				selectedParentID := idByBlockMap[node.selectedParent]
				fullDataStr := fmt.Sprintf("blues: %v, selectedParent: %v, score: %v",
					bluesIDs, selectedParentID, node.blueScore)
				if blockData.ExpectedScore != node.blueScore {
					t.Errorf("Test %s: Block %s expected to have score %v but got %v (fulldata: %v)",
						info.Name(), blockData.ID, blockData.ExpectedScore, node.blueScore, fullDataStr)
				}
				if blockData.ExpectedSelectedParent != selectedParentID {
					t.Errorf("Test %s: Block %s expected to have selected parent %v but got %v (fulldata: %v)",
						info.Name(), blockData.ID, blockData.ExpectedSelectedParent, selectedParentID, fullDataStr)
				}
				if !reflect.DeepEqual(blockData.ExpectedBlues, bluesIDs) {
					t.Errorf("Test %s: Block %s expected to have blues %v but got %v (fulldata: %v)",
						info.Name(), blockData.ID, blockData.ExpectedBlues, bluesIDs, fullDataStr)
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
			if !checkReds(test.ExpectedReds, reds) {
				redsIDs := make([]string, 0, len(reds))
				for id := range reds {
					redsIDs = append(redsIDs, id)
				}
				sort.Strings(redsIDs)
				sort.Strings(test.ExpectedReds)
				t.Errorf("Test %s: Expected reds %v but got %v", info.Name(), test.ExpectedReds, redsIDs)
			}
		}()

		return nil
	})

	if err != nil {
		t.Fatal(err)
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
	currentBlockA := dag.Params.GenesisBlock
	for i := dagconfig.KType(0); i < dag.Params.K; i++ {
		newBlock := prepareAndProcessBlockByParentMsgBlocks(t, dag, currentBlockA)
		currentBlockA = newBlock
	}

	// Prepare another block chain with size K beginning with the genesis block
	currentBlockB := dag.Params.GenesisBlock
	for i := dagconfig.KType(0); i < dag.Params.K; i++ {
		newBlock := prepareAndProcessBlockByParentMsgBlocks(t, dag, currentBlockB)
		currentBlockB = newBlock
	}

	// Get references to the tips of the two chains
	blockNodeA, ok := dag.index.LookupNode(currentBlockA.BlockHash())
	if !ok {
		t.Fatalf("block %s does not exist in the DAG", currentBlockA.BlockHash())
	}

	blockNodeB, ok := dag.index.LookupNode(currentBlockB.BlockHash())
	if !ok {
		t.Fatalf("block %s does not exist in the DAG", currentBlockB.BlockHash())
	}

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
	block1 := prepareAndProcessBlockByParentMsgBlocks(t, dag, dag.Params.GenesisBlock)
	block2 := prepareAndProcessBlockByParentMsgBlocks(t, dag, dag.Params.GenesisBlock)

	// Add a child block to the previous two blocks
	block3 := prepareAndProcessBlockByParentMsgBlocks(t, dag, block1, block2)

	// Clear the reachability store
	dag.reachabilityTree.store.loaded = map[daghash.Hash]*reachabilityData{}

	dbTx, err := dag.databaseContext.NewTx()
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
	blockNode3, ok := dag.index.LookupNode(block3.BlockHash())
	if !ok {
		t.Fatalf("block %s does not exist in the DAG", block3.BlockHash())
	}
	_, err = dag.ghostdag(blockNode3)
	if err == nil {
		t.Fatalf("TestGHOSTDAGErrors: ghostdag unexpectedly succeeded")
	}
	expectedErrSubstring := "couldn't find reachability data"
	if !strings.Contains(err.Error(), expectedErrSubstring) {
		t.Fatalf("TestGHOSTDAGErrors: ghostdag returned wrong error. "+
			"Want: %s, got: %s", expectedErrSubstring, err)
	}
}
