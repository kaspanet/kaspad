package blockprocessor_test

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/pkg/errors"
	"testing"
	"time"
)

func addBlock(tcSyncer, tcSyncee testapi.TestConsensus, parentHashes []*externalapi.DomainHash, t *testing.T) *externalapi.DomainHash {
	block, _, err := tcSyncer.BuildBlockWithParents(parentHashes, nil, nil)
	if err != nil {
		t.Fatalf("BuildBlockWithParents: %+v", err)
	}

	_, err = tcSyncer.ValidateAndInsertBlock(block)
	if err != nil {
		t.Fatalf("ValidateAndInsertBlock: %+v", err)
	}

	_, err = tcSyncee.ValidateAndInsertBlock(&externalapi.DomainBlock{
		Header:       block.Header,
		Transactions: nil,
	})
	if err != nil {
		t.Fatalf("ValidateAndInsertBlock: %+v", err)
	}

	return consensushashing.BlockHash(block)
}

func TestValidateAndInsertImportedPruningPoint(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		// This is done to reduce the pruning depth to 6 blocks
		finalityDepth := 3
		params.FinalityDuration = time.Duration(finalityDepth) * params.TargetTimePerBlock
		params.K = 0

		factory := consensus.NewFactory()

		tcSyncer, teardownSyncer, err := factory.NewTestConsensus(params, false, "TestValidateAndInsertPruningPointSyncer")
		if err != nil {
			t.Fatalf("Error setting up tcSyncer: %+v", err)
		}
		defer teardownSyncer(false)

		tcSyncee, teardownSyncee, err := factory.NewTestConsensus(params, false, "TestValidateAndInsertPruningPointSyncee")
		if err != nil {
			t.Fatalf("Error setting up tcSyncee: %+v", err)
		}
		defer teardownSyncee(false)

		tipHash := params.GenesisHash
		for i := 0; i < finalityDepth-2; i++ {
			tipHash = addBlock(tcSyncer, tcSyncee, []*externalapi.DomainHash{tipHash}, t)
		}

		// Add block in the anticone of the pruning point to test such situation
		pruningPointAnticoneBlock := addBlock(tcSyncer, tcSyncee, []*externalapi.DomainHash{tipHash}, t)
		tipHash = addBlock(tcSyncer, tcSyncee, []*externalapi.DomainHash{tipHash}, t)
		nextPruningPoint := addBlock(tcSyncer, tcSyncee, []*externalapi.DomainHash{tipHash}, t)

		tipHash = addBlock(tcSyncer, tcSyncee, []*externalapi.DomainHash{pruningPointAnticoneBlock, nextPruningPoint}, t)

		// Add blocks until the pruning point changes
		for {
			tipHash = addBlock(tcSyncer, tcSyncee, []*externalapi.DomainHash{tipHash}, t)

			pruningPoint, err := tcSyncer.PruningPoint()
			if err != nil {
				t.Fatalf("PruningPoint: %+v", err)
			}

			if !pruningPoint.Equal(params.GenesisHash) {
				break
			}
		}

		pruningPoint, err := tcSyncer.PruningPoint()
		if err != nil {
			t.Fatalf("PruningPoint: %+v", err)
		}

		if !pruningPoint.Equal(nextPruningPoint) {
			t.Fatalf("Unexpected pruning point %s", pruningPoint)
		}

		pruningPointUTXOs, err := tcSyncer.GetPruningPointUTXOs(pruningPoint, nil, 1000)
		if err != nil {
			t.Fatalf("GetPruningPointUTXOs: %+v", err)
		}
		err = tcSyncee.AppendImportedPruningPointUTXOs(pruningPointUTXOs)
		if err != nil {
			t.Fatalf("AppendImportedPruningPointUTXOs: %+v", err)
		}

		tip, err := tcSyncer.GetBlock(tipHash)
		if err != nil {
			t.Fatalf("GetBlock: %+v", err)
		}

		// Check that ValidateAndInsertImportedPruningPoint fails for invalid pruning point
		err = tcSyncee.ValidateAndInsertImportedPruningPoint(tip)
		if !errors.Is(err, ruleerrors.ErrUnexpectedPruningPoint) {
			t.Fatalf("Unexpected error: %+v", err)
		}

		pruningPointBlock, err := tcSyncer.GetBlock(pruningPoint)
		if err != nil {
			t.Fatalf("GetBlock: %+v", err)
		}

		invalidPruningPointBlock := pruningPointBlock.Clone()
		invalidPruningPointBlock.Transactions[0].Version += 1

		// Check that ValidateAndInsertImportedPruningPoint fails for invalid block
		err = tcSyncee.ValidateAndInsertImportedPruningPoint(invalidPruningPointBlock)
		if !errors.Is(err, ruleerrors.ErrBadMerkleRoot) {
			t.Fatalf("Unexpected error: %+v", err)
		}

		err = tcSyncee.ClearImportedPruningPointData()
		if err != nil {
			t.Fatalf("ClearImportedPruningPointData: %+v", err)
		}
		err = tcSyncee.AppendImportedPruningPointUTXOs(makeFakeUTXOs())
		if err != nil {
			t.Fatalf("AppendImportedPruningPointUTXOs: %+v", err)
		}

		// Check that ValidateAndInsertImportedPruningPoint fails if the UTXO commitment doesn't fit the provided UTXO set.
		err = tcSyncee.ValidateAndInsertImportedPruningPoint(pruningPointBlock)
		if !errors.Is(err, ruleerrors.ErrBadPruningPointUTXOSet) {
			t.Fatalf("Unexpected error: %+v", err)
		}

		err = tcSyncee.ClearImportedPruningPointData()
		if err != nil {
			t.Fatalf("ClearImportedPruningPointData: %+v", err)
		}
		err = tcSyncee.AppendImportedPruningPointUTXOs(pruningPointUTXOs)
		if err != nil {
			t.Fatalf("AppendImportedPruningPointUTXOs: %+v", err)
		}

		// Check that ValidateAndInsertImportedPruningPoint works given the right arguments.
		err = tcSyncee.ValidateAndInsertImportedPruningPoint(pruningPointBlock)
		if err != nil {
			t.Fatalf("ValidateAndInsertImportedPruningPoint: %+v", err)
		}

		virtualSelectedParent, err := tcSyncer.GetVirtualSelectedParent()
		if err != nil {
			t.Fatalf("GetVirtualSelectedParent: %+v", err)
		}

		missingBlockBodyHashes, err := tcSyncee.GetMissingBlockBodyHashes(virtualSelectedParent)
		if err != nil {
			t.Fatalf("GetMissingBlockBodyHashes: %+v", err)
		}

		for _, missingHash := range missingBlockBodyHashes {
			block, err := tcSyncer.GetBlock(missingHash)
			if err != nil {
				t.Fatalf("GetBlock: %+v", err)
			}

			_, err = tcSyncee.ValidateAndInsertBlock(block)
			if err != nil {
				t.Fatalf("ValidateAndInsertBlock: %+v", err)
			}
		}

		synceeTips, err := tcSyncee.Tips()
		if err != nil {
			t.Fatalf("Tips: %+v", err)
		}

		syncerTips, err := tcSyncer.Tips()
		if err != nil {
			t.Fatalf("Tips: %+v", err)
		}

		if !externalapi.HashesEqual(synceeTips, syncerTips) {
			t.Fatalf("Syncee's tips are %s while syncer's are %s", synceeTips, syncerTips)
		}

		synceePruningPoint, err := tcSyncee.PruningPoint()
		if err != nil {
			t.Fatalf("PruningPoint: %+v", err)
		}

		if !synceePruningPoint.Equal(pruningPoint) {
			t.Fatalf("The syncee pruning point has not changed as exepcted")
		}
	})
}

// TestValidateAndInsertPruningPointWithSideBlocks makes sure that when a node applies a UTXO-Set downloaded during
// IBD, while it already has a non-empty UTXO-Set originating from blocks mined on top of genesis - the resulting
// UTXO set is correct
func TestValidateAndInsertPruningPointWithSideBlocks(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		// This is done to reduce the pruning depth to 6 blocks
		finalityDepth := 3
		params.FinalityDuration = time.Duration(finalityDepth) * params.TargetTimePerBlock
		params.K = 0

		factory := consensus.NewFactory()

		tcSyncer, teardownSyncer, err := factory.NewTestConsensus(params, false, "TestValidateAndInsertPruningPointSyncer")
		if err != nil {
			t.Fatalf("Error setting up tcSyncer: %+v", err)
		}
		defer teardownSyncer(false)

		tcSyncee, teardownSyncee, err := factory.NewTestConsensus(params, false, "TestValidateAndInsertPruningPointSyncee")
		if err != nil {
			t.Fatalf("Error setting up tcSyncee: %+v", err)
		}
		defer teardownSyncee(false)

		// Mine 2 block in the syncee on top of genesis
		side, _, err := tcSyncee.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, &externalapi.DomainCoinbaseData{ScriptPublicKey: &externalapi.ScriptPublicKey{}, ExtraData: []byte{1, 2}}, nil)
		if err != nil {
			t.Fatal(err)
		}
		_, _, err = tcSyncee.AddBlock([]*externalapi.DomainHash{side}, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		tipHash := params.GenesisHash
		for i := 0; i < finalityDepth-2; i++ {
			tipHash = addBlock(tcSyncer, tcSyncee, []*externalapi.DomainHash{tipHash}, t)
		}

		tipHash = addBlock(tcSyncer, tcSyncee, []*externalapi.DomainHash{tipHash}, t)
		nextPruningPoint := addBlock(tcSyncer, tcSyncee, []*externalapi.DomainHash{tipHash}, t)

		tipHash = addBlock(tcSyncer, tcSyncee, []*externalapi.DomainHash{nextPruningPoint}, t)

		// Add blocks until the pruning point changes
		for {
			tipHash = addBlock(tcSyncer, tcSyncee, []*externalapi.DomainHash{tipHash}, t)

			pruningPoint, err := tcSyncer.PruningPoint()
			if err != nil {
				t.Fatalf("PruningPoint: %+v", err)
			}

			if !pruningPoint.Equal(params.GenesisHash) {
				break
			}
		}

		pruningPoint, err := tcSyncer.PruningPoint()
		if err != nil {
			t.Fatalf("PruningPoint: %+v", err)
		}

		if !pruningPoint.Equal(nextPruningPoint) {
			t.Fatalf("Unexpected pruning point %s", pruningPoint)
		}

		pruningPointUTXOs, err := tcSyncer.GetPruningPointUTXOs(pruningPoint, nil, 1000)
		if err != nil {
			t.Fatalf("GetPruningPointUTXOs: %+v", err)
		}
		err = tcSyncee.AppendImportedPruningPointUTXOs(pruningPointUTXOs)
		if err != nil {
			t.Fatalf("AppendImportedPruningPointUTXOs: %+v", err)
		}

		// Check that ValidateAndInsertPruningPoint works.
		pruningPointBlock, err := tcSyncer.GetBlock(pruningPoint)
		if err != nil {
			t.Fatalf("GetBlock: %+v", err)
		}
		err = tcSyncee.ValidateAndInsertImportedPruningPoint(pruningPointBlock)
		if err != nil {
			t.Fatalf("ValidateAndInsertPruningPoint: %+v", err)
		}

		// Insert the rest of the blocks atop pruning point
		virtualSelectedParent, err := tcSyncer.GetVirtualSelectedParent()
		if err != nil {
			t.Fatalf("GetVirtualSelectedParent: %+v", err)
		}

		missingBlockBodyHashes, err := tcSyncee.GetMissingBlockBodyHashes(virtualSelectedParent)
		if err != nil {
			t.Fatalf("GetMissingBlockBodyHashes: %+v", err)
		}

		for _, missingHash := range missingBlockBodyHashes {
			block, err := tcSyncer.GetBlock(missingHash)
			if err != nil {
				t.Fatalf("GetBlock: %+v", err)
			}

			_, err = tcSyncee.ValidateAndInsertBlock(block)
			if err != nil {
				t.Fatalf("ValidateAndInsertBlock: %+v", err)
			}
		}

		// Verify that syncee and syncer tips are equal
		synceeTips, err := tcSyncee.Tips()
		if err != nil {
			t.Fatalf("Tips: %+v", err)
		}

		syncerTips, err := tcSyncer.Tips()
		if err != nil {
			t.Fatalf("Tips: %+v", err)
		}

		if !externalapi.HashesEqual(synceeTips, syncerTips) {
			t.Fatalf("Syncee's tips are %s while syncer's are %s", synceeTips, syncerTips)
		}

		// Verify that syncee and syncer pruning points are equal
		synceePruningPoint, err := tcSyncee.PruningPoint()
		if err != nil {
			t.Fatalf("PruningPoint: %+v", err)
		}

		if !synceePruningPoint.Equal(pruningPoint) {
			t.Fatalf("The syncee pruning point has not changed as exepcted")
		}

		pruningPointOld := pruningPoint

		// Add blocks until the pruning point moves, and verify it moved to the same point on both syncer and syncee
		for {
			block, _, err := tcSyncer.BuildBlockWithParents([]*externalapi.DomainHash{tipHash}, nil, nil)
			if err != nil {
				t.Fatalf("BuildBlockWithParents: %+v", err)
			}

			_, err = tcSyncer.ValidateAndInsertBlock(block)
			if err != nil {
				t.Fatalf("ValidateAndInsertBlock: %+v", err)
			}

			_, err = tcSyncee.ValidateAndInsertBlock(block)
			if err != nil {
				t.Fatalf("ValidateAndInsertBlock: %+v", err)
			}

			tipHash = consensushashing.BlockHash(block)

			pruningPoint, err = tcSyncer.PruningPoint()
			if err != nil {
				t.Fatalf("PruningPoint: %+v", err)
			}

			if !pruningPoint.Equal(pruningPointOld) {
				break
			}
		}

		synceePruningPoint, err = tcSyncee.PruningPoint()
		if err != nil {
			t.Fatalf("PruningPoint: %+v", err)
		}

		if !synceePruningPoint.Equal(pruningPoint) {
			t.Fatalf("The syncee pruning point(%s) is not equal to syncer pruning point (%s) after it moved. "+
				"pruning point before move: %s", synceePruningPoint, pruningPoint, pruningPointOld)
		}
	})
}

func makeFakeUTXOs() []*externalapi.OutpointAndUTXOEntryPair {
	return []*externalapi.OutpointAndUTXOEntryPair{
		{
			Outpoint: &externalapi.DomainOutpoint{
				TransactionID: externalapi.DomainTransactionID{},
				Index:         0,
			},
			UTXOEntry: utxo.NewUTXOEntry(
				0,
				&externalapi.ScriptPublicKey{
					Script:  nil,
					Version: 0,
				},
				false,
				0,
			),
		},
		{
			Outpoint: &externalapi.DomainOutpoint{
				TransactionID: externalapi.DomainTransactionID{},
				Index:         1,
			},
			UTXOEntry: utxo.NewUTXOEntry(
				2,
				&externalapi.ScriptPublicKey{
					Script:  nil,
					Version: 0,
				},
				true,
				3,
			),
		},
	}
}

func TestGetPruningPointUTXOs(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		// This is done to reduce the pruning depth to 8 blocks
		finalityDepth := 4
		params.FinalityDuration = time.Duration(finalityDepth) * params.TargetTimePerBlock
		params.K = 0

		params.BlockCoinbaseMaturity = 0

		factory := consensus.NewFactory()
		testConsensus, teardown, err := factory.NewTestConsensus(params, false, "TestGetPruningPointUTXOs")
		if err != nil {
			t.Fatalf("Error setting up testConsensus: %+v", err)
		}
		defer teardown(false)

		// Create a block that accepts the genesis coinbase so that we won't have script problems down the line
		emptyCoinbase := &externalapi.DomainCoinbaseData{
			ScriptPublicKey: &externalapi.ScriptPublicKey{
				Script:  nil,
				Version: 0,
			},
		}
		blockAboveGeneis, err := testConsensus.BuildBlock(emptyCoinbase, nil)
		if err != nil {
			t.Fatalf("Error building block above genesis: %+v", err)
		}
		_, err = testConsensus.ValidateAndInsertBlock(blockAboveGeneis)
		if err != nil {
			t.Fatalf("Error validating and inserting block above genesis: %+v", err)
		}

		// Create a block whose coinbase we could spend
		scriptPublicKey, redeemScript := testutils.OpTrueScript()
		coinbaseData := &externalapi.DomainCoinbaseData{ScriptPublicKey: scriptPublicKey}
		blockWithSpendableCoinbase, err := testConsensus.BuildBlock(coinbaseData, nil)
		if err != nil {
			t.Fatalf("Error building block with spendable coinbase: %+v", err)
		}
		_, err = testConsensus.ValidateAndInsertBlock(blockWithSpendableCoinbase)
		if err != nil {
			t.Fatalf("Error validating and inserting block with spendable coinbase: %+v", err)
		}

		// Create a transaction that adds a lot of UTXOs to the UTXO set
		transactionToSpend := blockWithSpendableCoinbase.Transactions[0]
		signatureScript, err := txscript.PayToScriptHashSignatureScript(redeemScript, nil)
		if err != nil {
			t.Fatalf("Error creating signature script: %+v", err)
		}
		input := &externalapi.DomainTransactionInput{
			PreviousOutpoint: externalapi.DomainOutpoint{
				TransactionID: *consensushashing.TransactionID(transactionToSpend),
				Index:         0,
			},
			SignatureScript: signatureScript,
			Sequence:        constants.MaxTxInSequenceNum,
		}

		outputs := make([]*externalapi.DomainTransactionOutput, 1125)
		for i := 0; i < len(outputs); i++ {
			outputs[i] = &externalapi.DomainTransactionOutput{
				ScriptPublicKey: scriptPublicKey,
				Value:           10000,
			}
		}
		spendingTransaction := &externalapi.DomainTransaction{
			Version: constants.MaxTransactionVersion,
			Inputs:  []*externalapi.DomainTransactionInput{input},
			Outputs: outputs,
			Payload: []byte{},
		}

		// Create a block with that includes the above transaction
		includingBlock, err := testConsensus.BuildBlock(emptyCoinbase, []*externalapi.DomainTransaction{spendingTransaction})
		if err != nil {
			t.Fatalf("Error building including block: %+v", err)
		}
		_, err = testConsensus.ValidateAndInsertBlock(includingBlock)
		if err != nil {
			t.Fatalf("Error validating and inserting including block: %+v", err)
		}

		// Add enough blocks to move the pruning point
		for {
			block, err := testConsensus.BuildBlock(emptyCoinbase, nil)
			if err != nil {
				t.Fatalf("Error building block: %+v", err)
			}
			_, err = testConsensus.ValidateAndInsertBlock(block)
			if err != nil {
				t.Fatalf("Error validating and inserting block: %+v", err)
			}

			pruningPoint, err := testConsensus.PruningPoint()
			if err != nil {
				t.Fatalf("Error getting the pruning point: %+v", err)
			}
			if !pruningPoint.Equal(params.GenesisHash) {
				break
			}
		}
		pruningPoint, err := testConsensus.PruningPoint()
		if err != nil {
			t.Fatalf("Error getting the pruning point: %+v", err)
		}

		// Get pruning point UTXOs in a loop
		var allOutpointAndUTXOEntryPairs []*externalapi.OutpointAndUTXOEntryPair
		step := 100
		var fromOutpoint *externalapi.DomainOutpoint
		for {
			outpointAndUTXOEntryPairs, err := testConsensus.GetPruningPointUTXOs(pruningPoint, fromOutpoint, step)
			if err != nil {
				t.Fatalf("Error getting pruning point UTXOs: %+v", err)
			}
			allOutpointAndUTXOEntryPairs = append(allOutpointAndUTXOEntryPairs, outpointAndUTXOEntryPairs...)
			fromOutpoint = outpointAndUTXOEntryPairs[len(outpointAndUTXOEntryPairs)-1].Outpoint

			if len(outpointAndUTXOEntryPairs) < step {
				break
			}
		}

		// Make sure the length of the UTXOs is exactly spendingTransaction.Outputs + 2 coinbase outputs
		if len(allOutpointAndUTXOEntryPairs) != len(outputs)+2 {
			t.Fatalf("Returned an unexpected amount of UTXOs. "+
				"Want: %d, got: %d", len(outputs)+2, len(allOutpointAndUTXOEntryPairs))
		}

		// Make sure all spendingTransaction.Outputs are in the returned UTXOs
		spendingTransactionID := consensushashing.TransactionID(spendingTransaction)
		for i := range outputs {
			found := false
			for _, outpointAndUTXOEntryPair := range allOutpointAndUTXOEntryPairs {
				outpoint := outpointAndUTXOEntryPair.Outpoint
				if outpoint.TransactionID == *spendingTransactionID && outpoint.Index == uint32(i) {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("Outpoint %s:%d not found amongst the returned UTXOs", spendingTransactionID, i)
			}
		}
	})
}

func BenchmarkGetPruningPointUTXOs(b *testing.B) {
	params := dagconfig.DevnetParams

	// This is done to reduce the pruning depth to 200 blocks
	finalityDepth := 100
	params.FinalityDuration = time.Duration(finalityDepth) * params.TargetTimePerBlock
	params.K = 0

	params.SkipProofOfWork = true
	params.BlockCoinbaseMaturity = 0

	factory := consensus.NewFactory()
	testConsensus, teardown, err := factory.NewTestConsensus(&params, false, "TestGetPruningPointUTXOs")
	if err != nil {
		b.Fatalf("Error setting up testConsensus: %+v", err)
	}
	defer teardown(false)

	// Create a block whose coinbase we could spend
	scriptPublicKey, redeemScript := testutils.OpTrueScript()
	coinbaseData := &externalapi.DomainCoinbaseData{ScriptPublicKey: scriptPublicKey}
	blockWithSpendableCoinbase, err := testConsensus.BuildBlock(coinbaseData, nil)
	if err != nil {
		b.Fatalf("Error building block with spendable coinbase: %+v", err)
	}
	_, err = testConsensus.ValidateAndInsertBlock(blockWithSpendableCoinbase)
	if err != nil {
		b.Fatalf("Error validating and inserting block with spendable coinbase: %+v", err)
	}

	addBlockWithLotsOfOutputs := func(b *testing.B, transactionToSpend *externalapi.DomainTransaction) *externalapi.DomainBlock {
		// Create a transaction that adds a lot of UTXOs to the UTXO set
		signatureScript, err := txscript.PayToScriptHashSignatureScript(redeemScript, nil)
		if err != nil {
			b.Fatalf("Error creating signature script: %+v", err)
		}
		input := &externalapi.DomainTransactionInput{
			PreviousOutpoint: externalapi.DomainOutpoint{
				TransactionID: *consensushashing.TransactionID(transactionToSpend),
				Index:         0,
			},
			SignatureScript: signatureScript,
			Sequence:        constants.MaxTxInSequenceNum,
		}
		outputs := make([]*externalapi.DomainTransactionOutput, 1125)
		for i := 0; i < len(outputs); i++ {
			outputs[i] = &externalapi.DomainTransactionOutput{
				ScriptPublicKey: scriptPublicKey,
				Value:           10000,
			}
		}
		transaction := &externalapi.DomainTransaction{
			Version: constants.MaxTransactionVersion,
			Inputs:  []*externalapi.DomainTransactionInput{input},
			Outputs: outputs,
			Payload: []byte{},
		}

		// Create a block that includes the above transaction
		block, err := testConsensus.BuildBlock(coinbaseData, []*externalapi.DomainTransaction{transaction})
		if err != nil {
			b.Fatalf("Error building block: %+v", err)
		}
		_, err = testConsensus.ValidateAndInsertBlock(block)
		if err != nil {
			b.Fatalf("Error validating and inserting block: %+v", err)
		}

		return block
	}

	// Add finalityDepth blocks, each containing lots of outputs
	tip := blockWithSpendableCoinbase
	for i := 0; i < finalityDepth; i++ {
		tip = addBlockWithLotsOfOutputs(b, tip.Transactions[0])
	}

	// Add enough blocks to move the pruning point
	for {
		block, err := testConsensus.BuildBlock(coinbaseData, nil)
		if err != nil {
			b.Fatalf("Error building block: %+v", err)
		}
		_, err = testConsensus.ValidateAndInsertBlock(block)
		if err != nil {
			b.Fatalf("Error validating and inserting block: %+v", err)
		}

		pruningPoint, err := testConsensus.PruningPoint()
		if err != nil {
			b.Fatalf("Error getting the pruning point: %+v", err)
		}
		if !pruningPoint.Equal(params.GenesisHash) {
			break
		}
	}
	pruningPoint, err := testConsensus.PruningPoint()
	if err != nil {
		b.Fatalf("Error getting the pruning point: %+v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Get pruning point UTXOs in a loop
		step := 100
		var fromOutpoint *externalapi.DomainOutpoint
		for {
			outpointAndUTXOEntryPairs, err := testConsensus.GetPruningPointUTXOs(pruningPoint, fromOutpoint, step)
			if err != nil {
				b.Fatalf("Error getting pruning point UTXOs: %+v", err)
			}
			fromOutpoint = outpointAndUTXOEntryPairs[len(outpointAndUTXOEntryPairs)-1].Outpoint

			if len(outpointAndUTXOEntryPairs) < step {
				break
			}
		}
	}
}
