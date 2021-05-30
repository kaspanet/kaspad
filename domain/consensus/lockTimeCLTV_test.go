package consensus_test

import (
	"errors"
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"testing"
)

// TestCheckLockTimeVerifyConditionedByBlockHeight verifies that an output locked by the CLTV script is spendable only after
// the block height reached the set target.
func TestCheckLockTimeVerifyConditionedByBlockHeight(t *testing.T) {
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
		foundingTransaction, err := testutils.CreateTransaction(blockC.Transactions[transactionhelper.CoinbaseTransactionIndex], fees)
		if err != nil {
			t.Fatalf("Error creating foundingTransaction: %v", err)
		}
		blockDHash, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{blockCHash}, nil,
			[]*externalapi.DomainTransaction{foundingTransaction})
		if err != nil {
			t.Fatalf("Error creating blockD: %v", err)
		}

		//Create a CLTV script:
		numOfBlocksToWait := int64(30)
		redeemScriptCLTV, err := createScriptCLTV(numOfBlocksToWait)
		if err != nil {
			t.Fatalf("Failed to create a script using createScriptCLTV: %v", err)
		}
		p2shScriptCLTV, err := txscript.PayToScriptHashScript(redeemScriptCLTV)
		if err != nil {
			t.Fatalf("Failed to create a pay-to-script-hash script : %v", err)
		}
		scriptPublicKeyCLTV := externalapi.ScriptPublicKey{
			Version: constants.MaxScriptPublicKeyVersion,
			Script:  p2shScriptCLTV,
		}
		transactionWithLockedOutput, err := createTransactionWithLockedOutput(foundingTransaction, fees, &scriptPublicKeyCLTV)
		// BlockE contains the locked output (locked by CLTV).
		// This block should be valid since CLTV script locked only the output.
		blockEHash, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{blockDHash}, nil,
			[]*externalapi.DomainTransaction{transactionWithLockedOutput})
		if err != nil {
			t.Fatalf("Error creating blockE: %v", err)
		}
		// Create a transaction that tries to spend the locked output.
		transactionThatSpentTheLockedOutput, err := createTransactionThatSpentTheLockedOutput(transactionWithLockedOutput,
			fees, redeemScriptCLTV, numOfBlocksToWait)
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

		// Add x blocks to release the locked output, where x = 'numOfBlocksToWait'.
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

// TestCheckLockTimeVerifyConditionedByAbsoluteTime verifies that an output locked by the CLTV script is spendable only after
// the time is reached to the set target (compared to the past median time)).
func TestCheckLockTimeVerifyConditionedByAbsoluteTime(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		consensusConfig.BlockCoinbaseMaturity = 0
		factory := consensus.NewFactory()
		testConsensus, teardown, err := factory.NewTestConsensus(consensusConfig, "TestCheckLockTimeVerifyConditionedByAbsoluteTime")
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
		foundingTransaction, err := testutils.CreateTransaction(blockC.Transactions[transactionhelper.CoinbaseTransactionIndex], fees)
		if err != nil {
			t.Fatalf("Error creating foundingTransaction: %v", err)
		}
		blockDHash, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{blockCHash}, nil,
			[]*externalapi.DomainTransaction{foundingTransaction})
		if err != nil {
			t.Fatalf("Error creating blockD: %v", err)
		}
		blockD, err := testConsensus.GetBlock(blockDHash)
		if err != nil {
			t.Fatalf("Failed getting blockC: %v", err)
		}
		//Create a CLTV script:
		numOfSecondsToWait := int64(12 * 1000)
		lockTimeStamp := blockD.Header.TimeInMilliseconds() + numOfSecondsToWait
		redeemScriptCLTV, err := createScriptCLTV(lockTimeStamp)
		if err != nil {
			t.Fatalf("Failed to create a script using createScriptCLTV: %v", err)
		}
		p2shScriptCLTV, err := txscript.PayToScriptHashScript(redeemScriptCLTV)
		if err != nil {
			t.Fatalf("Failed to create a pay-to-script-hash script : %v", err)
		}
		scriptPublicKeyCLTV := externalapi.ScriptPublicKey{
			Version: constants.MaxScriptPublicKeyVersion,
			Script:  p2shScriptCLTV,
		}
		transactionWithLockedOutput, err := createTransactionWithLockedOutput(foundingTransaction, fees, &scriptPublicKeyCLTV)
		// BlockE contains the locked output (locked by CLTV).
		// This block should be valid since CLTV script locked only the output.
		blockEHash, _, err := testConsensus.AddBlock([]*externalapi.DomainHash{blockDHash}, nil,
			[]*externalapi.DomainTransaction{transactionWithLockedOutput})
		if err != nil {
			t.Fatalf("Error creating blockE: %v", err)
		}
		blockE, err := testConsensus.GetBlock(blockEHash)
		if err != nil {
			t.Fatalf("Failed getting blockE: %v", err)
		}
		// Create a transaction that tries to spend the locked output.
		transactionThatSpentTheLockedOutput, err := createTransactionThatSpentTheLockedOutput(transactionWithLockedOutput,
			fees, redeemScriptCLTV, lockTimeStamp)
		if err != nil {
			t.Fatalf("Error creating transactionThatSpentTheLockedOutput: %v", err)
		}
		// Add a block that contains a transaction that tries to spend the locked output before the time, and therefore should be failed.
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
		timeStampBlockE := blockE.Header.TimeInMilliseconds()
		stagingArea := model.NewStagingArea()
		// Make sure the time limitation has passed.
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
			if pastMedianTime > lockTimeStamp {
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

func createScriptCLTV(blockHeightOrAbsoluteTime int64) ([]byte, error) {
	scriptBuilder := txscript.NewScriptBuilder()
	scriptBuilder.AddInt64(blockHeightOrAbsoluteTime)
	scriptBuilder.AddOp(txscript.OpCheckLockTimeVerify)
	scriptBuilder.AddOp(txscript.OpTrue)
	return scriptBuilder.Script()
}

func createTransactionWithLockedOutput(txToSpend *externalapi.DomainTransaction, fee uint64,
	scriptPublicKeyCLTV *externalapi.ScriptPublicKey) (*externalapi.DomainTransaction, error) {

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
		ScriptPublicKey: scriptPublicKeyCLTV,
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
	redeemScript []byte, lockTime int64) (*externalapi.DomainTransaction, error) {

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
		Sequence:        0xffffffff - 1,
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
		LockTime: uint64(lockTime), // less than 500 million interpreted as a block height, and above as an UNIX timestamp.
	}, nil
}
