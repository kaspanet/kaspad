package consensusstatemanager_test

import (
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/dagconfig"
)

func TestVirtualDiff(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(params, false, "TestVirtualDiff")
		if err != nil {
			t.Fatalf("Error setting up tc: %+v", err)
		}
		defer teardown(false)

		// Add block A over the genesis
		blockHash, blockInsertionResult, err := tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error adding block A: %+v", err)
		}

		block, err := tc.BlockStore().Block(tc.DatabaseContext(), blockHash)
		if err != nil {
			t.Fatalf("Block: %+v", err)
		}

		virtualUTXODiff := blockInsertionResult.VirtualUTXODiff
		if virtualUTXODiff.ToRemove().Len() != 0 {
			t.Fatalf("Unexpected length %d for virtualUTXODiff.ToRemove()", virtualUTXODiff.ToRemove().Len())
		}

		if virtualUTXODiff.ToAdd().Len() != 1 {
			t.Fatalf("Unexpected length %d for virtualUTXODiff.ToAdd()", virtualUTXODiff.ToAdd().Len())
		}

		iterator := virtualUTXODiff.ToAdd().Iterator()
		iterator.First()

		outpoint, entry, err := iterator.Get()
		if err != nil {
			t.Fatalf("TestVirtualDiff: %+v", err)
		}

		if !outpoint.Equal(&externalapi.DomainOutpoint{
			TransactionID: *consensushashing.TransactionID(block.Transactions[0]),
			Index:         0,
		}) {
			t.Fatalf("Unexpected outpoint %s", outpoint)
		}

		if !entry.Equal(utxo.NewUTXOEntry(
			block.Transactions[0].Outputs[0].Value,
			block.Transactions[0].Outputs[0].ScriptPublicKey,
			true,
			2, //Expected virtual blue score
		)) {
			t.Fatalf("Unexpected entry %s", entry)
		}
	})
}
