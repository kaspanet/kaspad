package blockdag

import (
	"fmt"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/util/daghash"
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
		parents                             []string
		id                                  string //id is a virtual entity that is used only for tests so we can define relations between blocks without knowing their hash
		expectedWindowWithoutGenesisPadding []string
		expectedWindowWithGenesisPadding    []string
		expectedOKWithoutGenesisPadding     bool
	}{
		{
			parents:                             []string{"A"},
			id:                                  "B",
			expectedWindowWithGenesisPadding:    []string{"A", "A", "A", "A", "A", "A", "A", "A", "A", "A"},
			expectedWindowWithoutGenesisPadding: nil,
			expectedOKWithoutGenesisPadding:     false,
		},
		{
			parents:                             []string{"B"},
			id:                                  "C",
			expectedWindowWithGenesisPadding:    []string{"B", "A", "A", "A", "A", "A", "A", "A", "A", "A"},
			expectedWindowWithoutGenesisPadding: nil,
			expectedOKWithoutGenesisPadding:     false,
		},
		{
			parents:                             []string{"B"},
			id:                                  "D",
			expectedWindowWithGenesisPadding:    []string{"B", "A", "A", "A", "A", "A", "A", "A", "A", "A"},
			expectedWindowWithoutGenesisPadding: nil,
			expectedOKWithoutGenesisPadding:     false,
		},
		{
			parents:                             []string{"C", "D"},
			id:                                  "E",
			expectedWindowWithGenesisPadding:    []string{"D", "C", "B", "A", "A", "A", "A", "A", "A", "A"},
			expectedWindowWithoutGenesisPadding: nil,
			expectedOKWithoutGenesisPadding:     false,
		},
		{
			parents:                             []string{"C", "D"},
			id:                                  "F",
			expectedWindowWithGenesisPadding:    []string{"D", "C", "B", "A", "A", "A", "A", "A", "A", "A"},
			expectedWindowWithoutGenesisPadding: nil,
			expectedOKWithoutGenesisPadding:     false,
		},
		{
			parents:                             []string{"A"},
			id:                                  "G",
			expectedWindowWithGenesisPadding:    []string{"A", "A", "A", "A", "A", "A", "A", "A", "A", "A"},
			expectedWindowWithoutGenesisPadding: nil,
			expectedOKWithoutGenesisPadding:     false,
		},
		{
			parents:                             []string{"G"},
			id:                                  "H",
			expectedWindowWithGenesisPadding:    []string{"G", "A", "A", "A", "A", "A", "A", "A", "A", "A"},
			expectedWindowWithoutGenesisPadding: nil,
			expectedOKWithoutGenesisPadding:     false,
		},
		{
			parents:                             []string{"H", "F"},
			id:                                  "I",
			expectedWindowWithGenesisPadding:    []string{"F", "D", "C", "B", "A", "A", "A", "A", "A", "A"},
			expectedWindowWithoutGenesisPadding: nil,
			expectedOKWithoutGenesisPadding:     false,
		},
		{
			parents:                             []string{"I"},
			id:                                  "J",
			expectedWindowWithGenesisPadding:    []string{"I", "F", "D", "C", "B", "A", "A", "A", "A", "A"},
			expectedWindowWithoutGenesisPadding: nil,
			expectedOKWithoutGenesisPadding:     false,
		},
		{
			parents:                             []string{"J"},
			id:                                  "K",
			expectedWindowWithGenesisPadding:    []string{"J", "I", "F", "D", "C", "B", "A", "A", "A", "A"},
			expectedWindowWithoutGenesisPadding: nil,
			expectedOKWithoutGenesisPadding:     false,
		},
		{
			parents:                             []string{"K"},
			id:                                  "L",
			expectedWindowWithGenesisPadding:    []string{"K", "J", "I", "F", "D", "C", "B", "A", "A", "A"},
			expectedWindowWithoutGenesisPadding: nil,
			expectedOKWithoutGenesisPadding:     false,
		},
		{
			parents:                             []string{"L"},
			id:                                  "M",
			expectedWindowWithGenesisPadding:    []string{"L", "K", "J", "I", "F", "D", "C", "B", "A", "A"},
			expectedWindowWithoutGenesisPadding: nil,
			expectedOKWithoutGenesisPadding:     false,
		},
		{
			parents:                             []string{"M"},
			id:                                  "N",
			expectedWindowWithGenesisPadding:    []string{"M", "L", "K", "J", "I", "F", "D", "C", "B", "A"},
			expectedWindowWithoutGenesisPadding: []string{"M", "L", "K", "J", "I", "F", "D", "C", "B", "A"},
			expectedOKWithoutGenesisPadding:     true,
		},
		{
			parents:                             []string{"N"},
			id:                                  "O",
			expectedWindowWithGenesisPadding:    []string{"N", "M", "L", "K", "J", "I", "F", "D", "C", "B"},
			expectedWindowWithoutGenesisPadding: []string{"N", "M", "L", "K", "J", "I", "F", "D", "C", "B"},
			expectedOKWithoutGenesisPadding:     true,
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

		window, ok := blueBlockWindow(node, windowSize, true)
		if !ok {
			t.Errorf("when padWithGenesis is set to true, ok should always be true")
		}
		if err := checkWindowIDs(window, blockData.expectedWindowWithGenesisPadding, idByBlockMap); err != nil {
			t.Errorf("Unexpected values for window with genesis padding for block %s: %s", blockData.id, err)
		}

		window, ok = blueBlockWindow(node, windowSize, false)
		if ok != blockData.expectedOKWithoutGenesisPadding {
			t.Errorf("Unexpected ok value for window without genesis padding for block %s: expected ok to be %t but got %t", blockData.id, blockData.expectedOKWithoutGenesisPadding, ok)
		}
		if ok {
			if err := checkWindowIDs(window, blockData.expectedWindowWithoutGenesisPadding, idByBlockMap); err != nil {
				t.Errorf("Unexpected values for widnow without genesis padding for block %s: %s", blockData.id, err)
			}
		}
	}
}

func checkWindowIDs(window []*blockNode, expectedIDs []string, idByBlockMap map[*blockNode]string) error {
	if len(window) != len(expectedIDs) {

	}
	ids := make([]string, len(window))
	for i, node := range window {
		ids[i] = idByBlockMap[node]
	}
	if !reflect.DeepEqual(ids, expectedIDs) {
		return fmt.Errorf("window expected to have blocks %s but got %s", expectedIDs, ids)
	}
	return nil
}
