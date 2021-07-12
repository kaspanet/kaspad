package consensus_test

import (
	"testing"

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
)

// TestCheckSequenceVerifyConditionedByDAAScore verifies that a locked output (by CSV script) is spendable
// only after a certain number of blocks have been added relative to the time the UTXO was mined.
// CSV - check sequence verify.
func TestCheckSequenceVerifyConditionedByDAAScore(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		consensusConfig.BlockCoinbaseMaturity = 0
		factory := consensus.NewFactory()
		testConsensus, teardown, err := factory.NewTestConsensus(consensusConfig, "TestCheckSequenceVerifyConditionedByDAAScore")
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
		blockDHash, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{blockCHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error creating blockD: %v", err)
		}
		blockD, err := testConsensus.GetBlock(blockDHash)
		if err != nil {
			t.Fatalf("Failed getting blockD: %v", err)
		}
		fees := uint64(1)
		// Create a CSV script
		numOfDAAScoreToWait := uint64(10)
		redeemScriptCSV, err := createScriptCSV(numOfDAAScoreToWait)
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
		transactionWithLockedOutput, err := CreateTransactionWithLockedOutput(
			blockD.Transactions[transactionhelper.CoinbaseTransactionIndex], fees, &scriptPublicKeyCSV)
		if err != nil {
			t.Fatalf("Error in CreateTransactionWithLockedOutput: %v", err)
		}
		// BlockE contains the locked output (locked by CSV).
		// This block should be valid since CSV script locked only the output.
		blockEHash, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{blockDHash}, nil,
			[]*externalapi.DomainTransaction{transactionWithLockedOutput})
		if err != nil {
			t.Fatalf("Error creating blockE: %v", err)
		}
		// bit 62 of sequence defines if it's conditioned by DAA score(set to 0) or by time (set to 1).
		sequenceTypeFlag := 0
		// Create a transaction that tries to spend the locked output.
		transactionThatSpentTheLockedOutput, err := createTransactionThatSpentTheLockedOutputRelativeLock(transactionWithLockedOutput,
			fees, redeemScriptCSV, numOfDAAScoreToWait, sequenceTypeFlag, blockEHash, &testConsensus)
		if err != nil {
			t.Fatalf("Error creating transactionThatSpentTheLockedOutput: %v", err)
		}
		// Add a block that contains a transaction that spends the locked output before the time, and therefore should be failed.
		_, _, err = testConsensus.AddBlock([]*externalapi.DomainHash{blockEHash}, nil,
			[]*externalapi.DomainTransaction{transactionThatSpentTheLockedOutput})
		if err == nil || !errors.Is(err, ruleerrors.ErrUnfinalizedTx) {
			t.Fatalf("Expected block to be invalid with err: %v, instead found: %v", ruleerrors.ErrUnfinalizedTx, err)
		}
		// Adds blocks until it reaches the DAA score target, so the locked output will be released.
		tipHash := blockEHash
		stagingArea := model.NewStagingArea()
		blockEDAAScore, err := testConsensus.DAABlocksStore().DAAScore(testConsensus.DatabaseContext(), stagingArea, blockEHash)
		if err != nil {
			t.Fatalf("Failed getting DAA score of blockE: %v", err)
		}
		targetDAAScore := blockEDAAScore + numOfDAAScoreToWait
		currentDAAScore, err := testConsensus.DAABlocksStore().DAAScore(testConsensus.DatabaseContext(), stagingArea, tipHash)
		if err != nil {
			t.Fatalf("Failed getting DAA score: %v", err)
		}
		for currentDAAScore <= targetDAAScore {
			tipHash, _, err = testConsensus.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("Error creating a tip: %v", err)
			}
			currentDAAScore, err = testConsensus.DAABlocksStore().DAAScore(testConsensus.DatabaseContext(), stagingArea, tipHash)
			if err != nil {
				t.Fatalf("Failed getting DAA score: %v", err)
			}
		}
		// Tries to spend the output that should be no longer locked.
		validBlock, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{tipHash}, nil,
			[]*externalapi.DomainTransaction{transactionThatSpentTheLockedOutput})
		if err != nil {
			t.Fatalf("The block should be valid since the output is not locked anymore. but got an error: %v", err)
		}
		validBlockStatus, err := testConsensus.BlockStatusStore().Get(testConsensus.DatabaseContext(), stagingArea,
			validBlock)
		if err != nil {
			t.Fatalf("Failed getting the status for validBlock: %v", err)
		}
		if !validBlockStatus.Equal(externalapi.StatusUTXOValid) {
			t.Fatalf("The status of validBlock should be: %v, but got: %v", externalapi.StatusUTXOValid,
				validBlockStatus)
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
		blockDHash, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{blockCHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error creating blockD: %v", err)
		}
		blockD, err := testConsensus.GetBlock(blockDHash)
		if err != nil {
			t.Fatalf("Failed getting blockD: %v", err)
		}
		fees := uint64(1)
		//create a CSV script
		timeToWait := uint64(14) // in seconds
		sequence := timeToWait | constants.SequenceLockTimeIsSeconds
		redeemScriptCSV, err := createScriptCSV(sequence)
		if err != nil {
			t.Fatalf("Failed to create a script using createScriptCSV: %v", err)
		}
		p2shScriptCSV, err := txscript.PayToScriptHashScript(redeemScriptCSV)
		if err != nil {
			t.Fatalf("Failed to create a pay-to-script-hash script : %v", err)
		}
		scriptPublicKeyCSV := externalapi.ScriptPublicKey{
			Version: constants.MaxScriptPublicKeyVersion,
			Script:  p2shScriptCSV,
		}
		transactionWithLockedOutput, err := CreateTransactionWithLockedOutput(blockD.Transactions[transactionhelper.CoinbaseTransactionIndex],
			fees, &scriptPublicKeyCSV)
		if err != nil {
			t.Fatalf("Error in CreateTransactionWithLockedOutput: %v", err)
		}
		// BlockE contains the locked output (locked by CSV).
		// This block should be valid since CSV script locked only the output.
		blockEHash, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{blockDHash}, nil,
			[]*externalapi.DomainTransaction{transactionWithLockedOutput})
		if err != nil {
			t.Fatalf("Error creating blockE: %v", err)
		}
		// bit 62 of sequence defines if it's conditioned by DAA score(set to 0) or by time (set to 1).
		sequenceTypeFlag := 1
		// Create a transaction that tries to spend the locked output.
		transactionThatSpentTheLockedOutput, err := createTransactionThatSpentTheLockedOutputRelativeLock(transactionWithLockedOutput,
			fees, redeemScriptCSV, timeToWait, sequenceTypeFlag, blockEHash, &testConsensus)
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

		lockTimeTarget := uint64(blockE.Header.TimeInMilliseconds()) + timeToWait*constants.SequenceLockTimeGranularity
		for i := int64(0); ; i++ {
			tipBlock, err := testConsensus.BuildBlock(&emptyCoinbase, nil)
			if err != nil {
				t.Fatalf("Error creating tip using BuildBlock: %v", err)
			}
			blockHeader := tipBlock.Header.ToMutable()
			blockHeader.SetTimeInMilliseconds(timeStampBlockE + i*constants.SequenceLockTimeGranularity)
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
			if uint64(pastMedianTime) > lockTimeTarget {
				break
			}
		}
		// Tries to spend the output that should be no longer locked
		validBlock, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{tipHash}, nil,
			[]*externalapi.DomainTransaction{transactionThatSpentTheLockedOutput})
		if err != nil {
			t.Fatalf("The block should be valid since the output is not locked anymore. but got an error: %v", err)
		}
		validBlockStatus, err := testConsensus.BlockStatusStore().Get(testConsensus.DatabaseContext(), stagingArea,
			validBlock)
		if err != nil {
			t.Fatalf("Failed getting the status for validBlock: %v", err)
		}
		if !validBlockStatus.Equal(externalapi.StatusUTXOValid) {
			t.Fatalf("The status of validBlock should be: %v, but got: %v", externalapi.StatusUTXOValid,
				validBlockStatus)
		}
	})
}

// TestRelativeTimeOnCheckSequenceVerify verifies that if the relative target is set to be X DAA score,
// and the absolute DAA score is X before having X DAA score more than the time the UTXO was mined, then the output will remain locked.
func TestRelativeTimeOnCheckSequenceVerify(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		consensusConfig.BlockCoinbaseMaturity = 0
		factory := consensus.NewFactory()
		testConsensus, teardown, err := factory.NewTestConsensus(consensusConfig, "TestRelativeTimeOnCheckSequenceVerify")
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
		blockDHash, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{blockCHash}, nil, nil)
		if err != nil {
			t.Fatalf("Error creating blockC: %v", err)
		}
		blockD, err := testConsensus.GetBlock(blockDHash)
		if err != nil {
			t.Fatalf("Failed getting blockC: %v", err)
		}
		fees := uint64(1)
		//create a CSV script
		numOfDAAScoreToWait := uint64(10)
		redeemScriptCSV, err := createScriptCSV(numOfDAAScoreToWait)
		if err != nil {
			t.Fatalf("Failed to create a script using createScriptCSV: %v", err)
		}
		p2shScriptCSV, err := txscript.PayToScriptHashScript(redeemScriptCSV)
		if err != nil {
			t.Fatalf("Failed to create a pay-to-script-hash script : %v", err)
		}
		scriptPublicKeyCSV := externalapi.ScriptPublicKey{
			Version: constants.MaxScriptPublicKeyVersion,
			Script:  p2shScriptCSV,
		}
		transactionWithLockedOutput, err := CreateTransactionWithLockedOutput(blockD.Transactions[transactionhelper.CoinbaseTransactionIndex],
			fees, &scriptPublicKeyCSV)
		if err != nil {
			t.Fatalf("Error in CreateTransactionWithLockedOutput: %v", err)
		}
		// BlockE contains the locked output (locked by CSV).
		// This block should be valid since CSV script locked only the output.
		blockEHash, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{blockDHash}, nil,
			[]*externalapi.DomainTransaction{transactionWithLockedOutput})
		if err != nil {
			t.Fatalf("Error creating blockE: %v", err)
		}
		// bit 62 of sequence defines if it's conditioned by DAA score(set to 0) or by time (set to 1).
		sequenceTypeFlag := 0
		// Create a transaction that tries to spend the locked output.
		transactionThatSpentTheLockedOutput, err := createTransactionThatSpentTheLockedOutputRelativeLock(transactionWithLockedOutput,
			fees, redeemScriptCSV, numOfDAAScoreToWait, sequenceTypeFlag, blockEHash, &testConsensus)
		if err != nil {
			t.Fatalf("Error creating transactionThatSpentTheLockedOutput: %v", err)
		}
		tipHash := blockEHash
		stagingArea := model.NewStagingArea()
		currentDAAScore, err := testConsensus.DAABlocksStore().DAAScore(testConsensus.DatabaseContext(), stagingArea, tipHash)
		if err != nil {
			t.Fatalf("Failed getting DAA score for tip: %v", err)
		}
		// Mines blocks until the DAA score will be the same as the relative number(but not enough to reach the relative target - relative
		// number + DAA score of the block which contains the locked output ) and verify that the output is still locked.
		for currentDAAScore != numOfDAAScoreToWait {
			tipHash, _, err = testConsensus.AddBlock([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("Error creating tip: %v", err)
			}
			currentDAAScore, err = testConsensus.DAABlocksStore().DAAScore(testConsensus.DatabaseContext(), stagingArea, tipHash)
			if err != nil {
				t.Fatalf("Failed getting DAA score for tip: %v", err)
			}
		}
		// After the above for loop, the latest block has 10 DAA score, but the output will be unlocked only when the DAA score will be 15,
		// so this block is expected to be considered invalid.
		_, _, err = testConsensus.AddBlock([]*externalapi.DomainHash{tipHash}, nil,
			[]*externalapi.DomainTransaction{transactionThatSpentTheLockedOutput})
		if err == nil || !errors.Is(err, ruleerrors.ErrUnfinalizedTx) {
			t.Fatalf("Expected block to be invalid with err: %v, instead found: %v", ruleerrors.ErrUnfinalizedTx, err)
		}
	})
}

func createScriptCSV(sequence uint64) ([]byte, error) {
	scriptBuilder := txscript.NewScriptBuilder()
	scriptBuilder.AddSequenceNumber(sequence)
	scriptBuilder.AddOp(txscript.OpCheckSequenceVerify)
	scriptBuilder.AddOp(txscript.OpTrue)
	return scriptBuilder.Script()
}

func createTransactionThatSpentTheLockedOutputRelativeLock(txToSpend *externalapi.DomainTransaction, fee uint64,
	redeemScript []byte, numOfDAAScoreOrTimeForRelativeWaiting uint64, sequenceTypeFlag int, lockedOutputBlockHash *externalapi.DomainHash,
	testConsensus *testapi.TestConsensus) (*externalapi.DomainTransaction, error) {

	var lockTime, sequence uint64
	if sequenceTypeFlag == 1 { // Conditioned by time:
		sequence = numOfDAAScoreOrTimeForRelativeWaiting // In seconds
		sequence |= constants.SequenceLockTimeIsSeconds
		lockedOutputBlock, err := (*testConsensus).GetBlock(lockedOutputBlockHash)
		if err != nil {
			return nil, err
		}
		stamp := uint64(lockedOutputBlock.Header.TimeInMilliseconds())
		lockTime = numOfDAAScoreOrTimeForRelativeWaiting*constants.SequenceLockTimeGranularity + stamp // In milliseconds
	} else { // conditioned by DAA score:
		sequence = numOfDAAScoreOrTimeForRelativeWaiting
		blockDAAScore, err := (*testConsensus).DAABlocksStore().DAAScore((*testConsensus).DatabaseContext(),
			model.NewStagingArea(), lockedOutputBlockHash)
		if err != nil {
			return nil, err
		}
		lockTime = numOfDAAScoreOrTimeForRelativeWaiting + blockDAAScore
	}
	if sequence&constants.SequenceLockTimeDisabled == constants.SequenceLockTimeDisabled {
		return nil, errors.New("The flag SequenceLockTimeDisabled is raised even though it's a relative lock.")
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
