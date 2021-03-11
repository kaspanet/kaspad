package reachabilitymanager_test

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"testing"
)

func TestReachabilityIsDAGAncestorOf(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {

		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(params, false, "TestReachabilityIsDAGAncestorOf")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		genesisHash := params.GenesisHash
		blockHashA, _, err := tc.AddBlock([]*externalapi.DomainHash{genesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %v", err)
		}

		blockHashB, _, err := tc.AddBlock([]*externalapi.DomainHash{blockHashA}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %v", err)
		}

		blockHashC, _, err := tc.AddBlock([]*externalapi.DomainHash{genesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %v", err)
		}

		blockHashD, _, err := tc.AddBlock([]*externalapi.DomainHash{blockHashA, blockHashC}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %v", err)
		}

		sharedBlockHash, _, err := tc.AddBlock([]*externalapi.DomainHash{blockHashB, blockHashD}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %v", err)
		}

		tests := []struct {
			firstBlockHash  *externalapi.DomainHash
			secondBlockHash *externalapi.DomainHash
			expectedResult  bool
		}{
			{
				firstBlockHash:  blockHashA,
				secondBlockHash: blockHashA,
				expectedResult:  true,
			},
			{
				firstBlockHash:  genesisHash,
				secondBlockHash: blockHashA,
				expectedResult:  true,
			},
			{
				firstBlockHash:  genesisHash,
				secondBlockHash: sharedBlockHash,
				expectedResult:  true,
			},
			{
				firstBlockHash:  blockHashC,
				secondBlockHash: blockHashD,
				expectedResult:  true,
			},
			{
				firstBlockHash:  blockHashA,
				secondBlockHash: blockHashD,
				expectedResult:  true,
			},
			{
				firstBlockHash:  blockHashC,
				secondBlockHash: blockHashB,
				expectedResult:  false,
			},
			{
				firstBlockHash:  blockHashB,
				secondBlockHash: blockHashD,
				expectedResult:  false,
			},
			{
				firstBlockHash:  blockHashB,
				secondBlockHash: blockHashA,
				expectedResult:  false,
			},
		}

		for _, test := range tests {
			isDAGAncestorOf, err := tc.ReachabilityManager().IsDAGAncestorOf(test.firstBlockHash, test.secondBlockHash)
			if err != nil {
				t.Fatalf("IsDAGAncestorOf: %v", err)
			}
			if isDAGAncestorOf != test.expectedResult {
				t.Fatalf("IsDAGAncestorOf: should returns %v but got %v", test.expectedResult, isDAGAncestorOf)
			}
		}
	})
}
