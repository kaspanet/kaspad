package blockdag

import (
	"github.com/daglabs/kaspad/dagconfig"
	"github.com/daglabs/kaspad/util/daghash"
	"github.com/pkg/errors"
	"reflect"
	"testing"
	"time"
)

func TestBlueBlockWindow(t *testing.T) {
	params := dagconfig.SimNetParams
	params.K = 1
	dag := newTestDAG(&params)

	windowSize := uint64(10)
	genesisNode := dag.genesis
	blockTime := genesisNode.Header().Timestamp
	blockByIDMap := make(map[string]*blockNode)
	idByBlockMap := make(map[*blockNode]string)
	blockByIDMap["A"] = genesisNode
	idByBlockMap[genesisNode] = "A"
	blockVersion := int32(0x10000000)

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
			parents:                          []string{"C", "D"},
			id:                               "E",
			expectedWindowWithGenesisPadding: []string{"D", "C", "B", "A", "A", "A", "A", "A", "A", "A"},
		},
		{
			parents:                          []string{"C", "D"},
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
		node := newTestNode(parents, blockVersion, 0, blockTime, dag.dagParams.K)
		node.hash = &daghash.Hash{} // It helps to predict hash order
		for i, char := range blockData.id {
			node.hash[i] = byte(char)
		}

		dag.index.AddNode(node)
		node.updateParentsChildren()

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
