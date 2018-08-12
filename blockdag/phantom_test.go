package blockdag

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/daglabs/btcd/dagconfig/daghash"

	"github.com/daglabs/btcd/dagconfig"
)

type testBlockData struct {
	parents                []string
	id                     string //id is a virtual entity that is used only for tests so we can define relations between blocks without knowing their hash
	expectedScore          uint64
	expectedSelectedParent string
	expectedBlues          []string
}

type hashIDPair struct {
	hash *daghash.Hash
	id   string
}

//TestPhantom iterate over several dag simulations, and checks
//that the blue score, blue set and selected parent of each
//block calculated as expected
func TestPhantom(t *testing.T) {
	netParams := dagconfig.SimNetParams

	blockVersion := int32(0x20000000)

	tests := []struct {
		k              uint32
		dagData        []*testBlockData
		virtualBlockID string
		expectedReds   []string
	}{
		{
			//Block hash order:DEBHICAKGJF
			k:              1,
			virtualBlockID: "K",
			expectedReds:   []string{"D"},
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
					expectedSelectedParent: "G",
					expectedBlues:          []string{"G"},
				},
				{
					parents:                []string{"F"},
					id:                     "J",
					expectedScore:          3,
					expectedSelectedParent: "F",
					expectedBlues:          []string{"F"},
				},
				{
					parents:                []string{"H", "I", "J"},
					id:                     "K",
					expectedScore:          9,
					expectedSelectedParent: "H",
					expectedBlues:          []string{"I", "G", "J", "F", "H"},
				},
			},
		},
		{
			//block hash order:DQKRLHOEBSIGUJNPCMTAFV
			k:              2,
			virtualBlockID: "V",
			expectedReds:   []string{"D", "J", "P"},
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
					expectedBlues:          []string{"L", "K", "H", "I", "E", "G", "B", "F"},
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
				{
					parents:                []string{"S", "U"},
					id:                     "V",
					expectedScore:          18,
					expectedSelectedParent: "S",
					expectedBlues:          []string{"U", "T", "S"},
				},
			},
		},
		{
			//Block hash order:NRSHBUXJTFGPDVCKEQIOWLMA
			k:              1,
			virtualBlockID: "X",
			expectedReds:   []string{"D", "F", "G", "H", "J", "K", "L", "N", "O", "Q", "R", "S", "U", "V"},
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
					expectedBlues:          []string{"I", "M", "C", "E", "B"},
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
					expectedSelectedParent: "L",
					expectedBlues:          []string{"K", "L"},
				},
				{
					parents:                []string{"U", "R"},
					id:                     "V",
					expectedScore:          6,
					expectedSelectedParent: "U",
					expectedBlues:          []string{"R", "U"},
				},
				{
					parents:                []string{"S", "U", "T"},
					id:                     "W",
					expectedScore:          8,
					expectedSelectedParent: "T",
					expectedBlues:          []string{"T"},
				},
				{
					parents:                []string{"V", "W", "H"},
					id:                     "X",
					expectedScore:          9,
					expectedSelectedParent: "W",
					expectedBlues:          []string{"W"},
				},
			},
		},
		{
			//Secret mining attack: The attacker is mining
			//blocks B,C,D,E,F,G,T in secret without propagating
			//them, so all blocks except T should be red, because
			//they don't follow the rules of PHANTOM that require
			//you to point to all the parents that you know, and
			//propagate your block as soon as it's mined

			//Block hash order: HRTGMKQBXDWSICYFONUPLEAJZ
			k:              1,
			virtualBlockID: "Y",
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
					expectedSelectedParent: "J",
					expectedBlues:          []string{"K", "L", "J"},
				},
				{
					parents:                []string{"J", "K", "L"},
					id:                     "N",
					expectedScore:          6,
					expectedSelectedParent: "J",
					expectedBlues:          []string{"K", "L", "J"},
				},
				{
					parents:                []string{"N", "M"},
					id:                     "O",
					expectedScore:          8,
					expectedSelectedParent: "N",
					expectedBlues:          []string{"M", "N"},
				},
				{
					parents:                []string{"N", "M"},
					id:                     "P",
					expectedScore:          8,
					expectedSelectedParent: "N",
					expectedBlues:          []string{"M", "N"},
				},
				{
					parents:                []string{"N", "M"},
					id:                     "Q",
					expectedScore:          8,
					expectedSelectedParent: "N",
					expectedBlues:          []string{"M", "N"},
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
					id:                     "V",
					expectedScore:          15,
					expectedSelectedParent: "U",
					expectedBlues:          []string{"T", "U"},
				},
				{
					parents:                []string{"T", "U"},
					id:                     "W",
					expectedScore:          15,
					expectedSelectedParent: "U",
					expectedBlues:          []string{"T", "U"},
				},
				{
					parents:                []string{"T", "U"},
					id:                     "X",
					expectedScore:          15,
					expectedSelectedParent: "U",
					expectedBlues:          []string{"T", "U"},
				},
				{
					parents:                []string{"V", "W", "X"},
					id:                     "Y",
					expectedScore:          18,
					expectedSelectedParent: "X",
					expectedBlues:          []string{"W", "V", "X"},
				},
			},
		},
		{
			//Censorship mining attack: The attacker is mining blocks B,C,D,E,F,G in secret without propagating them,
			//so all blocks except B should be red, because they don't follow the rules of
			//PHANTOM that require you to point to all the parents that you know

			//Block hash order:WZHOGBJMDSICRUYKTFQLEAPXN
			k:              1,
			virtualBlockID: "Y",
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
					expectedBlues:          []string{"H", "B", "I"},
				},
				{
					parents:                []string{"H", "I", "B"},
					id:                     "K",
					expectedScore:          4,
					expectedSelectedParent: "I",
					expectedBlues:          []string{"H", "B", "I"},
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
					expectedBlues:          []string{"J", "L", "K"},
				},
				{
					parents:                []string{"J", "K", "L", "C"},
					id:                     "N",
					expectedScore:          7,
					expectedSelectedParent: "K",
					expectedBlues:          []string{"J", "L", "K"},
				},
				{
					parents:                []string{"N", "M", "D"},
					id:                     "O",
					expectedScore:          9,
					expectedSelectedParent: "N",
					expectedBlues:          []string{"M", "N"},
				},
				{
					parents:                []string{"N", "M", "D"},
					id:                     "P",
					expectedScore:          9,
					expectedSelectedParent: "N",
					expectedBlues:          []string{"M", "N"},
				},
				{
					parents:                []string{"N", "M", "D"},
					id:                     "Q",
					expectedScore:          9,
					expectedSelectedParent: "N",
					expectedBlues:          []string{"M", "N"},
				},
				{
					parents:                []string{"O", "P", "Q", "E"},
					id:                     "R",
					expectedScore:          12,
					expectedSelectedParent: "P",
					expectedBlues:          []string{"O", "Q", "P"},
				},
				{
					parents:                []string{"O", "P", "Q", "E"},
					id:                     "S",
					expectedScore:          12,
					expectedSelectedParent: "P",
					expectedBlues:          []string{"O", "Q", "P"},
				},
				{
					parents:                []string{"G", "S", "R"},
					id:                     "T",
					expectedScore:          14,
					expectedSelectedParent: "R",
					expectedBlues:          []string{"S", "R"},
				},
				{
					parents:                []string{"S", "R", "F"},
					id:                     "U",
					expectedScore:          14,
					expectedSelectedParent: "R",
					expectedBlues:          []string{"S", "R"},
				},
				{
					parents:                []string{"T", "U"},
					id:                     "V",
					expectedScore:          16,
					expectedSelectedParent: "T",
					expectedBlues:          []string{"U", "T"},
				},
				{
					parents:                []string{"T", "U"},
					id:                     "W",
					expectedScore:          16,
					expectedSelectedParent: "T",
					expectedBlues:          []string{"U", "T"},
				},
				{
					parents:                []string{"T", "U"},
					id:                     "X",
					expectedScore:          16,
					expectedSelectedParent: "T",
					expectedBlues:          []string{"U", "T"},
				},
				{
					parents:                []string{"V", "W", "X"},
					id:                     "Y",
					expectedScore:          19,
					expectedSelectedParent: "W",
					expectedBlues:          []string{"V", "X", "W"},
				},
			},
		},
	}

	for i, test := range tests {
		netParams.K = test.k
		// Generate enough synthetic blocks for the rest of the test
		blockDAG := newTestDAG(&netParams)
		genesisNode := blockDAG.virtual.SelectedTip()
		blockTime := genesisNode.Header().Timestamp
		blockByIDMap := make(map[string]*blockNode)
		idByBlockMap := make(map[*blockNode]string)
		blockByIDMap["A"] = genesisNode
		idByBlockMap[genesisNode] = "A"

		for _, blockData := range test.dagData {
			blockTime = blockTime.Add(time.Second)
			parents := blockSet{}
			for _, parentID := range blockData.parents {
				parent := blockByIDMap[parentID]
				parents.add(parent)
			}
			node := newTestNode(parents, blockVersion, 0, blockTime, test.k)

			blockDAG.index.AddNode(node)
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

		for tip := blockByIDMap[test.virtualBlockID]; tip.selectedParent != nil; tip = tip.selectedParent {
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
