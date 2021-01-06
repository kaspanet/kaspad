package blockprocessor_test

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/blockheader"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/pkg/errors"
	"testing"
)

func TestBlockStatus(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(params, "TestBlockStatus")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		checkStatus := func(hash *externalapi.DomainHash, expectedStatus externalapi.BlockStatus) {
			blockStatus, err := tc.BlockStatusStore().Get(tc.DatabaseContext(), hash)
			if err != nil {
				t.Fatalf("BlockStatusStore().Get: %+v", err)
			}

			if blockStatus != expectedStatus {
				t.Fatalf("Expected to have status %s but got %s", expectedStatus, blockStatus)
			}
		}

		tipHash := params.GenesisHash
		for i := 0; i < 2; i++ {
			tipHash, _, err = tc.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("AddBlock: %+v", err)
			}

			checkStatus(tipHash, externalapi.StatusUTXOValid)
		}

		headerHash, _, err := tc.AddHeader([]*externalapi.DomainHash{tipHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		checkStatus(headerHash, externalapi.StatusHeaderOnly)

		nonChainBlockHash, _, err := tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		checkStatus(nonChainBlockHash, externalapi.StatusUTXOPendingVerification)

		disqualifiedBlock, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{tipHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}
		disqualifiedBlock.Header = blockheader.NewImmutableBlockHeader(
			disqualifiedBlock.Header.Version(),
			disqualifiedBlock.Header.ParentHashes(),
			disqualifiedBlock.Header.HashMerkleRoot(),
			externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{}), // This should disqualify the block
			disqualifiedBlock.Header.UTXOCommitment(),
			disqualifiedBlock.Header.TimeInMilliseconds(),
			disqualifiedBlock.Header.Bits(),
			disqualifiedBlock.Header.Nonce(),
		)

		_, err = tc.ValidateAndInsertBlock(disqualifiedBlock)
		if err != nil {
			t.Fatalf("ValidateAndInsertBlock: %+v", err)
		}

		checkStatus(consensushashing.BlockHash(disqualifiedBlock), externalapi.StatusDisqualifiedFromChain)

		invalidBlock, _, err := tc.BuildBlockWithParents([]*externalapi.DomainHash{tipHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}
		invalidBlock.Header = blockheader.NewImmutableBlockHeader(
			disqualifiedBlock.Header.Version(),
			disqualifiedBlock.Header.ParentHashes(),
			externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{}), // This should invalidate the block
			disqualifiedBlock.Header.AcceptedIDMerkleRoot(),
			disqualifiedBlock.Header.UTXOCommitment(),
			disqualifiedBlock.Header.TimeInMilliseconds(),
			disqualifiedBlock.Header.Bits(),
			disqualifiedBlock.Header.Nonce(),
		)

		_, err = tc.ValidateAndInsertBlock(invalidBlock)
		if err == nil {
			t.Fatalf("block is expected to be invalid")
		}
		if !errors.As(err, &ruleerrors.RuleError{}) {
			t.Fatalf("ValidateAndInsertBlock: %+v", err)
		}

		checkStatus(consensushashing.BlockHash(invalidBlock), externalapi.StatusInvalid)
	})
}
