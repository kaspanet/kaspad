package ghost

import (
	"encoding/json"
	"fmt"
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"os"
	"reflect"
	"testing"
)

func TestGHOST(t *testing.T) {
	testChain := []struct {
		parents            []string
		id                 string
		expectedGHOSTChain []string
	}{
		{
			parents:            []string{"A"},
			id:                 "B",
			expectedGHOSTChain: []string{"A", "B"},
		},
		{
			parents:            []string{"B"},
			id:                 "C",
			expectedGHOSTChain: []string{"A", "B", "C"},
		},
		{
			parents:            []string{"B"},
			id:                 "D",
			expectedGHOSTChain: []string{"A", "B", "D"},
		},
		{
			parents:            []string{"C", "D"},
			id:                 "E",
			expectedGHOSTChain: []string{"A", "B", "D", "E"},
		},
		{
			parents:            []string{"C", "D"},
			id:                 "F",
			expectedGHOSTChain: []string{"A", "B", "D", "F"},
		},
		{
			parents:            []string{"A"},
			id:                 "G",
			expectedGHOSTChain: []string{"A", "B", "D", "F"},
		},
		{
			parents:            []string{"G"},
			id:                 "H",
			expectedGHOSTChain: []string{"A", "B", "D", "F"},
		},
		{
			parents:            []string{"H", "F"},
			id:                 "I",
			expectedGHOSTChain: []string{"A", "B", "D", "F", "I"},
		},
		{
			parents:            []string{"I"},
			id:                 "J",
			expectedGHOSTChain: []string{"A", "B", "D", "F", "I", "J"},
		},
	}

	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()
		tc, tearDown, err := factory.NewTestConsensus(consensusConfig, "TestBlockWindow")
		if err != nil {
			t.Fatalf("NewTestConsensus: %s", err)
		}
		defer tearDown(false)

		blockByIDMap := make(map[string]*externalapi.DomainHash)
		idByBlockMap := make(map[externalapi.DomainHash]string)
		blockByIDMap["A"] = consensusConfig.GenesisHash
		idByBlockMap[*consensusConfig.GenesisHash] = "A"
		mostRecentHash := consensusConfig.GenesisHash

		for _, blockData := range testChain {
			parents := hashset.New()
			for _, parentID := range blockData.parents {
				parent := blockByIDMap[parentID]
				parents.Add(parent)
			}

			blockHash := addBlockWithHashSmallerThan(t, tc, parents.ToSlice(), mostRecentHash)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}
			blockByIDMap[blockData.id] = blockHash
			idByBlockMap[*blockHash] = blockData.id
			mostRecentHash = blockHash

			subDAG := convertDAGtoSubDAG(t, consensusConfig, tc)
			ghostChainHashes := GHOST(subDAG)
			ghostChainIDs := make([]string, len(ghostChainHashes))
			for i, ghostChainHash := range ghostChainHashes {
				ghostChainIDs[i] = idByBlockMap[*ghostChainHash]
			}
			if !reflect.DeepEqual(ghostChainIDs, blockData.expectedGHOSTChain) {
				t.Errorf("After adding block ID %s, GHOST chain expected to have IDs %s but got IDs %s",
					blockData.id, blockData.expectedGHOSTChain, ghostChainIDs)
			}
		}
	})
}

// addBlockWithHashSmallerThan adds a block to the DAG with the given parents such that its
// hash is smaller than `maxHash`. This ensures that the GHOST chain calculated from the
// DAG is deterministic
func addBlockWithHashSmallerThan(t *testing.T, tc testapi.TestConsensus,
	parentHashes []*externalapi.DomainHash, maxHash *externalapi.DomainHash) *externalapi.DomainHash {

	var block *externalapi.DomainBlock
	blockHash := maxHash
	for maxHash.LessOrEqual(blockHash) {
		var err error
		block, _, err = tc.BuildBlockWithParents(parentHashes, nil, nil)
		if err != nil {
			t.Fatalf("BuildBlockWithParents: %+v", err)
		}
		blockHash = consensushashing.BlockHash(block)
	}

	_, err := tc.ValidateAndInsertBlock(block, true)
	if err != nil {
		t.Fatalf("ValidateAndInsertBlock: %+v", err)
	}

	return blockHash
}

func convertDAGtoSubDAG(t *testing.T, consensusConfig *consensus.Config, tc testapi.TestConsensus) *model.SubDAG {
	genesisHash := consensusConfig.GenesisHash

	stagingArea := model.NewStagingArea()
	tipHashes, err := tc.ConsensusStateStore().Tips(stagingArea, tc.DatabaseContext())
	if err != nil {
		t.Fatalf("Tips: %+v", err)
	}

	subDAG := &model.SubDAG{
		RootHashes: []*externalapi.DomainHash{genesisHash},
		TipHashes:  tipHashes,
		Blocks:     map[externalapi.DomainHash]*model.SubDAGBlock{},
	}

	visited := hashset.New()
	queue := tc.DAGTraversalManager().NewDownHeap(stagingArea)
	err = queue.PushSlice(tipHashes)
	if err != nil {
		t.Fatalf("PushSlice: %+v", err)
	}
	for queue.Len() > 0 {
		blockHash := queue.Pop()
		visited.Add(blockHash)
		dagChildHashes, err := tc.DAGTopologyManager().Children(stagingArea, blockHash)
		if err != nil {
			t.Fatalf("Children: %+v", err)
		}
		childHashes := []*externalapi.DomainHash{}
		for _, dagChildHash := range dagChildHashes {
			if dagChildHash.Equal(model.VirtualBlockHash) {
				continue
			}
			childHashes = append(childHashes, dagChildHash)
		}

		dagParentHashes, err := tc.DAGTopologyManager().Parents(stagingArea, blockHash)
		if err != nil {
			t.Fatalf("Parents: %+v", err)
		}
		parentHashes := []*externalapi.DomainHash{}
		for _, dagParentHash := range dagParentHashes {
			if dagParentHash.Equal(model.VirtualGenesisBlockHash) {
				continue
			}
			parentHashes = append(parentHashes, dagParentHash)
			if !visited.Contains(dagParentHash) {
				err := queue.Push(dagParentHash)
				if err != nil {
					t.Fatalf("Push: %+v", err)
				}
			}
		}

		subDAG.Blocks[*blockHash] = &model.SubDAGBlock{
			BlockHash:    blockHash,
			ParentHashes: parentHashes,
			ChildHashes:  childHashes,
		}
	}
	return subDAG
}

type jsonBlock struct {
	ID      string   `json:"ID"`
	Parents []string `json:"Parents"`
}

type testJSON struct {
	Blocks []*jsonBlock `json:"blocks"`
}

func BenchmarkGHOST(b *testing.B) {
	b.StopTimer()

	// Load JSON
	jsonFile, err := os.Open("benchmark_data.json")
	if err != nil {
		b.Fatalf("Open: %+v", err)
	}
	defer jsonFile.Close()

	test := &testJSON{}
	decoder := json.NewDecoder(jsonFile)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&test)
	if err != nil {
		b.Fatalf("Decode: %+v", err)
	}

	// Convert JSON data to a SubDAG
	subDAG := &model.SubDAG{
		RootHashes: []*externalapi.DomainHash{},
		TipHashes:  []*externalapi.DomainHash{},
		Blocks:     make(map[externalapi.DomainHash]*model.SubDAGBlock, len(test.Blocks)),
	}
	blockIDToHash := func(blockID string) *externalapi.DomainHash {
		blockHashHex := fmt.Sprintf("%064s", blockID)
		blockHash, err := externalapi.NewDomainHashFromString(blockHashHex)
		if err != nil {
			b.Fatalf("NewDomainHashFromString: %+v", err)
		}
		return blockHash
	}
	for _, block := range test.Blocks {
		blockHash := blockIDToHash(block.ID)

		parentHashes := []*externalapi.DomainHash{}
		for _, parentID := range block.Parents {
			parentHash := blockIDToHash(parentID)
			parentHashes = append(parentHashes, parentHash)
		}

		subDAG.Blocks[*blockHash] = &model.SubDAGBlock{
			BlockHash:    blockHash,
			ParentHashes: parentHashes,
			ChildHashes:  []*externalapi.DomainHash{},
		}
	}
	for _, block := range subDAG.Blocks {
		for _, parentHash := range block.ParentHashes {
			parentBlock := subDAG.Blocks[*parentHash]
			parentBlock.ChildHashes = append(parentBlock.ChildHashes, block.BlockHash)
		}
	}
	for _, block := range subDAG.Blocks {
		if len(block.ParentHashes) == 0 {
			subDAG.RootHashes = append(subDAG.RootHashes, block.BlockHash)
		}
		if len(block.ChildHashes) == 0 {
			subDAG.TipHashes = append(subDAG.TipHashes, block.BlockHash)
		}
	}

	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		GHOST(subDAG)
	}
}
