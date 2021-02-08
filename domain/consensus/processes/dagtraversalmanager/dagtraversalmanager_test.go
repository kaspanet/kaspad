package dagtraversalmanager_test

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"testing"
)

const sizeOfTheChainThatCommonToAllGraphs int = 5

func TestBlockAtDepth(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {

		factory := consensus.NewFactory()
		tc, tearDown, err := factory.NewTestConsensus(params, false,
			"TestBlockAtDepth")
		if err != nil {
			t.Fatalf("Failed creating a NewTestConsensus: %s", err)
		}
		defer tearDown(false)

		// The first test compares the result of BlockAtDepth to the result of looping over the SelectedChain on a single chain DAG.
		const depth uint64 = 2
		highHash, err := createAChainDAG(params.GenesisHash, tc)
		if err != nil {
			t.Fatalf("Failed creating a Chain DAG In BlockAtDepthTEST: %+v", err)
		}
		currentBlockHash := highHash
		currentBlockData, err := tc.GHOSTDAGDataStore().Get(tc.DatabaseContext(), currentBlockHash)
		if err != nil {
			t.Fatalf("Failed getting GHOSTDAGData for block with hash %s: %+v", currentBlockHash.String(), err)
		}

		for i := uint64(0); i <= depth; i++ {
			if currentBlockData.SelectedParent() == nil {
				break
			}
			currentBlockHash = currentBlockData.SelectedParent()
			currentBlockData, err = tc.GHOSTDAGDataStore().Get(tc.DatabaseContext(), currentBlockHash)
			if err != nil {
				t.Fatalf("Failed getting GHOSTDAGData for block with hash %s: %+v", currentBlockHash.String(), err)
			}
		}
		expectedBlockHash := currentBlockHash
		actualBlockHash, err := tc.DAGTraversalManager().BlockAtDepth(highHash, depth)
		if err != nil {
			t.Fatalf("Failed on BlockAtDepth: %+v", err)
		}
		if !actualBlockHash.Equal(expectedBlockHash) {
			t.Fatalf("Expected block %s but got %s", expectedBlockHash, actualBlockHash)
		}

		// The second test compares the results of BlockAtDepth of 2 children that have the same selectedParent.
		tc, tearDown, err = factory.NewTestConsensus(params, false,
			"TestBlockAtDepth")
		if err != nil {
			t.Fatalf("Failed creating a NewTestConsensus: %s", err)
		}
		firstChild, secondChild, err := createADAGTwoChildrenWithSameSelectedParent(params.GenesisHash, tc)
		if err != nil {
			t.Fatalf("Failed creating a DAG InBlockAtDepthTEST: %+v", err)
		}
		actualBlockHash, err = tc.DAGTraversalManager().BlockAtDepth(firstChild, depth)
		if err != nil {
			t.Fatalf("Failed at BlockAtDepth: %+v", err)
		}
		expectedSameHash, err := tc.DAGTraversalManager().BlockAtDepth(secondChild, depth)
		if err != nil {
			t.Fatalf("Failed in BlockAtDepth: %+v", err)
		}
		if !actualBlockHash.Equal(expectedSameHash) {
			t.Fatalf("Expected block %s but got %s", expectedSameHash, actualBlockHash)
		}

		// The third test compares results of BlockAtDepth on two different chains, on the same DAG, and validates they merge at the correct point.
		const firstDepth uint64 = 3
		const secondDepth uint64 = 2
		tc, tearDown, err = factory.NewTestConsensus(params, false,
			"TestBlockAtDepth")
		if err != nil {
			t.Fatalf("Failed creating a NewTestConsensus: %s", err)
		}
		firstChild, secondChild, err = createADAGWithTwoDifferentChains(params.GenesisHash, tc)
		if err != nil {
			t.Fatalf("Failed creating a DAG in lockAtDepthTEST: %+v", err)
		}

		actualBlockHash, err = tc.DAGTraversalManager().BlockAtDepth(firstChild, firstDepth)
		if err != nil {
			t.Fatalf("Failed in BlockAtDepth: %+v", err)
		}
		expectedSameHash, err = tc.DAGTraversalManager().BlockAtDepth(secondChild, secondDepth)
		if err != nil {
			t.Fatalf("Failed in BlockAtDepth: %+v", err)
		}

		if !actualBlockHash.Equal(expectedSameHash) {
			t.Fatalf("Expected block %s but got %s", expectedSameHash, actualBlockHash)
		}
		expectedDiffHash, err := tc.DAGTraversalManager().BlockAtDepth(secondChild, secondDepth-1)
		if err != nil {
			t.Fatalf("Failed in BlockAtDepth: %+v", err)
		}
		if actualBlockHash.Equal(expectedDiffHash) {
			t.Fatalf("Expected to a differente block")
		}
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

func createAChainDAG(genesisHash *externalapi.DomainHash, tc testapi.TestConsensus) (*externalapi.DomainHash, error) {
	block := genesisHash
	var err error
	for i := 0; i < sizeOfTheChainThatCommonToAllGraphs; i++ {
		block, _, err = tc.AddBlock([]*externalapi.DomainHash{block}, nil, nil)
		if err != nil {
			return nil, err
		}
	}
	return block, nil
}

func createADAGTwoChildrenWithSameSelectedParent(genesisHash *externalapi.DomainHash, tc testapi.TestConsensus) (*externalapi.DomainHash, *externalapi.DomainHash, error) {
	block := genesisHash
	var err error
	for i := 0; i < sizeOfTheChainThatCommonToAllGraphs; i++ {
		block, _, err = tc.AddBlock([]*externalapi.DomainHash{block}, nil, nil)
		if err != nil {
			return nil, nil, err
		}
	}
	firstChild, _, err := tc.AddBlock([]*externalapi.DomainHash{block}, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	secondChild, _, err := tc.AddBlock([]*externalapi.DomainHash{block}, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	return firstChild, secondChild, nil
}

func createADAGWithTwoDifferentChains(genesisHash *externalapi.DomainHash, tc testapi.TestConsensus) (*externalapi.DomainHash, *externalapi.DomainHash, error) {
	const sizeOfTheFirstChildSubChainDAG int = 3
	const sizeOfTheSecondChildSubChainDAG int = 2
	block := genesisHash
	var err error
	for i := 0; i < sizeOfTheChainThatCommonToAllGraphs; i++ {
		block, _, err = tc.AddBlock([]*externalapi.DomainHash{block}, nil, nil)
		if err != nil {
			return nil, nil, err
		}
	}
	firstChainHash, _, err := tc.AddBlock([]*externalapi.DomainHash{block}, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	secondChainHash, _, err := tc.AddBlock([]*externalapi.DomainHash{block}, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	block = firstChainHash
	for i := 0; i < sizeOfTheFirstChildSubChainDAG; i++ {
		block, _, err = tc.AddBlock([]*externalapi.DomainHash{block}, nil, nil)
		if err != nil {
			return nil, nil, err
		}
	}
	firstChainHash = block

	block = secondChainHash
	for i := 0; i < sizeOfTheSecondChildSubChainDAG; i++ {
		block, _, err = tc.AddBlock([]*externalapi.DomainHash{block}, nil, nil)
		if err != nil {
			return nil, nil, err
		}
	}
	secondChainHash = block
	return firstChainHash, secondChainHash, nil
}
