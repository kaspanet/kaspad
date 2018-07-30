package blockdag

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/daglabs/btcd/dagconfig/daghash"

	"github.com/daglabs/btcd/dagconfig"
)

type testBlockData struct {
	parents                []string
	id                     string
	expectedScore          int64
	expectedSelectedParent string
	expectedBlues          []string
}

type hashIDPair struct {
	hash *daghash.Hash
	id   string
}

func TestBlues(t *testing.T) {
	netParams := &dagconfig.SimNetParams

	blockVersion := int32(0x20000000)

	tests := []struct {
		k              uint
		dagData        []*testBlockData
		virtualBlockID string
		expectedReds   []string
	}{
		{
			//Block hash order:IJDFGBHCEA
			k: 1,
			dagData: []*testBlockData{
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
		{
			//block hash order:LOSPUHDFNIBRGKJQTCMEA
			k: 2,
			dagData: []*testBlockData{
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
					parents:                []string{"C"},
					id:                     "G",
					expectedScore:          2,
					expectedSelectedParent: "C",
					expectedBlues:          []string{"C"},
				},
				{
					parents:                []string{"G"},
					id:                     "H",
					expectedScore:          3,
					expectedSelectedParent: "G",
					expectedBlues:          []string{"G"},
				},
				{
					parents:                []string{"E"},
					id:                     "I",
					expectedScore:          3,
					expectedSelectedParent: "E",
					expectedBlues:          []string{"E"},
				},
				{
					parents:                []string{"E"},
					id:                     "J",
					expectedScore:          3,
					expectedSelectedParent: "E",
					expectedBlues:          []string{"E"},
				},
				{
					parents:                []string{"I"},
					id:                     "K",
					expectedScore:          4,
					expectedSelectedParent: "I",
					expectedBlues:          []string{"I"},
				},
				{
					parents:                []string{"K", "H"},
					id:                     "L",
					expectedScore:          5,
					expectedSelectedParent: "K",
					expectedBlues:          []string{"K"},
				},
				{
					parents:                []string{"F", "L"},
					id:                     "M",
					expectedScore:          10,
					expectedSelectedParent: "F",
					expectedBlues:          []string{"L", "K", "H", "I", "G", "E", "B", "F"},
				},
				{
					parents:                []string{"G", "K"},
					id:                     "N",
					expectedScore:          7,
					expectedSelectedParent: "G",
					expectedBlues:          []string{"K", "I", "E", "B", "G"},
				},
				{
					parents:                []string{"J", "N"},
					id:                     "O",
					expectedScore:          8,
					expectedSelectedParent: "N",
					expectedBlues:          []string{"N"},
				},
				{
					parents:                []string{"D"},
					id:                     "P",
					expectedScore:          3,
					expectedSelectedParent: "D",
					expectedBlues:          []string{"D"},
				},
				{
					parents:                []string{"O", "P"},
					id:                     "Q",
					expectedScore:          10,
					expectedSelectedParent: "P",
					expectedBlues:          []string{"O", "N", "K", "I", "J", "E", "P"},
				},
				{
					parents:                []string{"L", "Q"},
					id:                     "R",
					expectedScore:          11,
					expectedSelectedParent: "Q",
					expectedBlues:          []string{"Q"},
				},
				{
					parents:                []string{"M", "R"},
					id:                     "S",
					expectedScore:          15,
					expectedSelectedParent: "M",
					expectedBlues:          []string{"R", "Q", "O", "N", "M"},
				},
				{
					parents:                []string{"H", "F"},
					id:                     "T",
					expectedScore:          5,
					expectedSelectedParent: "F",
					expectedBlues:          []string{"H", "G", "F"},
				},
				{
					parents:                []string{"M", "T"},
					id:                     "U",
					expectedScore:          12,
					expectedSelectedParent: "M",
					expectedBlues:          []string{"T", "M"},
				},
			},
		},
		{
			//Block hash order: JTVNIFGWDBQLREUHMPSCOKA
			k:              1,
			virtualBlockID: "X",
			expectedReds:   []string{"B", "C", "E", "F", "G", "H", "I", "J", "M", "N", "O", "P", "R"},
			dagData: []*testBlockData{
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
					parents:                []string{"A"},
					id:                     "D",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"A"},
					id:                     "E",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"B"},
					id:                     "F",
					expectedScore:          2,
					expectedSelectedParent: "B",
					expectedBlues:          []string{"B"},
				},
				{
					parents:                []string{"B"},
					id:                     "G",
					expectedScore:          2,
					expectedSelectedParent: "B",
					expectedBlues:          []string{"B"},
				},
				{
					parents:                []string{"C"},
					id:                     "H",
					expectedScore:          2,
					expectedSelectedParent: "C",
					expectedBlues:          []string{"C"},
				},
				{
					parents:                []string{"C"},
					id:                     "I",
					expectedScore:          2,
					expectedSelectedParent: "C",
					expectedBlues:          []string{"C"},
				},
				{
					parents:                []string{"B"},
					id:                     "J",
					expectedScore:          2,
					expectedSelectedParent: "B",
					expectedBlues:          []string{"B"},
				},
				{
					parents:                []string{"D"},
					id:                     "K",
					expectedScore:          2,
					expectedSelectedParent: "D",
					expectedBlues:          []string{"D"},
				},
				{
					parents:                []string{"D"},
					id:                     "L",
					expectedScore:          2,
					expectedSelectedParent: "D",
					expectedBlues:          []string{"D"},
				},
				{
					parents:                []string{"E"},
					id:                     "M",
					expectedScore:          2,
					expectedSelectedParent: "E",
					expectedBlues:          []string{"E"},
				},
				{
					parents:                []string{"E"},
					id:                     "N",
					expectedScore:          2,
					expectedSelectedParent: "E",
					expectedBlues:          []string{"E"},
				},
				{
					parents:                []string{"F", "G", "J"},
					id:                     "O",
					expectedScore:          5,
					expectedSelectedParent: "G",
					expectedBlues:          []string{"J", "F", "G"},
				},
				{
					parents:                []string{"B", "M", "I"},
					id:                     "P",
					expectedScore:          6,
					expectedSelectedParent: "B",
					expectedBlues:          []string{"I", "M", "E", "C", "B"},
				},
				{
					parents:                []string{"K", "E"},
					id:                     "Q",
					expectedScore:          4,
					expectedSelectedParent: "E",
					expectedBlues:          []string{"K", "D", "E"},
				},
				{
					parents:                []string{"L", "N"},
					id:                     "R",
					expectedScore:          3,
					expectedSelectedParent: "L",
					expectedBlues:          []string{"L"},
				},
				{
					parents:                []string{"I", "Q"},
					id:                     "S",
					expectedScore:          5,
					expectedSelectedParent: "Q",
					expectedBlues:          []string{"Q"},
				},
				{
					parents:                []string{"K", "P"},
					id:                     "T",
					expectedScore:          7,
					expectedSelectedParent: "P",
					expectedBlues:          []string{"P"},
				},
				{
					parents:                []string{"K", "L"},
					id:                     "U",
					expectedScore:          4,
					expectedSelectedParent: "K",
					expectedBlues:          []string{"L", "K"},
				},
				{
					parents:                []string{"U", "R"},
					id:                     "V",
					expectedScore:          5,
					expectedSelectedParent: "U",
					expectedBlues:          []string{"U"},
				},
				{
					parents:                []string{"S", "U", "T"},
					id:                     "W",
					expectedScore:          8,
					expectedSelectedParent: "U",
					expectedBlues:          []string{"T", "S", "Q", "U"},
				},
				{
					parents:                []string{"V", "W", "H"},
					id:                     "X",
					expectedScore:          10,
					expectedSelectedParent: "W",
					expectedBlues:          []string{"V", "W"},
				},
			},
		},
		{
			//Secret mining attack
			//Block hash order: LNRWXUFJQMBGOSPDCHTIZKAYE
			k:              1,
			virtualBlockID: "Z",
			expectedReds:   []string{"B", "C", "D", "E", "F", "G"},
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
					parents:                []string{"C"},
					id:                     "D",
					expectedScore:          3,
					expectedSelectedParent: "C",
					expectedBlues:          []string{"C"},
				},
				{
					parents:                []string{"D"},
					id:                     "E",
					expectedScore:          4,
					expectedSelectedParent: "D",
					expectedBlues:          []string{"D"},
				},
				{
					parents:                []string{"E"},
					id:                     "F",
					expectedScore:          5,
					expectedSelectedParent: "E",
					expectedBlues:          []string{"E"},
				},
				{
					parents:                []string{"F"},
					id:                     "G",
					expectedScore:          6,
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
					parents:                []string{"H", "I"},
					id:                     "J",
					expectedScore:          3,
					expectedSelectedParent: "I",
					expectedBlues:          []string{"H", "I"},
				},
				{
					parents:                []string{"H", "I"},
					id:                     "K",
					expectedScore:          3,
					expectedSelectedParent: "I",
					expectedBlues:          []string{"H", "I"},
				},
				{
					parents:                []string{"I"},
					id:                     "L",
					expectedScore:          2,
					expectedSelectedParent: "I",
					expectedBlues:          []string{"I"},
				},
				{
					parents:                []string{"J", "K", "L"},
					id:                     "M",
					expectedScore:          6,
					expectedSelectedParent: "K",
					expectedBlues:          []string{"L", "J", "K"},
				},
				{
					parents:                []string{"J", "K", "L"},
					id:                     "N",
					expectedScore:          6,
					expectedSelectedParent: "K",
					expectedBlues:          []string{"L", "J", "K"},
				},
				{
					parents:                []string{"N", "M"},
					id:                     "O",
					expectedScore:          8,
					expectedSelectedParent: "M",
					expectedBlues:          []string{"N", "M"},
				},
				{
					parents:                []string{"N", "M"},
					id:                     "P",
					expectedScore:          8,
					expectedSelectedParent: "M",
					expectedBlues:          []string{"N", "M"},
				},
				{
					parents:                []string{"N", "M"},
					id:                     "Q",
					expectedScore:          8,
					expectedSelectedParent: "M",
					expectedBlues:          []string{"N", "M"},
				},
				{
					parents:                []string{"O", "P", "Q"},
					id:                     "R",
					expectedScore:          11,
					expectedSelectedParent: "P",
					expectedBlues:          []string{"Q", "O", "P"},
				},
				{
					parents:                []string{"O", "P", "Q"},
					id:                     "S",
					expectedScore:          11,
					expectedSelectedParent: "P",
					expectedBlues:          []string{"Q", "O", "P"},
				},
				{
					parents:                []string{"G", "S", "R"},
					id:                     "T",
					expectedScore:          13,
					expectedSelectedParent: "S",
					expectedBlues:          []string{"R", "S"},
				},
				{
					parents:                []string{"S", "R"},
					id:                     "U",
					expectedScore:          13,
					expectedSelectedParent: "S",
					expectedBlues:          []string{"R", "S"},
				},
				{
					parents:                []string{"T", "U"},
					id:                     "W",
					expectedScore:          15,
					expectedSelectedParent: "T",
					expectedBlues:          []string{"U", "T"},
				},
				{
					parents:                []string{"T", "U"},
					id:                     "X",
					expectedScore:          15,
					expectedSelectedParent: "T",
					expectedBlues:          []string{"U", "T"},
				},
				{
					parents:                []string{"T", "U"},
					id:                     "Y",
					expectedScore:          15,
					expectedSelectedParent: "T",
					expectedBlues:          []string{"U", "T"},
				},
				{
					parents:                []string{"W", "X", "Y"},
					id:                     "Z",
					expectedScore:          18,
					expectedSelectedParent: "Y",
					expectedBlues:          []string{"W", "X", "Y"},
				},
			},
		},
		{
			//Censorship mining attack
			//Block hash order:LJNRFYPOKSBGZMXTDCHQIUAWE
			k:              1,
			virtualBlockID: "Z",
			expectedReds:   []string{"C", "D", "E", "F", "G"},
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
					parents:                []string{"C"},
					id:                     "D",
					expectedScore:          3,
					expectedSelectedParent: "C",
					expectedBlues:          []string{"C"},
				},
				{
					parents:                []string{"D"},
					id:                     "E",
					expectedScore:          4,
					expectedSelectedParent: "D",
					expectedBlues:          []string{"D"},
				},
				{
					parents:                []string{"E"},
					id:                     "F",
					expectedScore:          5,
					expectedSelectedParent: "E",
					expectedBlues:          []string{"E"},
				},
				{
					parents:                []string{"F"},
					id:                     "G",
					expectedScore:          6,
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
					parents:                []string{"H", "I", "B"},
					id:                     "J",
					expectedScore:          4,
					expectedSelectedParent: "I",
					expectedBlues:          []string{"B", "H", "I"},
				},
				{
					parents:                []string{"H", "I", "B"},
					id:                     "K",
					expectedScore:          4,
					expectedSelectedParent: "I",
					expectedBlues:          []string{"B", "H", "I"},
				},
				{
					parents:                []string{"I"},
					id:                     "L",
					expectedScore:          2,
					expectedSelectedParent: "I",
					expectedBlues:          []string{"I"},
				},
				{
					parents:                []string{"J", "K", "L", "C"},
					id:                     "M",
					expectedScore:          7,
					expectedSelectedParent: "K",
					expectedBlues:          []string{"L", "J", "K"},
				},
				{
					parents:                []string{"J", "K", "L", "C"},
					id:                     "N",
					expectedScore:          7,
					expectedSelectedParent: "K",
					expectedBlues:          []string{"L", "J", "K"},
				},
				{
					parents:                []string{"N", "M", "D"},
					id:                     "O",
					expectedScore:          9,
					expectedSelectedParent: "M",
					expectedBlues:          []string{"N", "M"},
				},
				{
					parents:                []string{"N", "M", "D"},
					id:                     "P",
					expectedScore:          9,
					expectedSelectedParent: "M",
					expectedBlues:          []string{"N", "M"},
				},
				{
					parents:                []string{"N", "M", "D"},
					id:                     "Q",
					expectedScore:          9,
					expectedSelectedParent: "M",
					expectedBlues:          []string{"N", "M"},
				},
				{
					parents:                []string{"O", "P", "Q", "E"},
					id:                     "R",
					expectedScore:          12,
					expectedSelectedParent: "Q",
					expectedBlues:          []string{"P", "O", "Q"},
				},
				{
					parents:                []string{"O", "P", "Q", "E"},
					id:                     "S",
					expectedScore:          12,
					expectedSelectedParent: "Q",
					expectedBlues:          []string{"P", "O", "Q"},
				},
				{
					parents:                []string{"G", "S", "R"},
					id:                     "T",
					expectedScore:          14,
					expectedSelectedParent: "S",
					expectedBlues:          []string{"R", "S"},
				},
				{
					parents:                []string{"S", "R", "F"},
					id:                     "U",
					expectedScore:          14,
					expectedSelectedParent: "S",
					expectedBlues:          []string{"R", "S"},
				},
				{
					parents:                []string{"T", "U"},
					id:                     "W",
					expectedScore:          16,
					expectedSelectedParent: "U",
					expectedBlues:          []string{"T", "U"},
				},
				{
					parents:                []string{"T", "U"},
					id:                     "X",
					expectedScore:          16,
					expectedSelectedParent: "U",
					expectedBlues:          []string{"T", "U"},
				},
				{
					parents:                []string{"T", "U"},
					id:                     "Y",
					expectedScore:          16,
					expectedSelectedParent: "U",
					expectedBlues:          []string{"T", "U"},
				},
				{
					parents:                []string{"W", "X", "Y"},
					id:                     "Z",
					expectedScore:          19,
					expectedSelectedParent: "W",
					expectedBlues:          []string{"Y", "X", "W"},
				},
			},
		},
	}

	for testNum, test := range tests {
		errorF := func(format string, args ...interface{}) {
			newArgs := make([]interface{}, 0, len(args)+1)
			newArgs = append(newArgs, testNum)
			for _, arg := range args {
				newArgs = append(newArgs, arg)
			}
			t.Errorf("Test %d: "+format, newArgs...)
		}
		phantomK = test.k
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
			blockTime = blockTime.Add(time.Second)
			parents := blockSet{}
			for _, parentID := range blockData.parents {
				parent := blockIDMap[parentID]
				parents.add(parent)
			}
			node := newFakeNode(parents, blockVersion, 0, blockTime)

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
				errorF("Block %v expected to have score %v but got %v (fulldata: %v)", blockData.id, blockData.expectedScore, node.blueScore, fullDataStr)
				continue
			}
			if blockData.expectedSelectedParent != selectedParentID {
				errorF("Block %v expected to have selected parent %v but got %v (fulldata: %v)", blockData.id, blockData.expectedSelectedParent, selectedParentID, fullDataStr)
				continue
			}
			if !checkBlues(blockData.expectedBlues, bluesIDs) {
				errorF("Block %v expected to have blues %v but got %v (fulldata: %v)", blockData.id, blockData.expectedBlues, bluesIDs, fullDataStr)
				continue
			}
		}

		if test.expectedReds != nil {
			reds := make(map[string]bool)

			checkReds := func() bool {
				if len(test.expectedReds) != len(reds) {
					return false
				}
				for _, redID := range test.expectedReds {
					if !reds[redID] {
						return false
					}
				}
				return true
			}

			for id := range blockIDMap {
				reds[id] = true
			}

			for tip := blockIDMap[test.virtualBlockID]; tip.selectedParent != nil; tip = tip.selectedParent {
				tipID := idBlockMap[tip]
				delete(reds, tipID)
				for _, blue := range tip.blues {
					blueID := idBlockMap[blue]
					delete(reds, blueID)
				}
			}
			if !checkReds() {
				redsIDs := make([]string, 0, len(reds))
				for id := range reds {
					redsIDs = append(redsIDs, id)
				}
				sort.Strings(redsIDs)
				sort.Strings(test.expectedReds)
				errorF("Expected reds %v but got %v", test.expectedReds, redsIDs)
			}
		}

		// pairs := make([]*hashIDPair, 0, len(blockIDMap))

		// for id, node := range blockIDMap {
		// 	pairs = append(pairs, &hashIDPair{
		// 		id:   id,
		// 		hash: &node.hash,
		// 	})
		// }

		// sort.Slice(pairs, func(i, j int) bool {
		// 	return pairs[i].hash.Cmp(pairs[j].hash) > 0
		// })

		// fmt.Printf("Block hash order:")

		// for _, pair := range pairs {
		// 	fmt.Print(pair.id)
		// }
		// fmt.Printf("\n")

	}
}
