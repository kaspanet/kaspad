package blockvalidator_test

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/pkg/errors"
	"testing"
)

func TestChainedTransactions(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		params.BlockCoinbaseMaturity = 0

		factory := consensus.NewFactory()

		tc, teardown, err := factory.NewTestConsensus(params, "TestUTXOCommitment")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown()

		block1Hash, err := tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("AddBlock: %+v", err)
		}

		block1, err := tc.GetBlock(block1Hash)
		if err != nil {
			t.Fatalf("Error getting block1: %+v", err)
		}

		tx1, err := testutils.CreateTransaction(block1.Transactions[0])
		if err != nil {
			t.Fatalf("Error creating spendingTransaction1: %+v", err)
		}

		chainedTx, err := testutils.CreateTransaction(tx1)
		if err != nil {
			t.Fatalf("Error creating chainedTx: %+v", err)
		}

		// Check that a block is invalid if it contains chained transactions
		_, err = tc.AddBlock([]*externalapi.DomainHash{block1Hash}, nil,
			[]*externalapi.DomainTransaction{tx1, chainedTx})
		if !errors.Is(err, ruleerrors.ErrChainedTransactions) {
			t.Fatalf("unexpected error %+v", err)
		}

		block2Hash, err := tc.AddBlock([]*externalapi.DomainHash{block1Hash}, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error %+v", err)
		}

		block2, err := tc.GetBlock(block2Hash)
		if err != nil {
			t.Fatalf("Error getting block2: %+v", err)
		}

		tx2, err := testutils.CreateTransaction(block2.Transactions[0])
		if err != nil {
			t.Fatalf("Error creating spendingTransaction1: %+v", err)
		}

		// Check that a block is valid if it contains two non chained transactions
		_, err = tc.AddBlock([]*externalapi.DomainHash{block2Hash}, nil,
			[]*externalapi.DomainTransaction{tx1, tx2})
		if err != nil {
			t.Fatalf("unexpected error %+v", err)
		}
	})
}
