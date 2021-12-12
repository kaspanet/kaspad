package consensusstatemanager_test

import (
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
)

func TestVirtualDiff(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestVirtualDiff")
		if err != nil {
			t.Fatalf("Error setting up tc: %+v", err)
		}
		defer teardown(false)

		// Add block A over the genesis
		blockAHash, virtualChangeSet, err := tc.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error adding block A: %+v", err)
		}

		virtualUTXODiff := virtualChangeSet.VirtualUTXODiff
		if virtualUTXODiff.ToRemove().Len() != 0 {
			t.Fatalf("Unexpected length %d for virtualUTXODiff.ToRemove()", virtualUTXODiff.ToRemove().Len())
		}

		// Because the genesis is not in block A's DAA window, block A's coinbase doesn't pay to it, so it has no outputs.
		if virtualUTXODiff.ToAdd().Len() != 0 {
			t.Fatalf("Unexpected length %d for virtualUTXODiff.ToAdd()", virtualUTXODiff.ToAdd().Len())
		}

		blockBHash, virtualChangeSet, err := tc.AddBlock([]*externalapi.DomainHash{blockAHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error adding block A: %+v", err)
		}

		blockB, err := tc.BlockStore().Block(tc.DatabaseContext(), model.NewStagingArea(), blockBHash)
		if err != nil {
			t.Fatalf("Block: %+v", err)
		}

		virtualUTXODiff = virtualChangeSet.VirtualUTXODiff
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
			TransactionID: *consensushashing.TransactionID(blockB.Transactions[0]),
			Index:         0,
		}) {
			t.Fatalf("Unexpected outpoint %s", outpoint)
		}

		if !entry.Equal(utxo.NewUTXOEntry(
			blockB.Transactions[0].Outputs[0].Value,
			blockB.Transactions[0].Outputs[0].ScriptPublicKey,
			true,
			consensusConfig.GenesisBlock.Header.DAAScore()+2, //Expected virtual DAA score
		)) {
			t.Fatalf("Unexpected entry %s", entry)
		}
	})
}
