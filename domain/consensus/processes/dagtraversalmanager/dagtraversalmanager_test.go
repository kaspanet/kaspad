package dagtraversalmanager_test

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"testing"
)

func TestBlockAtDepth(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		factory := consensus.NewFactory()
		tc, tearDown, err := factory.NewTestConsensus(params, false,
			"TestBlockAtDepth")
		if err != nil {
			t.Fatalf("NewTestConsensus: %s", err)
		}
		defer tearDown(false)

		// "checkExpectedBlockChainDag" - compare BlockAtDepth to a loop on the SelectedChainIterator on a single chain DAG.
		checkExpectedBlockChainDag := func(highHash *externalapi.DomainHash, depth uint64) {
			currentBlock := highHash
			currentBlockData, err := tc.GHOSTDAGDataStore().Get(tc.DatabaseContext(), currentBlock)
			if err != nil {
				t.Fatalf("GHOSTDAGDataStore().Get: %+v", err)
			}
			for i := uint64(0); i <= depth; i++ {
				if currentBlockData.SelectedParent() == nil {
					break
				}
				currentBlock = currentBlockData.SelectedParent()
				currentBlockData, err = tc.GHOSTDAGDataStore().Get(tc.DatabaseContext(), currentBlock)
				if err != nil {
					t.Fatalf("GHOSTDAGDataStore().Get: %+v", err)
				}
			}
			expected := currentBlock
			blockHash, err := tc.DAGTraversalManager().BlockAtDepth(highHash, depth)
			if err != nil {
				t.Fatalf("BlockAtDepth: %+v", err)
			}

			if !blockHash.Equal(expected) {
				t.Fatalf("Expected block %s but got %s", expected, blockHash)
			}
		}

		// "checkExpectedBlockTwoChildrenSameSelectedParent" checks blockAtDepth on 2 children with the same selectedParent.
		checkExpectedBlockTwoChildrenSameSelectedParent := func(firstChild, secondChild *externalapi.DomainHash, depth uint64) {
			blockHash, err := tc.DAGTraversalManager().BlockAtDepth(firstChild, depth)
			if err != nil {
				t.Fatalf("BlockAtDepth: %+v", err)
			}
			expectedSameHash, err := tc.DAGTraversalManager().BlockAtDepth(secondChild, depth)
			if err != nil {
				t.Fatalf("BlockAtDepth: %+v", err)
			}
			if !blockHash.Equal(expectedSameHash) {
				t.Fatalf("Expected block %s but got %s", expectedSameHash, blockHash)
			}
		}

		// "checkExpectedBlockTwoChildrenSameSelectedParent" checks on different chains, on the same DAG, that they merge at the correct point.
		checkExpectedBlockOnTwoDiffChain := func(firstChild, secondChild *externalapi.DomainHash, firstDepth, secondDepth uint64) {
			blockHash, err := tc.DAGTraversalManager().BlockAtDepth(firstChild, firstDepth)
			if err != nil {
				t.Fatalf("BlockAtDepth: %+v", err)
			}
			expectedSameHash, err := tc.DAGTraversalManager().BlockAtDepth(secondChild, secondDepth)
			if err != nil {
				t.Fatalf("BlockAtDepth: %+v", err)
			}
			if !blockHash.Equal(expectedSameHash) {
				t.Fatalf("Expected block %s but got %s", expectedSameHash, blockHash)
			}
			expectedDiffHash, err := tc.DAGTraversalManager().BlockAtDepth(secondChild, secondDepth-1)
			if err != nil {
				t.Fatalf("BlockAtDepth: %+v", err)
			}
			if blockHash.Equal(expectedDiffHash) {
				t.Fatalf("Expected to differente block")
			}

		}

		highHash, err := createAChainDag(params.GenesisHash, tc)
		if err != nil {
			t.Fatalf("createAChainDagInBlockAtDepth: %+v", err)
		}
		checkExpectedBlockChainDag(highHash, 2)

		firstChild, secondChild, err := expendDagToTwoChildrenSameSelectedParent(highHash, tc)
		if err != nil {
			t.Fatalf("createADagInBlockAtDepth: %+v", err)
		}
		checkExpectedBlockTwoChildrenSameSelectedParent(firstChild, secondChild, 2)

		firstChild, secondChild, err = expendDagToDiffChains(firstChild, secondChild, tc)
		if err != nil {
			t.Fatalf("createADagInBlockAtDepth: %+v", err)
		}
		checkExpectedBlockOnTwoDiffChain(firstChild, secondChild, 3, 2)

	})
}

func TestLowestChainBlockAboveOrEqualToBlueScore(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		params.FinalityDuration = 10 * params.TargetTimePerBlock
		factory := consensus.NewFactory()
		tc, tearDown, err := factory.NewTestConsensus(params, false,
			"TestLowestChainBlockAboveOrEqualToBlueScore")
		if err != nil {
			t.Fatalf("NewTestConsensus: %s", err)
		}
		defer tearDown(false)
		checkExpectedBlock := func(highHash *externalapi.DomainHash, blueScore uint64, expected *externalapi.DomainHash) {
			blockHash, err := tc.DAGTraversalManager().LowestChainBlockAboveOrEqualToBlueScore(highHash, blueScore)
			if err != nil {
				t.Fatalf("LowestChainBlockAboveOrEqualToBlueScore: %+v", err)
			}

			if !blockHash.Equal(expected) {
				t.Fatalf("Expected block %s but got %s", expected, blockHash)
			}
		}

		checkBlueScore := func(blockHash *externalapi.DomainHash, expectedBlueScoe uint64) {
			ghostdagData, err := tc.GHOSTDAGDataStore().Get(tc.DatabaseContext(), blockHash)
			if err != nil {
				t.Fatalf("GHOSTDAGDataStore().Get: %+v", err)
			}

			if ghostdagData.BlueScore() != expectedBlueScoe {
				t.Fatalf("Expected blue score %d but got %d", expectedBlueScoe, ghostdagData.BlueScore())
			}
		}

		chain := []*externalapi.DomainHash{params.GenesisHash}
		tipHash := params.GenesisHash
		for i := 0; i < 9; i++ {
			var err error
			tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			chain = append(chain, tipHash)
		}

		sideChain1TipHash, _, err := tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{sideChain1TipHash, tipHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		chain = append(chain, tipHash)
		blueScore11BlockHash := tipHash
		checkBlueScore(blueScore11BlockHash, 11)

		for i := 0; i < 5; i++ {
			var err error
			tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			chain = append(chain, tipHash)
		}

		sideChain2TipHash, _, err := tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{sideChain2TipHash, tipHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}
		chain = append(chain, tipHash)

		blueScore18BlockHash := tipHash
		checkBlueScore(blueScore18BlockHash, 18)

		for i := 0; i < 3; i++ {
			var err error
			tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			chain = append(chain, tipHash)
		}

		// Check by exact blue score
		checkExpectedBlock(tipHash, 0, params.GenesisHash)
		checkExpectedBlock(tipHash, 5, chain[5])
		checkExpectedBlock(tipHash, 19, chain[len(chain)-3])

		// Check by non exact blue score
		checkExpectedBlock(tipHash, 17, blueScore18BlockHash)
		checkExpectedBlock(tipHash, 10, blueScore11BlockHash)
	})
}

func createAChainDag(genesisHash *externalapi.DomainHash, tc testapi.TestConsensus) (*externalapi.DomainHash, error) {

	block1, _, err := tc.AddBlock([]*externalapi.DomainHash{genesisHash}, nil, nil)
	if err != nil {
		return nil, err
	}
	block2, _, err := tc.AddBlock([]*externalapi.DomainHash{block1}, nil, nil)
	if err != nil {
		return nil, err
	}
	block3, _, err := tc.AddBlock([]*externalapi.DomainHash{block2}, nil, nil)
	if err != nil {
		return nil, err
	}
	block4, _, err := tc.AddBlock([]*externalapi.DomainHash{block3}, nil, nil)
	if err != nil {
		return nil, err
	}
	block5, _, err := tc.AddBlock([]*externalapi.DomainHash{block4}, nil, nil)
	if err != nil {
		return nil, err
	}
	return block5, nil

}
func expendDagToTwoChildrenSameSelectedParent(parentHash *externalapi.DomainHash, tc testapi.TestConsensus) (*externalapi.DomainHash, *externalapi.DomainHash, error) {

	firstChild, _, err := tc.AddBlock([]*externalapi.DomainHash{parentHash}, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	secondChild, _, err := tc.AddBlock([]*externalapi.DomainHash{parentHash}, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	return firstChild, secondChild, nil

}

func expendDagToDiffChains(firstChainHash, secondChainHash *externalapi.DomainHash, tc testapi.TestConsensus) (*externalapi.DomainHash, *externalapi.DomainHash, error) {
	block := firstChainHash
	var err error
	for i := 0; i < 3; i++ {
		block, _, err = tc.AddBlock([]*externalapi.DomainHash{block}, nil, nil)
		if err != nil {
			return nil, nil, err
		}
	}
	firstChainHash = block

	block = secondChainHash
	for i := 0; i < 2; i++ {
		block, _, err = tc.AddBlock([]*externalapi.DomainHash{block}, nil, nil)
		if err != nil {
			return nil, nil, err
		}
	}
	secondChainHash = block
	return firstChainHash, secondChainHash, nil

}
