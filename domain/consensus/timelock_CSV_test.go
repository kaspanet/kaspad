package consensus_test

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"testing"
)

// TestCheckSequenceVerifyConditionedByBlockHeight verifies that an locked output (by CSC script) is spendable only after
// a certain number of blocks have added relative to the time the UTXO was mined.
// CSV - check sequence verify.
func TestCheckSequenceVerifyConditionedByBlockHeight(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		consensusConfig.BlockCoinbaseMaturity = 0
		factory := consensus.NewFactory()
		testConsensus, teardown, err := factory.NewTestConsensus(consensusConfig, "TestCheckLockTimeVerifyConditionedByBlockHeight")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)
		blockAHash, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{testConsensus.DAGParams().GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error creating blockA: %v", err)
		}
		blockBHash, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{blockAHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error creating blockB: %v", err)
		}
		blockCHash, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{blockBHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error creating blockC: %v", err)
		}
		blockC, err := testConsensus.GetBlock(blockCHash)
		if err != nil {
			t.Fatalf("Failed getting blockC: %v", err)
		}
		fees := uint64(1)
		fundingTransaction, err := testutils.CreateTransaction(blockC.Transactions[transactionhelper.CoinbaseTransactionIndex], fees)
		if err != nil {
			t.Fatalf("Error creating foundingTransaction: %v", err)
		}
		blockDHash, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{blockCHash}, nil,
			[]*externalapi.DomainTransaction{fundingTransaction})
		if err != nil {
			t.Fatalf("Failed creating blockD: %v", err)
		}
		//create a CSV script
		numOfBlocksToWait := int64(15)
		redeemScriptCSV, err := createCheckSequenceVerifyScript(numOfBlocksToWait)
		if err != nil {
			t.Fatalf("Failed to create a script using createCheckSequenceVerifyScript: %v", err)
		}
		p2shScriptCSV, err := txscript.PayToScriptHashScript(redeemScriptCSV)
		if err != nil {
			t.Fatalf("Failed to create a pay-to-script-hash script : %v", err)
		}
		scriptPublicKeyCSV := externalapi.ScriptPublicKey{
			Version: constants.MaxScriptPublicKeyVersion,
			Script:  p2shScriptCSV,
		}
		// the 23 bit in the nsequence defined if its conditioned by blockHeight(set to 0) / time (set to 1)
		blockHeightFlag := 0
		transactionWithLockedOutput, err := createTransactionWithLockedOutput(fundingTransaction, p2shScriptCSV, blockHeightFlag)
		if err != nil {
			t.Fatalf("Error in createTransactionWithLockedOutput: %v", err)
		}

	})
}

func createCheckSequenceVerifyScript(numOfBlocks int64) ([]byte, error) {
	scriptBuilder := txscript.NewScriptBuilder()
	scriptBuilder.AddOp(txscript.OpCheckSequenceVerify)
	scriptBuilder.AddInt64(numOfBlocks)
	scriptBuilder.AddOp(txscript.OpCheckSig)
	scriptBuilder.AddOp(txscript.OpTrue)
	return scriptBuilder.Script()
}
