package consensus_test

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/pkg/errors"
	"testing"
)

//todo: Fix the relative block, open an issue on the isFinalized that should be with daa score instead of the blue score, add a test
// with CSV with time.

// TestCheckSequenceVerifyConditionedByBlockHeight verifies that locked output (by CSV script) is spendable only after
// a certain number of blocks have added relative to the time the UTXO was mined.
// CSV - check sequence verify.
func TestCheckSequenceVerifyConditionedByBlockHeight(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		consensusConfig.BlockCoinbaseMaturity = 0
		factory := consensus.NewFactory()
		testConsensus, teardown, err := factory.NewTestConsensus(consensusConfig, "TestCheckSequenceVerifyConditionedByBlockHeight")
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
		numOfBlocksToWait := int64(10)
		if numOfBlocksToWait > 0xffff {
			t.Fatalf("More than the maximum number of blocks allowed.")
		}
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
		transactionWithLockedOutput, err := createTransactionWithLockedOutput(fundingTransaction, fees, &scriptPublicKeyCSV)
		if err != nil {
			t.Fatalf("Error in createTransactionWithLockedOutput: %v", err)
		}
		// BlockE contains the locked output (locked by CSV).
		// This block should be valid since CSV script locked only the output.
		blockEHash, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{blockDHash}, nil,
			[]*externalapi.DomainTransaction{transactionWithLockedOutput})
		if err != nil {
			t.Fatalf("Error creating blockE: %v", err)
		}
		// The 23-bit of sequence defines if it's conditioned by block height(set to 0) or by time (set to 1).
		blockHeightBitFlag := 0
		// Create a transaction that tries to spend the locked output.
		transactionThatSpentTheLockedOutput, err := createTransactionThatSpentTheLockedOutput(transactionWithLockedOutput,
			fees, redeemScriptCSV, uint64(numOfBlocksToWait), blockHeightBitFlag, blockEHash, &testConsensus)
		if err != nil {
			t.Fatalf("Error creating transactionThatSpentTheLockedOutput: %v", err)
		}
		// Add a block that contains a transaction that spends the locked output before the time, and therefore should be failed.
		// (x blocks should be added before the output will be spendable, where x = 'numOfBlocksToWait' ).
		_, _, err = testConsensus.AddBlock([]*externalapi.DomainHash{blockEHash}, nil,
			[]*externalapi.DomainTransaction{transactionThatSpentTheLockedOutput})
		if err == nil || !errors.Is(err, ruleerrors.ErrUnfinalizedTx) {
			t.Fatalf("Expected block to be invalid with err: %v, instead found: %v", ruleerrors.ErrUnfinalizedTx, err)
		}
		//Add x blocks to release the locked output, where x = 'numOfBlocksToWait'.
		tipHash := blockEHash
		for i := int64(0); i < numOfBlocksToWait; i++ {
			tipHash, _, err = testConsensus.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("Error creating tip: %v", err)
			}
		}
		// Tries to spend the output that should be no longer locked
		_, _, err = testConsensus.AddBlock([]*externalapi.DomainHash{tipHash}, nil,
			[]*externalapi.DomainTransaction{transactionThatSpentTheLockedOutput})
		if err != nil {
			t.Fatalf("The block should be valid since the output is not locked anymore. but got an error: %v", err)
		}
	})
}

// TestCheckSequenceVerifyConditionedByAbsoluteTime verifies that locked output (by CSV script) is spendable only after
// the time is reached to the set target relative to the time the UTXO was mined (compared to the past median time).
func TestCheckSequenceVerifyConditionedByAbsoluteTime(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		consensusConfig.BlockCoinbaseMaturity = 0
		factory := consensus.NewFactory()
		testConsensus, teardown, err := factory.NewTestConsensus(consensusConfig, "TestCheckSequenceVerifyConditionedByAbsoluteTime")
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
		timeToWait := int64(14 * 1000)
		if timeToWait > 0xffff {
			t.Fatalf("More than the allowed time")
		}
		redeemScriptCSV, err := createCheckSequenceVerifyScript(timeToWait)
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
		transactionWithLockedOutput, err := createTransactionWithLockedOutput(fundingTransaction, fees, &scriptPublicKeyCSV)
		if err != nil {
			t.Fatalf("Error in createTransactionWithLockedOutput: %v", err)
		}
		// BlockE contains the locked output (locked by CSV).
		// This block should be valid since CSV script locked only the output.
		blockEHash, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{blockDHash}, nil,
			[]*externalapi.DomainTransaction{transactionWithLockedOutput})
		if err != nil {
			t.Fatalf("Error creating blockE: %v", err)
		}
		// The 23-bit of sequence defines if it's conditioned by block height(set to 0) or by time (set to 1).
		blockHeightBitFlag := 1
		// Create a transaction that tries to spend the locked output.
		transactionThatSpentTheLockedOutput, err := createTransactionThatSpentTheLockedOutput(transactionWithLockedOutput,
			fees, redeemScriptCSV, uint64(timeToWait), blockHeightBitFlag, blockEHash, &testConsensus)
		if err != nil {
			t.Fatalf("Error creating transactionThatSpentTheLockedOutput: %v", err)
		}
		// Add a block that contains a transaction that spends the locked output before the time, and therefore should be failed.
		// (x blocks should be added before the output will be spendable, where x = 'numOfBlocksToWait' ).
		_, _, err = testConsensus.AddBlock([]*externalapi.DomainHash{blockEHash}, nil,
			[]*externalapi.DomainTransaction{transactionThatSpentTheLockedOutput})
		if err == nil || !errors.Is(err, ruleerrors.ErrUnfinalizedTx) {
			t.Fatalf("Expected block to be invalid with err: %v, instead found: %v", ruleerrors.ErrUnfinalizedTx, err)
		}
		//Add x blocks to release the locked output, where x = 'numOfBlocksToWait'.
		//tipHash := blockEHash
		//for i := int64(0); i < numOfBlocksToWait; i++ {
		//	tipHash, _, err = testConsensus.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
		//	if err != nil {
		//		t.Fatalf("Error creating tip: %v", err)
		//	}
		//}
		// Tries to spend the output that should be no longer locked
		//_, _, err = testConsensus.AddBlock([]*externalapi.DomainHash{tipHash}, nil,
		//	[]*externalapi.DomainTransaction{transactionThatSpentTheLockedOutput})
		//if err != nil {
		//	t.Fatalf("The block should be valid since the output is not locked anymore. but got an error: %v", err)
		//}
	})
}

//TestRelativeTimeOnCheckSequenceVerify verifies that if the relative target is set to X blocks to wait and the absolute height
// will be X before adding all the blocks then the output will remain locked( until all the X blocks are added).
func TestRelativeTimeOnCheckSequenceVerify(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		consensusConfig.BlockCoinbaseMaturity = 0
		factory := consensus.NewFactory()
		testConsensus, teardown, err := factory.NewTestConsensus(consensusConfig, "TestRelativeTimeOnCheckSequenceVerify")
		if err != nil {
			t.Fatalf("Error setting up consensus: %+v", err)
		}
		defer teardown(false)

		currentNumOfBlocks := int64(0)
		blockAHash, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{testConsensus.DAGParams().GenesisHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error creating blockA: %v", err)
		}
		currentNumOfBlocks++
		blockBHash, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{blockAHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error creating blockB: %v", err)
		}
		currentNumOfBlocks++
		blockCHash, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{blockBHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error creating blockC: %v", err)
		}
		currentNumOfBlocks++
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
		currentNumOfBlocks++
		//create a CSV script
		numOfBlocksToWait := int64(10)
		if numOfBlocksToWait > 0xffff {
			t.Fatalf("More than the max number of blocks that allowed to set.")
		}
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
		transactionWithLockedOutput, err := createTransactionWithLockedOutput(fundingTransaction, fees, &scriptPublicKeyCSV)
		if err != nil {
			t.Fatalf("Error in createTransactionWithLockedOutput: %v", err)
		}
		// BlockE contains the locked output (locked by CSV).
		// This block should be valid since CSV script locked only the output.
		blockEHash, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{blockDHash}, nil,
			[]*externalapi.DomainTransaction{transactionWithLockedOutput})
		if err != nil {
			t.Fatalf("Error creating blockE: %v", err)
		}
		currentNumOfBlocks++
		// The 23-bit of sequence defines if it's conditioned by block height(set to 0) or by time (set to 1).
		blockHeightBitFlag := 0
		blockEDAAScore, err := testConsensus.DAABlocksStore().DAAScore(testConsensus.DatabaseContext(), model.NewStagingArea(), blockEHash)
		if err != nil {
			t.Fatalf("Failed to get DAAScore: %v", err)
		}
		// Create a transaction that tries to spend the locked output.
		transactionThatSpentTheLockedOutput, err := createTransactionThatSpentTheLockedOutput(transactionWithLockedOutput,
			fees, redeemScriptCSV, uint64(numOfBlocksToWait), blockHeightBitFlag, blockEHash, &testConsensus)
		if err != nil {
			t.Fatalf("Error creating transactionThatSpentTheLockedOutput: %v", err)
		}
		// Mines blocks so the block height will be the same as the relative number(but not enough to reach the relative target)
		// and verify that the output is still locked.
		// The blocks should be count from the block that contains the locked output and not as an absolute height.
		// If changing the numbers in this test, please verifies this for loop is still valid.
		tipHash := blockEHash
		for i := int64(0); i < numOfBlocksToWait-currentNumOfBlocks; i++ {
			tipHash, _, err = testConsensus.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("Error creating tip: %v", err)
			}
		}
		_, _, err = testConsensus.AddBlock([]*externalapi.DomainHash{tipHash}, nil,
			[]*externalapi.DomainTransaction{transactionThatSpentTheLockedOutput})
		if err == nil || !errors.Is(err, ruleerrors.ErrUnfinalizedTx) {
			t.Fatalf("Expected block to be invalid with err: %v, instead found: %v", ruleerrors.ErrUnfinalizedTx, err)
		}
	})
}

func createCheckSequenceVerifyScript(numOfBlocks int64) ([]byte, error) {
	scriptBuilder := txscript.NewScriptBuilder()
	scriptBuilder.AddOp(txscript.OpCheckSequenceVerify)
	scriptBuilder.AddInt64(numOfBlocks)
	scriptBuilder.AddOp(txscript.OpTrue)
	return scriptBuilder.Script()
}

func createTransactionWithLockedOutput(txToSpend *externalapi.DomainTransaction, fee uint64,
	scriptPublicKeyCSV *externalapi.ScriptPublicKey) (*externalapi.DomainTransaction, error) {

	_, redeemScript := testutils.OpTrueScript()
	signatureScript, err := txscript.PayToScriptHashSignatureScript(redeemScript, nil)
	if err != nil {
		return nil, err
	}
	input := &externalapi.DomainTransactionInput{
		PreviousOutpoint: externalapi.DomainOutpoint{
			TransactionID: *consensushashing.TransactionID(txToSpend),
			Index:         0,
		},
		SignatureScript: signatureScript,
		Sequence:        constants.MaxTxInSequenceNum,
	}
	output := &externalapi.DomainTransactionOutput{
		ScriptPublicKey: scriptPublicKeyCSV,
		Value:           txToSpend.Outputs[0].Value - fee,
	}
	return &externalapi.DomainTransaction{
		Version: constants.MaxTransactionVersion,
		Inputs:  []*externalapi.DomainTransactionInput{input},
		Outputs: []*externalapi.DomainTransactionOutput{output},
		Payload: []byte{},
	}, nil
}

func createTransactionThatSpentTheLockedOutput(txToSpend *externalapi.DomainTransaction, fee uint64,
	redeemScript []byte, lockTime uint64, flag23Bit int, lockedOutputBlockHash *externalapi.DomainHash,
	testConsensus testapi.TestConsensus) (*externalapi.DomainTransaction, error) {

	// the 31bit is off since its relative timelock.
	sequence := uint64(0)
	sequence |= lockTime
	// conditioned by absolute time:
	if flag23Bit == 1 {
		sequence |= 1 << 23
		lockedOutputBlock, err := testConsensus.GetBlock(lockedOutputBlockHash)
		if err != nil {
			return nil, err
		}
		lockTime += uint64(lockedOutputBlock.Header.TimeInMilliseconds())
	} else {
		// conditioned by block height:
		blockDAAScore, err := testConsensus.DAABlocksStore().DAAScore(testConsensus.DatabaseContext(),
			model.NewStagingArea(), lockedOutputBlockHash)
		if err != nil {
			return nil, err
		}
		lockTime += blockDAAScore
	}
	signatureScript, err := txscript.PayToScriptHashSignatureScript(redeemScript, []byte{})
	if err != nil {
		return nil, err
	}
	scriptPublicKeyOutput, _ := testutils.OpTrueScript()
	input := &externalapi.DomainTransactionInput{
		PreviousOutpoint: externalapi.DomainOutpoint{
			TransactionID: *consensushashing.TransactionID(txToSpend),
			Index:         0,
		},
		SignatureScript: signatureScript,
		Sequence:        sequence,
	}
	output := &externalapi.DomainTransactionOutput{
		ScriptPublicKey: scriptPublicKeyOutput,
		Value:           txToSpend.Outputs[0].Value - fee,
	}
	// If its conditioned by block height than
	if flag23Bit == 0 {
	}
	return &externalapi.DomainTransaction{
		Version:  constants.MaxTransactionVersion,
		Inputs:   []*externalapi.DomainTransactionInput{input},
		Outputs:  []*externalapi.DomainTransactionOutput{output},
		Payload:  []byte{},
		LockTime: lockTime,
	}, nil
}
