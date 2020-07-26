package blockdag

import (
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
	"reflect"
	"testing"
	"time"
)

func TestBlueBlockWindow(t *testing.T) {
	params := dagconfig.SimnetParams
	params.K = 1
	dag, teardownFunc, err := DAGSetup("TestBlueBlockWindow", true, Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("Failed to setup dag instance: %v", err)
	}
	defer teardownFunc()

	resetExtraNonceForTest()

	windowSize := uint64(10)
	genesisNode := dag.genesis
	blockTime := genesisNode.Header().Timestamp
	blockByIDMap := make(map[string]*blockNode)
	idByBlockMap := make(map[*blockNode]string)
	blockByIDMap["A"] = genesisNode
	idByBlockMap[genesisNode] = "A"

	blocksData := []*struct {
		parents                          []string
		id                               string //id is a virtual entity that is used only for tests so we can define relations between blocks without knowing their hash
		expectedWindowWithGenesisPadding []string
	}{
		{
			parents:                          []string{"A"},
			id:                               "B",
			expectedWindowWithGenesisPadding: []string{"A", "A", "A", "A", "A", "A", "A", "A", "A", "A"},
		},
		{
			parents:                          []string{"B"},
			id:                               "C",
			expectedWindowWithGenesisPadding: []string{"B", "A", "A", "A", "A", "A", "A", "A", "A", "A"},
		},
		{
			parents:                          []string{"B"},
			id:                               "D",
			expectedWindowWithGenesisPadding: []string{"B", "A", "A", "A", "A", "A", "A", "A", "A", "A"},
		},
		{
			parents:                          []string{"D", "C"},
			id:                               "E",
			expectedWindowWithGenesisPadding: []string{"D", "C", "B", "A", "A", "A", "A", "A", "A", "A"},
		},
		{
			parents:                          []string{"D", "C"},
			id:                               "F",
			expectedWindowWithGenesisPadding: []string{"D", "C", "B", "A", "A", "A", "A", "A", "A", "A"},
		},
		{
			parents:                          []string{"A"},
			id:                               "G",
			expectedWindowWithGenesisPadding: []string{"A", "A", "A", "A", "A", "A", "A", "A", "A", "A"},
		},
		{
			parents:                          []string{"G"},
			id:                               "H",
			expectedWindowWithGenesisPadding: []string{"G", "A", "A", "A", "A", "A", "A", "A", "A", "A"},
		},
		{
			parents:                          []string{"H", "F"},
			id:                               "I",
			expectedWindowWithGenesisPadding: []string{"F", "D", "C", "B", "A", "A", "A", "A", "A", "A"},
		},
		{
			parents:                          []string{"I"},
			id:                               "J",
			expectedWindowWithGenesisPadding: []string{"I", "F", "D", "C", "B", "A", "A", "A", "A", "A"},
		},
		{
			parents:                          []string{"J"},
			id:                               "K",
			expectedWindowWithGenesisPadding: []string{"J", "I", "F", "D", "C", "B", "A", "A", "A", "A"},
		},
		{
			parents:                          []string{"K"},
			id:                               "L",
			expectedWindowWithGenesisPadding: []string{"K", "J", "I", "F", "D", "C", "B", "A", "A", "A"},
		},
		{
			parents:                          []string{"L"},
			id:                               "M",
			expectedWindowWithGenesisPadding: []string{"L", "K", "J", "I", "F", "D", "C", "B", "A", "A"},
		},
		{
			parents:                          []string{"M"},
			id:                               "N",
			expectedWindowWithGenesisPadding: []string{"M", "L", "K", "J", "I", "F", "D", "C", "B", "A"},
		},
		{
			parents:                          []string{"N"},
			id:                               "O",
			expectedWindowWithGenesisPadding: []string{"N", "M", "L", "K", "J", "I", "F", "D", "C", "B"},
		},
	}

	for _, blockData := range blocksData {
		blockTime = blockTime.Add(time.Second)
		parents := blockSet{}
		for _, parentID := range blockData.parents {
			parent := blockByIDMap[parentID]
			parents.add(parent)
		}

		block, err := PrepareBlockForTest(dag, parents.hashes(), nil)
		if err != nil {
			t.Fatalf("block %v got unexpected error from PrepareBlockForTest: %v", blockData.id, err)
		}

		utilBlock := util.NewBlock(block)
		isOrphan, isDelayed, err := dag.ProcessBlock(utilBlock, BFNoPoWCheck)
		if err != nil {
			t.Fatalf("dag.ProcessBlock got unexpected error for block %v: %v", blockData.id, err)
		}
		if isDelayed {
			t.Fatalf("block %s "+
				"is too far in the future", blockData.id)
		}
		if isOrphan {
			t.Fatalf("block %v was unexpectedly orphan", blockData.id)
		}

		node, ok := dag.index.LookupNode(utilBlock.Hash())
		if !ok {
			t.Fatalf("block %s does not exist in the DAG", utilBlock.Hash())
		}

		blockByIDMap[blockData.id] = node
		idByBlockMap[node] = blockData.id

		window := blueBlockWindow(node, windowSize)
		if err := checkWindowIDs(window, blockData.expectedWindowWithGenesisPadding, idByBlockMap); err != nil {
			t.Errorf("Unexpected values for window for block %s: %s", blockData.id, err)
		}
	}
}

func checkWindowIDs(window []*blockNode, expectedIDs []string, idByBlockMap map[*blockNode]string) error {
	ids := make([]string, len(window))
	for i, node := range window {
		ids[i] = idByBlockMap[node]
	}
	if !reflect.DeepEqual(ids, expectedIDs) {
		return errors.Errorf("window expected to have blocks %s but got %s", expectedIDs, ids)
	}
	return nil
}
