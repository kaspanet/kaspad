package blockdag

import (
	"fmt"
	"sort"
	"strings"
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
		k       uint //TODO: for now it doesn't matter, and it just takes from dagParams
		dagData []*testBlockData
	}{
		// {
		// 	//Block hash order:IJDFGBHCEA
		// 	k: 1,
		// 	dagData: []*testBlockData{
		// 		{
		// 			parents:                []string{"A"},
		// 			id:                     "B",
		// 			expectedScore:          1,
		// 			expectedSelectedParent: "A",
		// 			expectedBlues:          []string{"A"},
		// 		},
		// 		{
		// 			parents:                []string{"A"},
		// 			id:                     "C",
		// 			expectedScore:          1,
		// 			expectedSelectedParent: "A",
		// 			expectedBlues:          []string{"A"},
		// 		},
		// 		{
		// 			parents:                []string{"B"},
		// 			id:                     "D",
		// 			expectedScore:          2,
		// 			expectedSelectedParent: "B",
		// 			expectedBlues:          []string{"B"},
		// 		},
		// 		{
		// 			parents:                []string{"B"},
		// 			id:                     "E",
		// 			expectedScore:          2,
		// 			expectedSelectedParent: "B",
		// 			expectedBlues:          []string{"B"},
		// 		},
		// 		{
		// 			parents:                []string{"C"},
		// 			id:                     "F",
		// 			expectedScore:          2,
		// 			expectedSelectedParent: "C",
		// 			expectedBlues:          []string{"C"},
		// 		},
		// 		{
		// 			parents:                []string{"C", "D"},
		// 			id:                     "G",
		// 			expectedScore:          4,
		// 			expectedSelectedParent: "C",
		// 			expectedBlues:          []string{"D", "B", "C"},
		// 		},
		// 		{
		// 			parents:                []string{"C", "E"},
		// 			id:                     "H",
		// 			expectedScore:          4,
		// 			expectedSelectedParent: "C",
		// 			expectedBlues:          []string{"E", "B", "C"},
		// 		},
		// 		{
		// 			parents:                []string{"E", "G"},
		// 			id:                     "I",
		// 			expectedScore:          5,
		// 			expectedSelectedParent: "E",
		// 			expectedBlues:          []string{"G", "D", "E"},
		// 		},
		// 		{
		// 			parents:                []string{"F"},
		// 			id:                     "J",
		// 			expectedScore:          3,
		// 			expectedSelectedParent: "F",
		// 			expectedBlues:          []string{"F"},
		// 		},
		// 	},
		// },
		// {
		// 	//block hash order:LOSPUHDFNIBRGKJQTCMEA
		// 	k: 2,
		// 	dagData: []*testBlockData{
		// 		{
		// 			parents:                []string{"A"},
		// 			id:                     "B",
		// 			expectedScore:          1,
		// 			expectedSelectedParent: "A",
		// 			expectedBlues:          []string{"A"},
		// 		},
		// 		{
		// 			parents:                []string{"A"},
		// 			id:                     "C",
		// 			expectedScore:          1,
		// 			expectedSelectedParent: "A",
		// 			expectedBlues:          []string{"A"},
		// 		},
		// 		{
		// 			parents:                []string{"B"},
		// 			id:                     "D",
		// 			expectedScore:          2,
		// 			expectedSelectedParent: "B",
		// 			expectedBlues:          []string{"B"},
		// 		},
		// 		{
		// 			parents:                []string{"B"},
		// 			id:                     "E",
		// 			expectedScore:          2,
		// 			expectedSelectedParent: "B",
		// 			expectedBlues:          []string{"B"},
		// 		},
		// 		{
		// 			parents:                []string{"C"},
		// 			id:                     "F",
		// 			expectedScore:          2,
		// 			expectedSelectedParent: "C",
		// 			expectedBlues:          []string{"C"},
		// 		},
		// 		{
		// 			parents:                []string{"C"},
		// 			id:                     "G",
		// 			expectedScore:          2,
		// 			expectedSelectedParent: "C",
		// 			expectedBlues:          []string{"C"},
		// 		},
		// 		{
		// 			parents:                []string{"G"},
		// 			id:                     "H",
		// 			expectedScore:          3,
		// 			expectedSelectedParent: "G",
		// 			expectedBlues:          []string{"G"},
		// 		},
		// 		{
		// 			parents:                []string{"E"},
		// 			id:                     "I",
		// 			expectedScore:          3,
		// 			expectedSelectedParent: "E",
		// 			expectedBlues:          []string{"E"},
		// 		},
		// 		{
		// 			parents:                []string{"E"},
		// 			id:                     "J",
		// 			expectedScore:          3,
		// 			expectedSelectedParent: "E",
		// 			expectedBlues:          []string{"E"},
		// 		},
		// 		{
		// 			parents:                []string{"I"},
		// 			id:                     "K",
		// 			expectedScore:          4,
		// 			expectedSelectedParent: "I",
		// 			expectedBlues:          []string{"I"},
		// 		},
		// 		{
		// 			parents:                []string{"K", "H"},
		// 			id:                     "L",
		// 			expectedScore:          5,
		// 			expectedSelectedParent: "K",
		// 			expectedBlues:          []string{"K"},
		// 		},
		// 		{
		// 			parents:                []string{"F", "L"},
		// 			id:                     "M",
		// 			expectedScore:          10,
		// 			expectedSelectedParent: "F",
		// 			expectedBlues:          []string{"L", "K", "H", "I", "G", "E", "B", "F"},
		// 		},
		// 		{
		// 			parents:                []string{"G", "K"},
		// 			id:                     "N",
		// 			expectedScore:          7,
		// 			expectedSelectedParent: "G",
		// 			expectedBlues:          []string{"K", "I", "E", "B", "G"},
		// 		},
		// 		{
		// 			parents:                []string{"J", "N"},
		// 			id:                     "O",
		// 			expectedScore:          8,
		// 			expectedSelectedParent: "N",
		// 			expectedBlues:          []string{"N"},
		// 		},
		// 		{
		// 			parents:                []string{"D"},
		// 			id:                     "P",
		// 			expectedScore:          3,
		// 			expectedSelectedParent: "D",
		// 			expectedBlues:          []string{"D"},
		// 		},
		// 		{
		// 			parents:                []string{"O", "P"},
		// 			id:                     "Q",
		// 			expectedScore:          10,
		// 			expectedSelectedParent: "P",
		// 			expectedBlues:          []string{"O", "N", "K", "I", "J", "E", "P"},
		// 		},
		// 		{
		// 			parents:                []string{"L", "Q"},
		// 			id:                     "R",
		// 			expectedScore:          11,
		// 			expectedSelectedParent: "Q",
		// 			expectedBlues:          []string{"Q"},
		// 		},
		// 		{
		// 			parents:                []string{"M", "R"},
		// 			id:                     "S",
		// 			expectedScore:          15,
		// 			expectedSelectedParent: "M",
		// 			expectedBlues:          []string{"R", "Q", "O", "N", "M"},
		// 		},
		// 		{
		// 			parents:                []string{"H", "F"},
		// 			id:                     "T",
		// 			expectedScore:          5,
		// 			expectedSelectedParent: "H",
		// 			expectedBlues:          []string{"F", "H"},
		// 		},
		// 		{
		// 			parents:                []string{"M", "T"},
		// 			id:                     "U",
		// 			expectedScore:          12,
		// 			expectedSelectedParent: "M",
		// 			expectedBlues:          []string{"T", "M"},
		// 		},
		// 	},
		// },
		{
			//Block hash order: JTVNIFGWDBQLREUHMPSCOKA
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
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"B"},
					id:                     "G",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"C"},
					id:                     "H",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"C"},
					id:                     "I",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"B"},
					id:                     "J",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"D"},
					id:                     "K",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"D"},
					id:                     "L",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"E"},
					id:                     "M",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"E"},
					id:                     "N",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"F", "G", "J"},
					id:                     "O",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"B", "M", "I"},
					id:                     "P",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"K", "E"},
					id:                     "Q",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"L", "N"},
					id:                     "R",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"I", "Q"},
					id:                     "S",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"K", "P"},
					id:                     "T",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"K", "L"},
					id:                     "U",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"U", "R"},
					id:                     "V",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
				{
					parents:                []string{"S", "U", "T"},
					id:                     "W",
					expectedScore:          1,
					expectedSelectedParent: "A",
					expectedBlues:          []string{"A"},
				},
			},
		},
	}

	for _, test := range tests {
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
			fmt.Printf("Block %v test:\n", blockData.id)
			blockTime = blockTime.Add(time.Second)
			parents := blockSet{}
			for _, parentID := range blockData.parents {
				parent := blockIDMap[parentID]
				parents.add(parent)
			}
			node := newFakeNode(parents, blockVersion, 0, blockTime)
			node.id = blockData.id
			fmt.Printf("hash %v: %v \n", blockData.id, node.hash)

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

		pairs := make([]*hashIDPair, 0, len(blockIDMap))

		for id, node := range blockIDMap {
			pairs = append(pairs, &hashIDPair{
				id:   id,
				hash: &node.hash,
			})
		}

		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].hash.Cmp(pairs[j].hash) > 0
		})

		fmt.Printf("Block hash order:")

		for _, pair := range pairs {
			fmt.Print(pair.id)
		}
		fmt.Printf("\n")

	}
}

func (bs blockSet) StringByID() string {
	ids := []string{}
	for _, node := range bs {
		id := "A"
		if node.id != "" {
			id = node.id
		}
		ids = append(ids, id)
	}
	return strings.Join(ids, ",")
}
