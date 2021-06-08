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

// TestCheckSequenceVerifyConditionedByBlockHeight verifies that locked output (by CSV script) is spendable
// only after a certain number of blocks have been added relative to the time the UTXO was mined.
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
		sequenceFlag := 0
		// Create a transaction that tries to spend the locked output.
		transactionThatSpentTheLockedOutput, err := createTransactionThatSpentTheLockedOutput(transactionWithLockedOutput,
			fees, redeemScriptCSV, uint64(numOfBlocksToWait), sequenceFlag, blockEHash, &testConsensus)
		if err != nil {
			t.Fatalf("Error creating transactionThatSpentTheLockedOutput: %v", err)
		}
		// Add a block that contains a transaction that spends the locked output before the time, and therefore should be failed.
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
		// Tries to spend the output that should be no longer locked.
		_, _, err = testConsensus.AddBlock([]*externalapi.DomainHash{tipHash}, nil,
			[]*externalapi.DomainTransaction{transactionThatSpentTheLockedOutput})
		if err != nil {
			t.Fatalf("The block should be valid since the output is not locked anymore. but got an error: %v", err)
		}
	})
}

// TestCheckSequenceVerifyConditionedByRelativeTime verifies that locked output (by CSV script) is spendable only after
// the time is reached to the set target relative to the time the UTXO was mined (compared to the past median time).
func TestCheckSequenceVerifyConditionedByRelativeTime(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		consensusConfig.BlockCoinbaseMaturity = 0
		factory := consensus.NewFactory()
		testConsensus, teardown, err := factory.NewTestConsensus(consensusConfig, "TestCheckSequenceVerifyConditionedByRelativeTime")
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
			t.Fatalf("More than the allowed time to set.")
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
		sequenceFlag := 1
		// Create a transaction that tries to spend the locked output.
		transactionThatSpentTheLockedOutput, err := createTransactionThatSpentTheLockedOutput(transactionWithLockedOutput,
			fees, redeemScriptCSV, uint64(timeToWait), sequenceFlag, blockEHash, &testConsensus)
		if err != nil {
			t.Fatalf("Error creating transactionThatSpentTheLockedOutput: %v", err)
		}
		// Add a block that contains a transaction that spends the locked output before the time, and therefore should be failed.
		_, _, err = testConsensus.AddBlock([]*externalapi.DomainHash{blockEHash}, nil,
			[]*externalapi.DomainTransaction{transactionThatSpentTheLockedOutput})
		if err == nil || !errors.Is(err, ruleerrors.ErrUnfinalizedTx) {
			t.Fatalf("Expected block to be invalid with err: %v, instead found: %v", ruleerrors.ErrUnfinalizedTx, err)
		}
		emptyCoinbase := externalapi.DomainCoinbaseData{
			ScriptPublicKey: &externalapi.ScriptPublicKey{
				Script:  nil,
				Version: 0,
			},
		}
		var tipHash *externalapi.DomainHash
		blockE, err := testConsensus.GetBlock(blockEHash)
		if err != nil {
			t.Fatalf("Failed to get blockE: %v", err)
		}
		timeStampBlockE := blockE.Header.TimeInMilliseconds()
		stagingArea := model.NewStagingArea()
		// Make sure the time limitation has passed.
		lockTimeTarget := blockE.Header.TimeInMilliseconds() + timeToWait
		for i := int64(0); ; i++ {
			tipBlock, err := testConsensus.BuildBlock(&emptyCoinbase, nil)
			if err != nil {
				t.Fatalf("Error creating tip using BuildBlock: %v", err)
			}
			blockHeader := tipBlock.Header.ToMutable()
			blockHeader.SetTimeInMilliseconds(timeStampBlockE + i*1000)
			tipBlock.Header = blockHeader.ToImmutable()
			_, err = testConsensus.ValidateAndInsertBlock(tipBlock)
			if err != nil {
				t.Fatalf("Error validating and inserting tip block: %v", err)
			}
			tipHash = consensushashing.BlockHash(tipBlock)
			pastMedianTime, err := testConsensus.PastMedianTimeManager().PastMedianTime(stagingArea, tipHash)
			if err != nil {
				t.Fatalf("Failed getting pastMedianTime: %v", err)
			}
			if pastMedianTime > lockTimeTarget {
				break
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

//TestRelativeTimeOnCheckSequenceVerify verifies that if the relative target is set to X blocks to wait, and the absolute height
// will be X before adding all the blocks, then the output will remain locked.
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
		sequenceFlag := 0
		// Create a transaction that tries to spend the locked output.
		transactionThatSpentTheLockedOutput, err := createTransactionThatSpentTheLockedOutput(transactionWithLockedOutput,
			fees, redeemScriptCSV, uint64(numOfBlocksToWait), sequenceFlag, blockEHash, &testConsensus)
		if err != nil {
			t.Fatalf("Error creating transactionThatSpentTheLockedOutput: %v", err)
		}
		// Mines blocks so the block height will be the same as the relative number(but not enough to reach the relative target)
		// and verify that the output is still locked.
		// For unlocked the output the blocks should be count from the block that contains the locked output and not as an absolute height.
		tipHash := blockEHash
		for currentNumOfBlocks == numOfBlocksToWait {
			tipHash, _, err = testConsensus.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("Error creating tip: %v", err)
			}
			currentNumOfBlocks++
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
	redeemScript []byte, lockTime uint64, sequenceFlag23Bit int, lockedOutputBlockHash *externalapi.DomainHash,
	testConsensus *testapi.TestConsensus) (*externalapi.DomainTransaction, error) {

	// the 31bit is off since its relative timelock.
	sequence := uint64(0)
	sequence |= lockTime
	// conditioned by absolute time:
	if sequenceFlag23Bit == 1 {
		sequence |= 1 << 23
		lockedOutputBlock, err := (*testConsensus).GetBlock(lockedOutputBlockHash)
		if err != nil {
			return nil, err
		}
		lockTime += uint64(lockedOutputBlock.Header.TimeInMilliseconds())
	} else {
		// conditioned by block height:
		blockDAAScore, err := (*testConsensus).DAABlocksStore().DAAScore((*testConsensus).DatabaseContext(),
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
	return &externalapi.DomainTransaction{
		Version:  constants.MaxTransactionVersion,
		Inputs:   []*externalapi.DomainTransactionInput{input},
		Outputs:  []*externalapi.DomainTransactionOutput{output},
		Payload:  []byte{},
		LockTime: lockTime,
	}, nil
}
