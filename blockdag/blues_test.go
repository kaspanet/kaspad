package blockdag

import (
	"fmt"
	"testing"
	"time"

	"github.com/daglabs/btcd/dagconfig"
)

type testBlockData struct {
	parents                []string
	id                     string
	expectedScore          int64
	expectedSelectedParent string
	expectedBlues          []string
}

func TestBlues(t *testing.T) {
	netParams := &dagconfig.SimNetParams

	blockVersion := int32(0x20000000)

	tests := []struct {
		k       uint //TODO: for now it doesn't matter, and it just takes from dagParams
		dagData []testBlockData
	}{
		{
			k: 1,
			dagData: []testBlockData{
				{
					parents:                []string{"A"},
					id:                     "B",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"A"},
					id:                     "C",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"B"},
					id:                     "D",
					expectedScore:          2,
					expectedSelectedParent: "B",
					expectedBlues:          []string{"B"},
				},
				{
					parents:                []string{"B"},
					id:                     "E",
					expectedScore:          2,
					expectedSelectedParent: "B",
					expectedBlues:          []string{"B"},
				},
				{
					parents:                []string{"C"},
					id:                     "F",
					expectedScore:          2,
					expectedSelectedParent: "C",
					expectedBlues:          []string{"C"},
				},
				{
					parents:                []string{"C", "D"},
					id:                     "G",
					expectedScore:          4,
					expectedSelectedParent: "C",
					expectedBlues:          []string{"D", "B", "C"},
				},
				{
					parents:                []string{"C", "E"},
					id:                     "H",
					expectedScore:          4,
					expectedSelectedParent: "C",
					expectedBlues:          []string{"E", "B", "C"},
				},
				{
					parents:                []string{"E", "G"},
					id:                     "I",
					expectedScore:          5,
					expectedSelectedParent: "E",
					expectedBlues:          []string{"G", "D", "E"},
				},
				{
					parents:                []string{"F"},
					id:                     "J",
					expectedScore:          3,
					expectedSelectedParent: "F",
					expectedBlues:          []string{"F"},
				},
			},
		},
	}

	for _, test := range tests {
		// Generate enough synthetic blocks for the rest of the test
		blockDag := newFakeDAG(netParams)
		genesisNode := blockDag.dag.SelectedTip()
		blockTime := genesisNode.Header().Timestamp
		blockIDMap := make(map[string]*blockNode)
		idBlockMap := make(map[*blockNode]string)
		blockIDMap["A"] = genesisNode
		idBlockMap[genesisNode] = "A"

		checkBlues := func(expected []string, got []string) bool {
			if len(expected) != len(got) {
				return false
			}
			for i, expectedID := range expected {
				if expectedID != got[i] {
					return false
				}
			}
			return true
		}

		for _, blockData := range test.dagData {
			fmt.Printf("Block %v test:\n", blockData.id)
			blockTime = blockTime.Add(time.Second)
			parents := blockSet{}
			for _, parentID := range blockData.parents {
				parent := blockIDMap[parentID]
				parents.add(parent)
			}
			node := newFakeNode(parents, blockVersion, 0, blockTime)
			fmt.Printf("hash: %v \n", node.hash)

			blockDag.index.AddNode(node)
			blockIDMap[blockData.id] = node
			idBlockMap[node] = blockData.id

			bluesIDs := make([]string, 0, len(node.blues))
			for _, blue := range node.blues {
				bluesIDs = append(bluesIDs, idBlockMap[blue])
			}
			selectedParentID := idBlockMap[node.selectedParent]
			fullDataStr := fmt.Sprintf("blues: %v, selectedParent: %v, score: %v", bluesIDs, selectedParentID, node.blueScore)
			if blockData.expectedScore != node.blueScore {
				t.Errorf("Block %v expected to have score %v but got %v (fulldata: %v)", blockData.id, blockData.expectedScore, node.blueScore, fullDataStr)
				continue
			}
			if blockData.expectedSelectedParent != selectedParentID {
				t.Errorf("Block %v expected to have selected parent %v but got %v (fulldata: %v)", blockData.id, blockData.expectedSelectedParent, selectedParentID, fullDataStr)
				continue
			}
			if !checkBlues(blockData.expectedBlues, bluesIDs) {
				t.Errorf("Block %v expected to have blues %v but got %v (fulldata: %v)", blockData.id, blockData.expectedBlues, bluesIDs, fullDataStr)
				continue
			}
			fmt.Printf("\n")
		}

		// for _, blockData := range test.dagData {
		// 	node := blockIDMap[blockData.id]
		// 	bluesIDs := make([]string, len(blues))
		// 	for i, blue := range blues {
		// 		bluesIDs[i] = idBlockMap[blue]
		// 	}
		// 	selectedParentID := idBlockMap[selectedParent]
		// 	fullDataStr := fmt.Sprintf("blues: %v, selectedParent: %v, score: %v", bluesIDs, selectedParentID, score)
		// 	if blockData.expectedScore != score {
		// 		t.Errorf("Block %v expected to have score %v but got %v (%v)", blockData.id, blockData.expectedScore, score, fullDataStr)
		// 		continue
		// 	}
		// 	if blockData.expectedSelectedParent != selectedParentID {
		// 		t.Errorf("Block %v expected to have selected parent %v but got %v", blockData.id, blockData.expectedSelectedParent, selectedParentID)
		// 		continue
		// 	}
		// 	if !checkBlues(blockData.expectedBlues, bluesIDs) {
		// 		t.Errorf("Block %v expected to have blues %v but got %v", blockData.id, blockData.expectedBlues, blues)
		// 		continue
		// 	}
		// }
	}
}

func addNode(blockDag *BlockDAG, node *blockNode) {
	blockDag.index.AddNode(node)

}
