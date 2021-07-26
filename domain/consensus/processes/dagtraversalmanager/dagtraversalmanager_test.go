package dagtraversalmanager_test

import (
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/model"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
)

const commonChainSize = 5
const depth uint64 = 2

//TestBlockAtDepthOnChainDag compares the result of BlockAtDepth to the result of looping over the SelectedChain on a single chain DAG.
func TestBlockAtDepthOnChainDag(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		stagingArea := model.NewStagingArea()

		factory := consensus.NewFactory()
		tc, tearDown, err := factory.NewTestConsensus(consensusConfig,
			"TestBlockAtDepthOnChainDag")
		if err != nil {
			t.Fatalf("Failed creating a NewTestConsensus: %s", err)
		}
		defer tearDown(false)

		highHash, err := createAChainDAG(consensusConfig.GenesisHash, tc)
		if err != nil {
			t.Fatalf("Failed creating a Chain DAG In BlockAtDepthTEST: %+v", err)
		}
		currentBlockHash := highHash
		currentBlockData, err := tc.GHOSTDAGDataStore().Get(tc.DatabaseContext(), stagingArea, currentBlockHash, false)
		if err != nil {
			t.Fatalf("Failed getting GHOSTDAGData for block with hash %s: %+v", currentBlockHash.String(), err)
		}

		for i := uint64(0); i <= depth; i++ {
			if currentBlockData.SelectedParent() == nil {
				break
			}
			currentBlockHash = currentBlockData.SelectedParent()
			currentBlockData, err = tc.GHOSTDAGDataStore().Get(tc.DatabaseContext(), stagingArea, currentBlockHash, false)
			if err != nil {
				t.Fatalf("Failed getting GHOSTDAGData for block with hash %s: %+v", currentBlockHash.String(), err)
			}
		}
		expectedBlockHash := currentBlockHash
		actualBlockHash, err := tc.DAGTraversalManager().BlockAtDepth(stagingArea, highHash, depth)
		if err != nil {
			t.Fatalf("Failed on BlockAtDepth: %+v", err)
		}
		if !actualBlockHash.Equal(expectedBlockHash) {
			t.Fatalf("Expected block %s but got %s", expectedBlockHash, actualBlockHash)
		}
	})
}

func createAChainDAG(genesisHash *externalapi.DomainHash, tc testapi.TestConsensus) (*externalapi.DomainHash, error) {
	block := genesisHash
	var err error
	for i := 0; i < commonChainSize; i++ {
		block, _, err = tc.AddBlock([]*externalapi.DomainHash{block}, nil, nil)
		if err != nil {
			return nil, err
		}
	}
	return block, nil
}

// TestBlockAtDepthOnDAGWhereTwoBlocksHaveSameSelectedParent compares the results of BlockAtDepth
// of 2 children that have the same selectedParent.
func TestBlockAtDepthOnDAGWhereTwoBlocksHaveSameSelectedParent(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()
		tc, tearDown, err := factory.NewTestConsensus(consensusConfig,
			"TestBlockAtDepthOnDAGWhereTwoBlocksHaveSameSelectedParent")
		if err != nil {
			t.Fatalf("Failed creating a NewTestConsensus: %s", err)
		}
		defer tearDown(false)

		stagingArea := model.NewStagingArea()

		firstChild, secondChild, err := createADAGTwoChildrenWithSameSelectedParent(consensusConfig.GenesisHash, tc)
		if err != nil {
			t.Fatalf("Failed creating a DAG where two blocks have same selected parent: %+v", err)
		}
		actualBlockHash, err := tc.DAGTraversalManager().BlockAtDepth(stagingArea, firstChild, depth)
		if err != nil {
			t.Fatalf("Failed at BlockAtDepth: %+v", err)
		}
		expectedSameHash, err := tc.DAGTraversalManager().BlockAtDepth(stagingArea, secondChild, depth)
		if err != nil {
			t.Fatalf("Failed in BlockAtDepth: %+v", err)
		}
		if !actualBlockHash.Equal(expectedSameHash) {
			t.Fatalf("Expected block %s but got %s", expectedSameHash, actualBlockHash)
		}
	})
}

func createADAGTwoChildrenWithSameSelectedParent(genesisHash *externalapi.DomainHash,
	tc testapi.TestConsensus) (*externalapi.DomainHash, *externalapi.DomainHash, error) {

	block := genesisHash
	var err error
	for i := 0; i < commonChainSize; i++ {
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

// TestBlockAtDepthOnDAGWithTwoDifferentChains compares results of BlockAtDepth on two different chains,
// on the same DAG, and validates they merge at the correct point.
func TestBlockAtDepthOnDAGWithTwoDifferentChains(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()
		tc, tearDown, err := factory.NewTestConsensus(consensusConfig,
			"TestBlockAtDepthOnDAGWithTwoDifferentChains")
		if err != nil {
			t.Fatalf("Failed creating a NewTestConsensus: %s", err)
		}
		defer tearDown(false)

		const sizeOfTheFirstChildSubChainDAG = 3
		const sizeOfTheSecondChildSubChainDAG = 2

		firstChild, secondChild, err := createADAGWithTwoDifferentChains(consensusConfig.GenesisHash, tc, sizeOfTheFirstChildSubChainDAG,
			sizeOfTheSecondChildSubChainDAG)
		if err != nil {
			t.Fatalf("Failed creating a DAG with two different chains in BlockAtDepthTEST: %+v", err)
		}

		stagingArea := model.NewStagingArea()

		actualBlockHash, err := tc.DAGTraversalManager().BlockAtDepth(stagingArea, firstChild, sizeOfTheFirstChildSubChainDAG)
		if err != nil {
			t.Fatalf("Failed in BlockAtDepth: %+v", err)
		}
		expectedSameHash, err := tc.DAGTraversalManager().BlockAtDepth(stagingArea, secondChild, sizeOfTheSecondChildSubChainDAG)
		if err != nil {
			t.Fatalf("Failed in BlockAtDepth: %+v", err)
		}

		if !actualBlockHash.Equal(expectedSameHash) {
			t.Fatalf("Expected block %s but got %s", expectedSameHash, actualBlockHash)
		}
		expectedDiffHash, err := tc.DAGTraversalManager().BlockAtDepth(stagingArea, secondChild, sizeOfTheSecondChildSubChainDAG-1)
		if err != nil {
			t.Fatalf("Failed in BlockAtDepth: %+v", err)
		}
		if actualBlockHash.Equal(expectedDiffHash) {
			t.Fatalf("Expected to a differente block")
		}
	})
}

func createADAGWithTwoDifferentChains(genesisHash *externalapi.DomainHash, tc testapi.TestConsensus,
	sizeOfTheFirstChildSubChainDAG int, sizeOfTheSecondChildSubChainDAG int) (*externalapi.DomainHash, *externalapi.DomainHash, error) {

	block := genesisHash
	var err error
	for i := 0; i < commonChainSize; i++ {
		block, _, err = tc.AddBlock([]*externalapi.DomainHash{block}, nil, nil)
		if err != nil {
			return nil, nil, err
		}
	}
	firstChainTipHash, _, err := tc.AddBlock([]*externalapi.DomainHash{block}, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	secondChainTipHash, _, err := tc.AddBlock([]*externalapi.DomainHash{block}, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	for i := 0; i < sizeOfTheFirstChildSubChainDAG; i++ {
		firstChainTipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{firstChainTipHash}, nil, nil)
		if err != nil {
			return nil, nil, err
		}
	}

	for i := 0; i < sizeOfTheSecondChildSubChainDAG; i++ {
		secondChainTipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{secondChainTipHash}, nil, nil)
		if err != nil {
			return nil, nil, err
		}
	}
	return firstChainTipHash, secondChainTipHash, nil
}

func TestLowestChainBlockAboveOrEqualToBlueScore(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		consensusConfig.FinalityDuration = 10 * consensusConfig.TargetTimePerBlock
		factory := consensus.NewFactory()
		tc, tearDown, err := factory.NewTestConsensus(consensusConfig,
			"TestLowestChainBlockAboveOrEqualToBlueScore")
		if err != nil {
			t.Fatalf("NewTestConsensus: %s", err)
		}
		defer tearDown(false)

		stagingArea := model.NewStagingArea()

		checkExpectedBlock := func(highHash *externalapi.DomainHash, blueScore uint64, expected *externalapi.DomainHash) {
			blockHash, err := tc.DAGTraversalManager().LowestChainBlockAboveOrEqualToBlueScore(stagingArea, highHash, blueScore)
			if err != nil {
				t.Fatalf("LowestChainBlockAboveOrEqualToBlueScore: %+v", err)
			}

			if !blockHash.Equal(expected) {
				t.Fatalf("Expected block %s but got %s", expected, blockHash)
			}
		}

		checkBlueScore := func(blockHash *externalapi.DomainHash, expectedBlueScoe uint64) {
			ghostdagData, err := tc.GHOSTDAGDataStore().Get(tc.DatabaseContext(), stagingArea, blockHash, false)
			if err != nil {
				t.Fatalf("GHOSTDAGDataStore().Get: %+v", err)
			}

			if ghostdagData.BlueScore() != expectedBlueScoe {
				t.Fatalf("Expected blue score %d but got %d", expectedBlueScoe, ghostdagData.BlueScore())
			}
		}

		chain := []*externalapi.DomainHash{consensusConfig.GenesisHash}
		tipHash := consensusConfig.GenesisHash
		for i := 0; i < 9; i++ {
			var err error
			tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			chain = append(chain, tipHash)
		}

		sideChain1TipHash, _, err := tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
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

		sideChain2TipHash, _, err := tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
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
		checkExpectedBlock(tipHash, 0, consensusConfig.GenesisHash)
		checkExpectedBlock(tipHash, 5, chain[5])
		checkExpectedBlock(tipHash, 19, chain[len(chain)-3])

		// Check by non exact blue score
		checkExpectedBlock(tipHash, 17, blueScore18BlockHash)
		checkExpectedBlock(tipHash, 10, blueScore11BlockHash)
	})
}
