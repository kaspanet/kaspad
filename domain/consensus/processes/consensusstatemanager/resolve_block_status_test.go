package consensusstatemanager_test

import (
	"errors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/model"

	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"

	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
)

func TestDoubleSpends(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		stagingArea := model.NewStagingArea()

		consensusConfig.BlockCoinbaseMaturity = 0

		factory := consensus.NewFactory()

		consensus, teardown, err := factory.NewTestConsensus(consensusConfig, "TestUTXOCommitment")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		// Mine chain of two blocks to fund our double spend
		firstBlockHash, _, err := consensus.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error creating firstBlock: %+v", err)
		}
		fundingBlockHash, _, err := consensus.AddBlock([]*externalapi.DomainHash{firstBlockHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error creating fundingBlock: %+v", err)
		}
		fundingBlock, err := consensus.GetBlock(fundingBlockHash)
		if err != nil {
			t.Fatalf("Error getting fundingBlock: %+v", err)
		}

		// Get funding transaction
		fundingTransaction := fundingBlock.Transactions[transactionhelper.CoinbaseTransactionIndex]

		// Create two transactions that spends the same output, but with different IDs
		spendingTransaction1, err := testutils.CreateTransaction(fundingTransaction, 1)
		if err != nil {
			t.Fatalf("Error creating spendingTransaction1: %+v", err)
		}
		spendingTransaction2, err := testutils.CreateTransaction(fundingTransaction, 1)
		if err != nil {
			t.Fatalf("Error creating spendingTransaction2: %+v", err)
		}
		spendingTransaction2.Outputs[0].Value-- // tweak the value to create a different ID
		spendingTransaction1ID := consensushashing.TransactionID(spendingTransaction1)
		spendingTransaction2ID := consensushashing.TransactionID(spendingTransaction2)
		if spendingTransaction1ID.Equal(spendingTransaction2ID) {
			t.Fatalf("spendingTransaction1 and spendingTransaction2 ids are equal")
		}

		// Mine a block with spendingTransaction1 and make sure that it's valid
		goodBlock1Hash, _, err := consensus.AddBlock([]*externalapi.DomainHash{fundingBlockHash}, nil,
			[]*externalapi.DomainTransaction{spendingTransaction1})
		if err != nil {
			t.Fatalf("Error adding goodBlock1: %+v", err)
		}
		goodBlock1Status, err := consensus.BlockStatusStore().Get(consensus.DatabaseContext(), stagingArea, goodBlock1Hash)
		if err != nil {
			t.Fatalf("Error getting status of goodBlock1: %+v", err)
		}
		if goodBlock1Status != externalapi.StatusUTXOValid {
			t.Fatalf("GoodBlock1 status expected to be '%s', but is '%s'", externalapi.StatusUTXOValid, goodBlock1Status)
		}

		// To check that a block containing the same transaction already in it's past is disqualified:
		// Add a block on top of goodBlock, containing spendingTransaction1, and make sure it's disqualified
		doubleSpendingBlock1Hash, _, err := consensus.AddBlock([]*externalapi.DomainHash{goodBlock1Hash}, nil,
			[]*externalapi.DomainTransaction{spendingTransaction1})
		if err != nil {
			t.Fatalf("Error adding doubleSpendingBlock1: %+v", err)
		}
		doubleSpendingBlock1Status, err := consensus.BlockStatusStore().Get(consensus.DatabaseContext(), stagingArea, doubleSpendingBlock1Hash)
		if err != nil {
			t.Fatalf("Error getting status of goodBlock: %+v", err)
		}
		if doubleSpendingBlock1Status != externalapi.StatusDisqualifiedFromChain {
			t.Fatalf("doubleSpendingBlock1 status expected to be '%s', but is '%s'",
				externalapi.StatusDisqualifiedFromChain, doubleSpendingBlock1Status)
		}

		// To check that a block containing a transaction that double-spends a transaction that
		// is in it's past is disqualified:
		// Add a block on top of goodBlock, containing spendingTransaction2, and make sure it's disqualified
		doubleSpendingBlock2Hash, _, err := consensus.AddBlock([]*externalapi.DomainHash{goodBlock1Hash}, nil,
			[]*externalapi.DomainTransaction{spendingTransaction2})
		if err != nil {
			t.Fatalf("Error adding doubleSpendingBlock2: %+v", err)
		}
		doubleSpendingBlock2Status, err := consensus.BlockStatusStore().Get(consensus.DatabaseContext(), stagingArea, doubleSpendingBlock2Hash)
		if err != nil {
			t.Fatalf("Error getting status of goodBlock: %+v", err)
		}
		if doubleSpendingBlock2Status != externalapi.StatusDisqualifiedFromChain {
			t.Fatalf("doubleSpendingBlock2 status expected to be '%s', but is '%s'",
				externalapi.StatusDisqualifiedFromChain, doubleSpendingBlock2Status)
		}

		// To make sure that a block double-spending itself is rejected:
		// Add a block on top of goodBlock, containing both spendingTransaction1 and spendingTransaction2, and make
		// sure AddBlock returns a RuleError
		_, _, err = consensus.AddBlock([]*externalapi.DomainHash{goodBlock1Hash}, nil,
			[]*externalapi.DomainTransaction{spendingTransaction1, spendingTransaction2})
		if err == nil {
			t.Fatalf("No error when adding a self-double-spending block")
		}
		if !errors.Is(err, ruleerrors.ErrDoubleSpendInSameBlock) {
			t.Fatalf("Adding self-double-spending block should have "+
				"returned ruleerrors.ErrDoubleSpendInSameBlock, but instead got: %+v", err)
		}

		// To make sure that a block containing the same transaction twice is rejected:
		// Add a block on top of goodBlock, containing spendingTransaction1 twice, and make
		// sure AddBlock returns a RuleError
		_, _, err = consensus.AddBlock([]*externalapi.DomainHash{goodBlock1Hash}, nil,
			[]*externalapi.DomainTransaction{spendingTransaction1, spendingTransaction1})
		if err == nil {
			t.Fatalf("No error when adding a block containing the same transactin twice")
		}
		if !errors.Is(err, ruleerrors.ErrDuplicateTx) {
			t.Fatalf("Adding block that contains the same transaction twice should have "+
				"returned ruleerrors.ErrDuplicateTx, but instead got: %+v", err)
		}

		// Check that a block will not get disqualified if it has a transaction that double spends
		// a transaction from its anticone.
		goodBlock2Hash, _, err := consensus.AddBlock([]*externalapi.DomainHash{fundingBlockHash}, nil,
			[]*externalapi.DomainTransaction{spendingTransaction2})
		if err != nil {
			t.Fatalf("Error adding goodBlock: %+v", err)
		}
		//use ResolveBlockStatus, since goodBlock2 might not be the selectedTip
		goodBlock2Status, err := consensus.ConsensusStateManager().ResolveBlockStatus(
			stagingArea, goodBlock2Hash, true)
		if err != nil {
			t.Fatalf("Error getting status of goodBlock: %+v", err)
		}
		if goodBlock2Status != externalapi.StatusUTXOValid {
			t.Fatalf("GoodBlock2 status expected to be '%s', but is '%s'", externalapi.StatusUTXOValid, goodBlock2Status)
		}
	})
}

// TestTransactionAcceptance checks that block transactions are accepted correctly when the merge set is sorted topologically.
// DAG diagram:
// genesis <- blockA <- blockB <- blockC   <- ..(chain of k-blocks).. lastBlockInChain <- blockD <- blockE <- blockF <- blockG
//                                ^								           ^									          |
//								  | redBlock <------------------------ blueChildOfRedBlock <-------------------------------
func TestTransactionAcceptance(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		stagingArea := model.NewStagingArea()
		consensusConfig.BlockCoinbaseMaturity = 0
		factory := consensus.NewFactory()
		testConsensus, teardown, err := factory.NewTestConsensus(consensusConfig, "TestTransactionAcceptance")
		if err != nil {
			t.Fatalf("Error setting up testConsensus: %+v", err)
		}
		defer teardown(false)

		blockHashA, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{consensusConfig.GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error creating blockA: %+v", err)
		}
		blockHashB, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{blockHashA}, nil, nil)
		if err != nil {
			t.Fatalf("Error creating blockB: %+v", err)
		}
		blockHashC, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{blockHashB}, nil, nil)
		if err != nil {
			t.Fatalf("Error creating blockC: %+v", err)
		}
		// Add a chain of K blocks above blockC so we'll
		// be able to mine a red block on top of it.
		chainTipHash := blockHashC
		for i := externalapi.KType(0); i < consensusConfig.K; i++ {
			var err error
			chainTipHash, _, err = testConsensus.AddBlock([]*externalapi.DomainHash{chainTipHash}, nil, nil)
			if err != nil {
				t.Fatalf("Error creating a block: %+v", err)
			}
		}
		lastBlockInChain := chainTipHash
		blockC, err := testConsensus.GetBlock(blockHashC)
		if err != nil {
			t.Fatalf("Error getting blockC: %+v", err)
		}
		fees := uint64(1)
		transactionFromBlockC := blockC.Transactions[transactionhelper.CoinbaseTransactionIndex]
		// transactionFromRedBlock is spending TransactionFromBlockC.
		transactionFromRedBlock, err := testutils.CreateTransaction(transactionFromBlockC, fees)
		if err != nil {
			t.Fatalf("Error creating a transactionFromRedBlock: %+v", err)
		}
		transactionFromRedBlockInput0UTXOEntry, err := testConsensus.ConsensusStateStore().
			UTXOByOutpoint(testConsensus.DatabaseContext(), stagingArea, &transactionFromRedBlock.Inputs[0].PreviousOutpoint)
		if err != nil {
			t.Fatalf("Error getting UTXOEntry for transactionFromRedBlockInput: %s", err)
		}
		redHash, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{blockHashC}, nil,
			[]*externalapi.DomainTransaction{transactionFromRedBlock})
		if err != nil {
			t.Fatalf("Error creating redBlock: %+v", err)
		}

		transactionFromBlueChildOfRedBlock, err := testutils.CreateTransaction(transactionFromRedBlock, fees)
		if err != nil {
			t.Fatalf("Error creating transactionFromBlueChildOfRedBlock: %+v", err)
		}
		transactionFromBlueChildOfRedBlockInput0UTXOEntry, err := testConsensus.ConsensusStateStore().
			UTXOByOutpoint(testConsensus.DatabaseContext(), stagingArea, &transactionFromBlueChildOfRedBlock.Inputs[0].PreviousOutpoint)
		if err != nil {
			t.Fatalf("Error getting UTXOEntry for transactionFromBlueChildOfRedBlockInput: %s", err)
		}
		blueChildOfRedBlockScriptPublicKey := &externalapi.ScriptPublicKey{Script: []byte{3}, Version: 0}
		// The blueChildOfRedBlock contains a transaction that spent an output from the red block.
		hashBlueChildOfRedBlock, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{lastBlockInChain, redHash},
			&externalapi.DomainCoinbaseData{
				ScriptPublicKey: blueChildOfRedBlockScriptPublicKey,
				ExtraData:       nil,
			}, []*externalapi.DomainTransaction{transactionFromBlueChildOfRedBlock})
		if err != nil {
			t.Fatalf("Error creating blueChildOfRedBlock: %+v", err)
		}

		// K blocks minded between blockC and blockD.
		blockHashD, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{lastBlockInChain}, nil, nil)
		if err != nil {
			t.Fatalf("Error creating blockD : %+v", err)
		}
		blockHashE, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{blockHashD}, nil, nil)
		if err != nil {
			t.Fatalf("Error creating blockD : %+v", err)
		}
		blockEScriptPublicKey := &externalapi.ScriptPublicKey{Script: []byte{4}, Version: 0}
		blockHashF, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{blockHashE},
			&externalapi.DomainCoinbaseData{
				ScriptPublicKey: blockEScriptPublicKey,
				ExtraData:       nil,
			}, nil)
		if err != nil {
			t.Fatalf("Error creating blockE: %+v", err)
		}
		blockFScriptPublicKey := &externalapi.ScriptPublicKey{Script: []byte{5}, Version: 0}
		blockHashG, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{blockHashF, hashBlueChildOfRedBlock},
			&externalapi.DomainCoinbaseData{
				ScriptPublicKey: blockFScriptPublicKey,
				ExtraData:       nil,
			}, nil)
		if err != nil {
			t.Fatalf("Error creating blockF: %+v", err)
		}

		acceptanceData, err := testConsensus.AcceptanceDataStore().Get(testConsensus.DatabaseContext(), stagingArea, blockHashG)
		if err != nil {
			t.Fatalf("Error getting acceptance data: %+v", err)
		}
		blueChildOfRedBlock, err := testConsensus.GetBlock(hashBlueChildOfRedBlock)
		if err != nil {
			t.Fatalf("Error getting blueChildOfRedBlock: %+v", err)
		}
		blockE, err := testConsensus.GetBlock(blockHashF)
		if err != nil {
			t.Fatalf("Error getting blockE: %+v", err)
		}
		redBlock, err := testConsensus.GetBlock(redHash)
		if err != nil {
			t.Fatalf("Error getting redBlock: %+v", err)
		}
		_, err = testConsensus.GetBlock(blockHashG)
		if err != nil {
			t.Fatalf("Error getting blockF: %+v", err)
		}
		updatedDAAScoreVirtualBlock := consensusConfig.GenesisBlock.Header.DAAScore() + 26
		//We expect the second transaction in the "blue block" (blueChildOfRedBlock) to be accepted because the merge set is ordered topologically
		//and the red block is ordered topologically before the "blue block" so the input is known in the UTXOSet.
		expectedAcceptanceData := externalapi.AcceptanceData{
			{
				BlockHash: blockHashF,
				TransactionAcceptanceData: []*externalapi.TransactionAcceptanceData{
					{
						Transaction:                 blockE.Transactions[0],
						Fee:                         0,
						IsAccepted:                  true,
						TransactionInputUTXOEntries: []externalapi.UTXOEntry{},
					},
				},
			},
			{
				BlockHash: redHash,
				TransactionAcceptanceData: []*externalapi.TransactionAcceptanceData{
					{ //Coinbase transaction outputs are added to the UTXO-set only if they are in the selected parent chain,
						// and this block isn't.
						Transaction:                 redBlock.Transactions[0],
						Fee:                         0,
						IsAccepted:                  false,
						TransactionInputUTXOEntries: []externalapi.UTXOEntry{},
					},
					{
						Transaction:                 redBlock.Transactions[1],
						Fee:                         fees,
						IsAccepted:                  true,
						TransactionInputUTXOEntries: []externalapi.UTXOEntry{transactionFromRedBlockInput0UTXOEntry},
					},
				},
			},
			{
				BlockHash: hashBlueChildOfRedBlock,
				TransactionAcceptanceData: []*externalapi.TransactionAcceptanceData{
					{ //Coinbase transaction outputs are added to the UTXO-set only if they are in the selected parent chain,
						// and this block isn't.
						Transaction:                 blueChildOfRedBlock.Transactions[0],
						Fee:                         0,
						IsAccepted:                  false,
						TransactionInputUTXOEntries: []externalapi.UTXOEntry{},
					},
					{ // The DAAScore was calculated by the virtual block pov. The DAAScore has changed since more blocks were added to the DAG.
						// So we will change the DAAScore in the UTXOEntryInput to the updated virtual DAAScore.
						Transaction: blueChildOfRedBlock.Transactions[1],
						Fee:         fees,
						IsAccepted:  true,
						TransactionInputUTXOEntries: []externalapi.UTXOEntry{
							utxo.NewUTXOEntry(transactionFromBlueChildOfRedBlockInput0UTXOEntry.Amount(),
								transactionFromBlueChildOfRedBlockInput0UTXOEntry.ScriptPublicKey(),
								transactionFromBlueChildOfRedBlockInput0UTXOEntry.IsCoinbase(), uint64(updatedDAAScoreVirtualBlock))},
					},
				},
			},
		}
		if !acceptanceData.Equal(expectedAcceptanceData) {
			t.Fatalf("The acceptance data is not the expected acceptance data")
		}
	})
}

func TestResolveBlockStatusSanity(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		stagingArea := model.NewStagingArea()

		consensus, teardown, err := consensus.NewFactory().NewTestConsensus(consensusConfig, "TestResolveBlockStatusSanity")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		genesisHash := consensusConfig.GenesisHash
		allHashes := []*externalapi.DomainHash{genesisHash}

		// Make sure that the status of genesisHash is valid
		genesisStatus, err := consensus.BlockStatusStore().Get(consensus.DatabaseContext(), stagingArea, genesisHash)
		if err != nil {
			t.Fatalf("error getting genesis status: %s", err)
		}
		if genesisStatus != externalapi.StatusUTXOValid {
			t.Fatalf("genesis is unexpectedly non-valid. Its status is: %s", genesisStatus)
		}

		chainLength := int(consensusConfig.K) + 1

		// Add a chain of blocks over the genesis and make sure all their
		// statuses are valid
		currentHash := genesisHash
		for i := 0; i < chainLength; i++ {
			addedBlockHash, _, err := consensus.AddBlock([]*externalapi.DomainHash{currentHash}, nil, nil)
			if err != nil {
				t.Fatalf("error adding block %d: %s", i, err)
			}
			blockStatus, err := consensus.BlockStatusStore().Get(consensus.DatabaseContext(), stagingArea, addedBlockHash)
			if err != nil {
				t.Fatalf("error getting block %d (%s) status: %s", i, addedBlockHash, err)
			}
			if blockStatus != externalapi.StatusUTXOValid {
				t.Fatalf("block %d (%s) is unexpectedly non-valid. Its status is: %s", i, addedBlockHash, blockStatus)
			}
			currentHash = addedBlockHash
			allHashes = append(allHashes, addedBlockHash)
		}

		// Add another chain of blocks over the genesis that's shorter than
		// the original chain by 1. Here we expect all the statuses to be
		// StatusUTXOPendingVerification
		currentHash = genesisHash
		for i := 0; i < chainLength-1; i++ {
			addedBlockHash, _, err := consensus.AddBlock([]*externalapi.DomainHash{currentHash}, nil, nil)
			if err != nil {
				t.Fatalf("error adding block %d: %s", i, err)
			}
			blockStatus, err := consensus.BlockStatusStore().Get(consensus.DatabaseContext(), stagingArea, addedBlockHash)
			if err != nil {
				t.Fatalf("error getting block %d (%s) status: %s", i, addedBlockHash, err)
			}
			if blockStatus != externalapi.StatusUTXOPendingVerification {
				t.Fatalf("block %d (%s) has unexpected status. "+
					"Want: %s, got: %s", i, addedBlockHash, externalapi.StatusUTXOPendingVerification, blockStatus)
			}
			currentHash = addedBlockHash
			allHashes = append(allHashes, addedBlockHash)
		}

		// Add another two blocks to the second chain. This should trigger
		// resolving the entire chain
		for i := 0; i < 2; i++ {
			addedBlockHash, _, err := consensus.AddBlock([]*externalapi.DomainHash{currentHash}, nil, nil)
			if err != nil {
				t.Fatalf("error adding block %d: %s", i, err)
			}
			currentHash = addedBlockHash
			allHashes = append(allHashes, addedBlockHash)
		}

		// Make sure that all the blocks in the DAG now have StatusUTXOValid
		for _, hash := range allHashes {
			blockStatus, err := consensus.BlockStatusStore().Get(consensus.DatabaseContext(), stagingArea, hash)
			if err != nil {
				t.Fatalf("error getting block %s status: %s", hash, err)
			}
			if blockStatus != externalapi.StatusUTXOValid {
				t.Fatalf("block %s is unexpectedly non-valid. Its status is: %s", hash, blockStatus)
			}
		}
	})
}
