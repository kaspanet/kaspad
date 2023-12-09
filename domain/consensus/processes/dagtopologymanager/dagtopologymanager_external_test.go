package dagtopologymanager_test

import (
	"testing"

	"github.com/zoomy-network/zoomyd/domain/consensus/model"

	"github.com/zoomy-network/zoomyd/domain/consensus"
	"github.com/zoomy-network/zoomyd/domain/consensus/model/externalapi"
	"github.com/zoomy-network/zoomyd/domain/consensus/utils/testutils"
)

func TestIsAncestorOf(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()
		tc, tearDown, err := factory.NewTestConsensus(consensusConfig, "TestIsAncestorOf")
		if err != nil {
			t.Fatalf("NewTestConsensus: %s", err)
		}
		defer tearDown(false)

		// Add a chain of two blocks above the genesis. This will be the
		// selected parent chain.
		blockA, _, err := tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		blockB, _, err := tc.AddBlock([]*externalapi.DomainHash{blockA}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %s", err)
		}

		// Add another block above the genesis
		blockC, _, err := tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %s", err)
		}

		// Add a block whose parents are the two tips
		blockD, _, err := tc.AddBlock([]*externalapi.DomainHash{blockB, blockC}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %s", err)
		}

		isAncestorOf, err := tc.DAGTopologyManager().IsAncestorOf(model.NewStagingArea(), blockC, blockD)
		if err != nil {
			t.Fatalf("IsAncestorOf: %s", err)
		}
		if !isAncestorOf {
			t.Fatalf("TestIsInPast: node C is unexpectedly not the past of node D")
		}
	})
}
