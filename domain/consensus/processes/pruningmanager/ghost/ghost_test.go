package ghost

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
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
			expectedGHOSTChain: []string{"A", "B", "C"},
		},
		{
			parents:            []string{"C", "D"},
			id:                 "E",
			expectedGHOSTChain: []string{"A", "B", "C", "E"},
		},
		{
			parents:            []string{"C", "D"},
			id:                 "F",
			expectedGHOSTChain: []string{"A", "B", "C", "E"},
		},
		{
			parents:            []string{"A"},
			id:                 "G",
			expectedGHOSTChain: []string{"A", "B", "C", "E"},
		},
		{
			parents:            []string{"G"},
			id:                 "H",
			expectedGHOSTChain: []string{"A", "B", "C", "E"},
		},
		{
			parents:            []string{"H", "F"},
			id:                 "I",
			expectedGHOSTChain: []string{"A", "B", "C", "F", "I"},
		},
		{
			parents:            []string{"I"},
			id:                 "J",
			expectedGHOSTChain: []string{"A", "B", "C", "F", "I", "J"},
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

		for _, blockData := range testChain {
			parents := hashset.New()
			for _, parentID := range blockData.parents {
				parent := blockByIDMap[parentID]
				parents.Add(parent)
			}

			blockHash, _, err := tc.AddBlock(parents.ToSlice(), nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}
			blockByIDMap[blockData.id] = blockHash
			idByBlockMap[*blockHash] = blockData.id

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

func convertDAGtoSubDAG(t *testing.T, consensusConfig *consensus.Config, tc testapi.TestConsensus) *model.SubDAG {
	genesisHash := consensusConfig.GenesisHash

	stagingArea := model.NewStagingArea()
	tipHashes, err := tc.ConsensusStateStore().Tips(stagingArea, tc.DatabaseContext())
	if err != nil {
		t.Fatalf("Tips: %+v", err)
	}

	subDAG := &model.SubDAG{
		GenesisHash: genesisHash,
		TipHashes:   tipHashes,
		Blocks:      map[externalapi.DomainHash]*model.SubDAGBlock{},
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
