package consensusstatemanager_test

import (
	"sort"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
)

func TestConsensusStateManager_pickVirtualParents(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		stagingArea := model.NewStagingArea()

		tc, teardown, err := consensus.NewFactory().NewTestConsensus(params, false, "TestConsensusStateManager_pickVirtualParents")
		if err != nil {
			t.Fatalf("Error setting up tc: %+v", err)
		}
		defer teardown(false)

		getSortedVirtualParents := func(tc testapi.TestConsensus) []*externalapi.DomainHash {
			virtualRelations, err := tc.BlockRelationStore().BlockRelation(tc.DatabaseContext(), stagingArea, model.VirtualBlockHash)
			if err != nil {
				t.Fatalf("Failed getting virtual block virtualRelations: %v", err)
			}

			block, err := tc.BuildBlock(&externalapi.DomainCoinbaseData{ScriptPublicKey: &externalapi.ScriptPublicKey{Script: nil, Version: 0}}, nil)
			if err != nil {
				t.Fatalf("Consensus failed building a block: %v", err)
			}
			blockParents := block.Header.ParentHashes()
			sort.Sort(testutils.NewTestGhostDAGSorter(stagingArea, virtualRelations.Parents, tc, t))
			sort.Sort(testutils.NewTestGhostDAGSorter(stagingArea, blockParents, tc, t))
			if !externalapi.HashesEqual(virtualRelations.Parents, blockParents) {
				t.Fatalf("Block relations and BuildBlock return different parents for virtual, %s != %s", virtualRelations.Parents, blockParents)
			}
			return virtualRelations.Parents
		}

		// We build 2*params.MaxBlockParents each one with blueWork higher than the other.
		parents := make([]*externalapi.DomainHash, 0, params.MaxBlockParents)
		for i := 0; i < 2*int(params.MaxBlockParents); i++ {
			lastBlock := params.GenesisHash
			for j := 0; j <= i; j++ {
				lastBlock, _, err = tc.AddBlock([]*externalapi.DomainHash{lastBlock}, nil, nil)
				if err != nil {
					t.Fatalf("Failed Adding block to tc: %v", err)
				}
			}
			parents = append(parents, lastBlock)
		}

		virtualParents := getSortedVirtualParents(tc)
		sort.Sort(testutils.NewTestGhostDAGSorter(stagingArea, parents, tc, t))

		// Make sure the first half of the blocks are with highest blueWork
		// we use (max+1)/2 because the first "half" is rounded up, so `(dividend + (divisor - 1)) / divisor` = `(max + (2-1))/2` = `(max+1)/2`
		for i := 0; i < int(params.MaxBlockParents+1)/2; i++ {
			if !virtualParents[i].Equal(parents[i]) {
				t.Fatalf("Expected block at %d to be equal, instead found %s != %s", i, virtualParents[i], parents[i])
			}
		}

		// Make sure the second half is the candidates with lowest blueWork
		end := len(parents) - int(params.MaxBlockParents)/2
		for i := (params.MaxBlockParents + 1) / 2; i < params.MaxBlockParents; i++ {
			if !virtualParents[i].Equal(parents[end]) {
				t.Fatalf("Expected block at %d to be equal, instead found %s != %s", i, virtualParents[i], parents[end])
			}
			end++
		}
		if end != len(parents) {
			t.Fatalf("Expected %d==%d", end, len(parents))
		}

		// Clear all tips.
		var virtualSelectedParent *externalapi.DomainHash
		for {
			block, err := tc.BuildBlock(&externalapi.DomainCoinbaseData{ScriptPublicKey: &externalapi.ScriptPublicKey{Script: nil, Version: 0}, ExtraData: nil}, nil)
			if err != nil {
				t.Fatalf("Failed building a block: %v", err)
			}
			_, err = tc.ValidateAndInsertBlock(block)
			if err != nil {
				t.Fatalf("Failed Inserting block to tc: %v", err)
			}
			virtualSelectedParent = consensushashing.BlockHash(block)
			if len(block.Header.ParentHashes()) == 1 {
				break
			}
		}
		// build exactly params.MaxBlockParents
		parents = make([]*externalapi.DomainHash, 0, params.MaxBlockParents)
		for i := 0; i < int(params.MaxBlockParents); i++ {
			block, _, err := tc.AddBlock([]*externalapi.DomainHash{virtualSelectedParent}, nil, nil)
			if err != nil {
				t.Fatalf("Failed Adding block to tc: %v", err)
			}
			parents = append(parents, block)
		}

		sort.Sort(testutils.NewTestGhostDAGSorter(stagingArea, parents, tc, t))
		virtualParents = getSortedVirtualParents(tc)
		if !externalapi.HashesEqual(virtualParents, parents) {
			t.Fatalf("Expected VirtualParents and parents to be equal, instead: %s != %s", virtualParents, parents)
		}
	})
}
